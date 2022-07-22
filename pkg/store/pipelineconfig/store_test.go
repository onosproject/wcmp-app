// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package pipelineconfig

import (
	"context"
	"testing"
	"time"

	"github.com/atomix/atomix-go-client/pkg/atomix/test"
	"github.com/atomix/atomix-go-client/pkg/atomix/test/rsm"
	p4rtapi "github.com/onosproject/onos-api/go/onos/p4rt/v1"
	"github.com/stretchr/testify/assert"
)

func TestConfigurationStore(t *testing.T) {
	test := test.NewTest(
		rsm.NewProtocol(),
		test.WithReplicas(1),
		test.WithPartitions(1),
	)
	assert.NoError(t, test.Start())
	defer test.Stop()

	client1, err := test.NewClient("node-1")
	assert.NoError(t, err)

	client2, err := test.NewClient("node-2")
	assert.NoError(t, err)

	store1, err := NewAtomixStore(client1)
	assert.NoError(t, err)

	store2, err := NewAtomixStore(client2)
	assert.NoError(t, err)

	target1 := p4rtapi.TargetID("target-1")
	target2 := p4rtapi.TargetID("target-2")

	ch := make(chan p4rtapi.ConfigurationEvent)
	err = store2.Watch(context.Background(), ch)
	assert.NoError(t, err)

	var target1DeviceConfig []byte

	targetConfigID1 := NewPipelineConfigID(target1, "basic", "1.0.0", "v1model")
	target1Config := &p4rtapi.PipelineConfig{
		ID:       targetConfigID1,
		TargetID: target1,
		Spec: &p4rtapi.PipelineConfigSpec{
			P4DeviceConfig: target1DeviceConfig,
		},
	}

	var target2DeviceConfig []byte

	targetConfigID2 := NewPipelineConfigID(target2, "basic", "1.0.0", "v1model")
	target2Config := &p4rtapi.PipelineConfig{
		ID:       targetConfigID2,
		TargetID: target2,
		Spec: &p4rtapi.PipelineConfigSpec{
			P4DeviceConfig: target2DeviceConfig,
		},
	}

	err = store1.Create(context.TODO(), target1Config)
	assert.NoError(t, err)
	assert.Equal(t, targetConfigID1, target1Config.ID)
	assert.NotEqual(t, p4rtapi.Revision(0), target1Config.Revision)

	err = store2.Create(context.TODO(), target2Config)
	assert.NoError(t, err)
	assert.Equal(t, targetConfigID2, target2Config.ID)
	assert.NotEqual(t, p4rtapi.Revision(0), target2Config.Revision)

	// Get the pipelineconfig
	target1Config, err = store2.Get(context.TODO(), targetConfigID1)
	assert.NoError(t, err)
	assert.NotNil(t, target1Config)
	assert.Equal(t, targetConfigID1, target1Config.ID)
	assert.NotEqual(t, p4rtapi.Revision(0), target1Config.Revision)

	// Verify events were received for the pipeline configs
	configurationEvent := nextEvent(t, ch)
	assert.NotNil(t, configurationEvent)
	configurationEvent = nextEvent(t, ch)
	assert.NotNil(t, configurationEvent)

	// Watch events for a specific pipeline pipelineconfig
	configurationCh := make(chan p4rtapi.ConfigurationEvent)
	err = store1.Watch(context.TODO(), configurationCh, WithPipelineConfigID(target2Config.ID))
	assert.NoError(t, err)

	// Update one of the pipeline configs
	revision := target2Config.Revision
	err = store1.Update(context.TODO(), target2Config)
	assert.NoError(t, err)
	assert.NotEqual(t, revision, target2Config.Revision)

	event := <-configurationCh
	assert.Equal(t, target2Config.ID, event.PipelineConfig.ID)

	// Lists configurations
	configurationList, err := store1.List(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(configurationList))

	// Read and then update the pipelineconfig
	target2Config, err = store2.Get(context.TODO(), targetConfigID2)
	assert.NoError(t, err)
	assert.NotNil(t, target2Config)
	target2Config.Status.State = p4rtapi.PipelineConfigState_PIPELINE_CONFIG_PENDING
	revision = target2Config.Revision
	err = store1.Update(context.TODO(), target2Config)
	assert.NoError(t, err)
	assert.NotEqual(t, revision, target2Config.Revision)

	event = <-configurationCh
	assert.Equal(t, target2Config.ID, event.PipelineConfig.ID)

	// Verify that concurrent updates fail
	target1Config11, err := store1.Get(context.TODO(), targetConfigID1)
	assert.NoError(t, err)
	target1Config12, err := store2.Get(context.TODO(), targetConfigID1)
	assert.NoError(t, err)

	target1Config11.Status.State = p4rtapi.PipelineConfigState_PIPELINE_CONFIG_PENDING
	err = store1.Update(context.TODO(), target1Config11)
	assert.NoError(t, err)

	target1Config12.Status.State = p4rtapi.PipelineConfigState_PIPELINE_CONFIG_FAILED
	err = store2.Update(context.TODO(), target1Config12)
	assert.Error(t, err)

	// Verify events were received again
	configurationEvent = nextEvent(t, ch)
	assert.NotNil(t, configurationEvent)
	configurationEvent = nextEvent(t, ch)
	assert.NotNil(t, configurationEvent)
	configurationEvent = nextEvent(t, ch)
	assert.NotNil(t, configurationEvent)

	// Checks list of pipelineconfig after deleting a pipelineconfig
	configurationList, err = store2.List(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(configurationList))

	err = store1.Close(context.TODO())
	assert.NoError(t, err)

	err = store2.Close(context.TODO())
	assert.NoError(t, err)

}

func nextEvent(t *testing.T, ch chan p4rtapi.ConfigurationEvent) *p4rtapi.PipelineConfig {
	select {
	case c := <-ch:
		return &c.PipelineConfig
	case <-time.After(5 * time.Second):
		t.FailNow()
	}
	return nil
}
