// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package topo

import (
	"context"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	toposdk "github.com/onosproject/onos-ric-sdk-go/pkg/topo"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// ClientTopo interface
type ClientTopo interface {
	GetControlRelations() ([]topoapi.Object, error)
	GetSwitchEntities() ([]topoapi.Object, error)
	GetControllerEntities() ([]topoapi.Object, error)
	WaitForTargetAvailable(ctx context.Context, t *testing.T, objectID topoapi.ID, timeout time.Duration) bool
}

// NewClientTopo creates a new topo SDK client
func NewClientTopo() (Client, error) {
	sdkClient, err := toposdk.NewClient()
	if err != nil {
		return Client{}, err
	}
	cl := Client{
		client: sdkClient,
	}
	return cl, nil
}

// Client topo SDK client
type Client struct {
	client toposdk.Client
}

func getFilter(kind string) *topoapi.Filters {
	controlRelationFilter := &topoapi.Filters{
		KindFilter: &topoapi.Filter{
			Filter: &topoapi.Filter_Equal_{
				Equal_: &topoapi.EqualFilter{
					Value: kind,
				},
			},
		},
	}
	return controlRelationFilter
}

// GetControlRelationFilter gets control relation filter
func GetControlRelationFilter() *topoapi.Filters {
	return getFilter(topoapi.ControlsKind)
}

// GetControlRelations returns a list of the control relations
func (c *Client) GetControlRelations() ([]topoapi.Object, error) {
	filter := GetControlRelationFilter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	relationsList, err := c.client.List(ctx, toposdk.WithListFilters(filter))
	cancel()
	return relationsList, err
}

// GetSwitchFilter gets switch filter
func GetSwitchFilter() *topoapi.Filters {
	return getFilter(topoapi.SwitchKind)
}

// GetSwitchEntities returns a list of the switches
func (c *Client) GetSwitchEntities() ([]topoapi.Object, error) {
	filter := GetSwitchFilter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	switchList, err := c.client.List(ctx, toposdk.WithListFilters(filter))
	cancel()
	return switchList, err
}

// GetControllerFilter gets controller filter
func GetControllerFilter() *topoapi.Filters {
	return getFilter(topoapi.ControllerKind)
}

// GetControllerEntities returns a list of the controllers
func (c *Client) GetControllerEntities() ([]topoapi.Object, error) {
	filter := GetControllerFilter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	controllerList, err := c.client.List(ctx, toposdk.WithListFilters(filter))
	cancel()
	return controllerList, err
}

// WaitForControlRelation waits to create control relation for a given target
func WaitForControlRelation(ctx context.Context, t *testing.T, predicate func(*topoapi.Relation, topoapi.Event) bool, timeout time.Duration) bool {
	cl, err := NewClientTopo()
	assert.NoError(t, err)
	stream := make(chan topoapi.Event)
	err = cl.client.Watch(ctx, stream, toposdk.WithWatchFilters(GetControlRelationFilter()))
	assert.NoError(t, err)
	for event := range stream {
		if predicate(event.Object.GetRelation(), event) {
			return true
		} // Otherwise, loop and wait for the next topo event
	}

	return false
}

// WaitForTargetAvailable waits for a target to become available
func WaitForTargetAvailable(ctx context.Context, t *testing.T, objectID topoapi.ID, timeout time.Duration) bool {
	return WaitForControlRelation(ctx, t, func(rel *topoapi.Relation, event topoapi.Event) bool {
		if rel.TgtEntityID != objectID {
			t.Logf("Topo %v event from %s (expected %s). Discarding\n", event.Type, rel.TgtEntityID, objectID)
			return false
		}

		if event.Type == topoapi.EventType_ADDED || event.Type == topoapi.EventType_UPDATED || event.Type == topoapi.EventType_NONE {
			cl, err := NewClientTopo()
			assert.NoError(t, err)
			_, err = cl.client.Get(ctx, event.Object.ID)
			if err == nil {
				t.Logf("Target %s is available", objectID)
				return true
			}
		}

		return false
	}, timeout)
}
