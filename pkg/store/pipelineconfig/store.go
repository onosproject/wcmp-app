// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package pipelineconfig

import (
	"context"
	"fmt"
	"github.com/atomix/atomix-go-client/pkg/atomix"
	_map "github.com/atomix/atomix-go-client/pkg/atomix/map"
	"github.com/atomix/atomix-go-framework/pkg/atomix/meta"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	p4rtapi "github.com/onosproject/onos-api/go/onos/p4rt/v1"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"sync"
	"time"
)

var log = logging.GetLogger()

// NewPipelineConfigID creates a new pipeline config ID
func NewPipelineConfigID(targetID p4rtapi.TargetID, pipelineName string, pipelineVersion string, pipelineArch string) p4rtapi.PipelineConfigID {
	return p4rtapi.PipelineConfigID(fmt.Sprintf("%s-%s-%s-%s", targetID, pipelineName, pipelineVersion, pipelineArch))

}

// Store P4 pipeline pipelineconfig store interface
type Store interface {
	// Get gets the pipelineconfig intended for a given target ID
	Get(ctx context.Context, id p4rtapi.PipelineConfigID) (*p4rtapi.PipelineConfig, error)

	// Create creates a p4 pipeline pipelineconfig
	Create(ctx context.Context, pipelineConfig *p4rtapi.PipelineConfig) error

	// Update updates a p4 pipeline pipelineconfig
	Update(ctx context.Context, pipelineConfig *p4rtapi.PipelineConfig) error

	// List lists all the pipelineconfig
	List(ctx context.Context) ([]*p4rtapi.PipelineConfig, error)

	// Watch watches pipelineconfig changes
	Watch(ctx context.Context, ch chan<- p4rtapi.ConfigurationEvent, opts ...WatchOption) error

	// UpdateStatus updates a pipelineconfig status
	UpdateStatus(ctx context.Context, pipelineConfig *p4rtapi.PipelineConfig) error

	Close(ctx context.Context) error
}

// NewAtomixStore returns a new persistent Store
func NewAtomixStore(client atomix.Client) (Store, error) {
	pipelineConfigs, err := client.GetMap(context.Background(), "wcmp-app-pipeline-configurations")
	if err != nil {
		return nil, errors.FromAtomix(err)
	}
	store := &configurationStore{
		pipelineConfigs: pipelineConfigs,
		cache:           make(map[p4rtapi.PipelineConfigID]*_map.Entry),
		watchers:        make(map[uuid.UUID]chan<- p4rtapi.ConfigurationEvent),
		eventCh:         make(chan p4rtapi.ConfigurationEvent, 1000),
	}
	if err := store.open(context.Background()); err != nil {
		return nil, err
	}
	return store, nil
}

type watchOptions struct {
	configurationID p4rtapi.PipelineConfigID
	replay          bool
}

// WatchOption is a pipelineconfig option for Watch calls
type WatchOption interface {
	apply(*watchOptions)
}

// watchReplyOption is an option to replay events on watch
type watchReplayOption struct {
}

func (o watchReplayOption) apply(options *watchOptions) {
	options.replay = true
}

// WithReplay returns a WatchOption that replays past changes
func WithReplay() WatchOption {
	return watchReplayOption{}
}

type watchIDOption struct {
	id p4rtapi.PipelineConfigID
}

func (o watchIDOption) apply(options *watchOptions) {
	options.configurationID = o.id
}

// WithPipelineConfigID returns a Watch option that watches for configurations based on a given pipeline pipelineconfig ID
func WithPipelineConfigID(id p4rtapi.PipelineConfigID) WatchOption {
	return watchIDOption{id: id}
}

type configurationStore struct {
	pipelineConfigs _map.Map
	cache           map[p4rtapi.PipelineConfigID]*_map.Entry
	cacheMu         sync.RWMutex
	watchers        map[uuid.UUID]chan<- p4rtapi.ConfigurationEvent
	watchersMu      sync.RWMutex
	eventCh         chan p4rtapi.ConfigurationEvent
}

func (s *configurationStore) open(ctx context.Context) error {
	ch := make(chan _map.Event)
	if err := s.pipelineConfigs.Watch(ctx, ch, _map.WithReplay()); err != nil {
		return err
	}
	go func() {
		for event := range ch {
			entry := event.Entry
			s.updateCache(&entry)
		}
	}()
	go s.processEvents()
	return nil
}
func (s *configurationStore) processEvents() {
	for event := range s.eventCh {
		s.watchersMu.RLock()
		for _, watcher := range s.watchers {
			watcher <- event
		}
		s.watchersMu.RUnlock()
	}
}

func (s *configurationStore) updateCache(newEntry *_map.Entry) {
	configurationID := p4rtapi.PipelineConfigID(newEntry.Key)

	// Use a double-checked lock when updating the cache.
	// First, check for a more recent version of the pipelineconfig already in the cache.
	s.cacheMu.RLock()
	entry, ok := s.cache[configurationID]
	s.cacheMu.RUnlock()
	if ok && entry.Revision >= newEntry.Revision {
		return
	}

	// The cache needs to be updated. Acquire a write lock and check once again
	// for a more recent version of the pipelineconfig.
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	entry, ok = s.cache[configurationID]
	if !ok {
		s.cache[configurationID] = newEntry
		var pipelineConfig p4rtapi.PipelineConfig
		if err := decodePipelineConfiguration(newEntry, &pipelineConfig); err != nil {
			log.Error(err)
		} else {
			s.eventCh <- p4rtapi.ConfigurationEvent{
				Type:           p4rtapi.ConfigurationEvent_CREATED,
				PipelineConfig: pipelineConfig,
			}
		}
	} else if newEntry.Revision > entry.Revision {
		// Add the pipelineconfig to the ID and index caches and publish an event.
		s.cache[configurationID] = newEntry
		var pipelineConfig p4rtapi.PipelineConfig
		if err := decodePipelineConfiguration(newEntry, &pipelineConfig); err != nil {
			log.Error(err)
		} else {
			s.eventCh <- p4rtapi.ConfigurationEvent{
				Type:           p4rtapi.ConfigurationEvent_UPDATED,
				PipelineConfig: pipelineConfig,
			}
		}
	}
}

func (s *configurationStore) Get(ctx context.Context, id p4rtapi.PipelineConfigID) (*p4rtapi.PipelineConfig, error) {
	// Check the ID cache for the latest version of the pipelineconfig.
	s.cacheMu.RLock()
	cachedEntry, ok := s.cache[id]
	s.cacheMu.RUnlock()
	if ok {
		pipelineConfig := &p4rtapi.PipelineConfig{}
		if err := decodePipelineConfiguration(cachedEntry, pipelineConfig); err != nil {
			return nil, errors.NewInvalid("pipelineconfig decoding failed: %v", err)
		}
		return pipelineConfig, nil
	}

	// If the pipelineconfig is not already in the cache, get it from the underlying primitive.
	entry, err := s.pipelineConfigs.Get(ctx, string(id))
	if err != nil {
		return nil, errors.FromAtomix(err)
	}

	// Decode and return the Configuration.
	configuration := &p4rtapi.PipelineConfig{}
	if err := decodePipelineConfiguration(entry, configuration); err != nil {
		return nil, errors.NewInvalid("pipelineconfig decoding failed: %v", err)
	}

	// Update the cache.
	s.updateCache(entry)
	return configuration, nil
}

func (s *configurationStore) Create(ctx context.Context, pipelineConfig *p4rtapi.PipelineConfig) error {
	if pipelineConfig.ID == "" {
		return errors.NewInvalid("no pipeline pipelineconfig ID specified")
	}
	if pipelineConfig.TargetID == "" {
		return errors.NewInvalid("no target ID specified")
	}
	if pipelineConfig.Revision != 0 {
		return errors.NewInvalid("cannot create pipeline pipelineconfig with revision")
	}
	if pipelineConfig.Version != 0 {
		return errors.NewInvalid("cannot create pipeline pipelineconfig with version")
	}
	pipelineConfig.Revision = 1
	pipelineConfig.Created = time.Now()
	pipelineConfig.Updated = time.Now()

	// Encode the pipelineconfig bytes.
	bytes, err := proto.Marshal(pipelineConfig)
	if err != nil {
		return errors.NewInvalid("pipeline pipelineconfig encoding failed: %v", err)
	}

	// Create the entry in the underlying map primitive.
	entry, err := s.pipelineConfigs.Put(ctx, string(pipelineConfig.ID), bytes, _map.IfNotSet())
	if err != nil {
		return errors.FromAtomix(err)
	}

	// Decode the pipleline pipelineconfig from the returned entry bytes.
	if err := decodePipelineConfiguration(entry, pipelineConfig); err != nil {
		return errors.NewInvalid("pipelineconfig decoding failed: %v", err)
	}

	// Update the cache.
	s.updateCache(entry)
	return nil
}

func (s *configurationStore) Update(ctx context.Context, pipelineConfig *p4rtapi.PipelineConfig) error {
	if pipelineConfig.ID == "" {
		return errors.NewInvalid("no pipelineconfig ID specified")
	}
	if pipelineConfig.TargetID == "" {
		return errors.NewInvalid("no target ID specified")
	}
	if pipelineConfig.Revision == 0 {
		return errors.NewInvalid("pipelineconfig must contain a revision on update")
	}
	if pipelineConfig.Version == 0 {
		return errors.NewInvalid("pipelineconfig must contain a version on update")
	}
	pipelineConfig.Revision++
	pipelineConfig.Updated = time.Now()

	// Encode the pipelineconfig bytes.
	bytes, err := proto.Marshal(pipelineConfig)
	if err != nil {
		return errors.NewInvalid("pipleline pipelineconfig encoding failed: %v", err)
	}

	// Update the entry in the underlying map primitive using the pipelineconfig version
	// as an optimistic lock.
	entry, err := s.pipelineConfigs.Put(ctx, string(pipelineConfig.ID), bytes, _map.IfMatch(meta.NewRevision(meta.Revision(pipelineConfig.Version))))
	if err != nil {
		return errors.FromAtomix(err)
	}

	// Decode the pipelineconfig from the returned entry bytes.
	if err := decodePipelineConfiguration(entry, pipelineConfig); err != nil {
		return errors.NewInvalid("pipelineconfig decoding failed: %v", err)
	}

	// Update the cache.
	s.updateCache(entry)
	return nil
}

func (s *configurationStore) UpdateStatus(ctx context.Context, pipelineConfig *p4rtapi.PipelineConfig) error {
	if pipelineConfig.ID == "" {
		return errors.NewInvalid("no pipeline pipelineconfig ID specified")
	}
	if pipelineConfig.TargetID == "" {
		return errors.NewInvalid("no target ID specified")
	}
	if pipelineConfig.Revision == 0 {
		return errors.NewInvalid("pipeline pipelineconfig must contain a revision on update")
	}
	if pipelineConfig.Version == 0 {
		return errors.NewInvalid("pipeline pipelineconfig must contain a version on update")
	}
	pipelineConfig.Updated = time.Now()

	// Encode the pipelineconfig bytes.
	bytes, err := proto.Marshal(pipelineConfig)
	if err != nil {
		return errors.NewInvalid("pipeline pipelineconfig encoding failed: %v", err)
	}

	// Update the entry in the underlying map primitive using the pipelineconfig version
	// as an optimistic lock.
	entry, err := s.pipelineConfigs.Put(ctx, string(pipelineConfig.ID), bytes, _map.IfMatch(meta.NewRevision(meta.Revision(pipelineConfig.Version))))
	if err != nil {
		return errors.FromAtomix(err)
	}

	// Decode the pipeline pipelineconfig from the returned entry bytes.
	if err := decodePipelineConfiguration(entry, pipelineConfig); err != nil {
		return errors.NewInvalid("pipeline pipelineconfig decoding failed: %v", err)
	}

	// Update the cache.
	s.updateCache(entry)
	return nil
}

func (s *configurationStore) List(ctx context.Context) ([]*p4rtapi.PipelineConfig, error) {
	mapCh := make(chan _map.Entry)
	if err := s.pipelineConfigs.Entries(ctx, mapCh); err != nil {
		return nil, errors.FromAtomix(err)
	}

	pipelineConfigs := make([]*p4rtapi.PipelineConfig, 0)

	for entry := range mapCh {
		pipelineConfig := &p4rtapi.PipelineConfig{}
		if err := decodePipelineConfiguration(&entry, pipelineConfig); err != nil {
			log.Error(err)
		} else {
			pipelineConfigs = append(pipelineConfigs, pipelineConfig)
		}
	}
	return pipelineConfigs, nil
}

func (s *configurationStore) Watch(ctx context.Context, ch chan<- p4rtapi.ConfigurationEvent, opts ...WatchOption) error {
	var options watchOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	watchCh := make(chan p4rtapi.ConfigurationEvent, 10)
	id := uuid.New()
	s.watchersMu.Lock()
	s.watchers[id] = watchCh
	s.watchersMu.Unlock()

	var replay []p4rtapi.ConfigurationEvent
	if options.replay {
		if options.configurationID == "" {
			s.cacheMu.RLock()
			replay = make([]p4rtapi.ConfigurationEvent, 0, len(s.cache))
			for _, entry := range s.cache {
				var pipelineConfig p4rtapi.PipelineConfig
				if err := decodePipelineConfiguration(entry, &pipelineConfig); err != nil {
					log.Errorw("error", err)
				} else {
					replay = append(replay, p4rtapi.ConfigurationEvent{
						Type:           p4rtapi.ConfigurationEvent_REPLAYED,
						PipelineConfig: pipelineConfig,
					})
				}
			}
			s.cacheMu.RUnlock()
		} else {
			s.cacheMu.RLock()
			entry, ok := s.cache[options.configurationID]
			if ok {
				var pipelineConfig p4rtapi.PipelineConfig
				if err := decodePipelineConfiguration(entry, &pipelineConfig); err != nil {
					log.Error(err)
				} else {
					replay = []p4rtapi.ConfigurationEvent{
						{
							Type:           p4rtapi.ConfigurationEvent_REPLAYED,
							PipelineConfig: pipelineConfig,
						},
					}
				}
			}
			s.cacheMu.RUnlock()
		}
	}

	go func() {
		defer close(ch)
		for _, event := range replay {
			ch <- event
		}
		for event := range watchCh {
			if options.configurationID == "" || event.PipelineConfig.ID == options.configurationID {
				ch <- event
			}
		}
	}()

	go func() {
		<-ctx.Done()
		s.watchersMu.Lock()
		delete(s.watchers, id)
		s.watchersMu.Unlock()
		close(watchCh)
	}()
	return nil
}

func (s *configurationStore) Close(ctx context.Context) error {
	err := s.pipelineConfigs.Close(ctx)
	if err != nil {
		return errors.FromAtomix(err)
	}
	return nil
}

func decodePipelineConfiguration(entry *_map.Entry, pipelineConfig *p4rtapi.PipelineConfig) error {
	if err := proto.Unmarshal(entry.Value, pipelineConfig); err != nil {
		return err
	}
	pipelineConfig.ID = p4rtapi.PipelineConfigID(entry.Key)
	pipelineConfig.Key = entry.Key
	pipelineConfig.Version = uint64(entry.Revision)
	return nil
}
