// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/wcmp-app/test/utils/p4rtsimulator"
	topoutils "github.com/onosproject/wcmp-app/test/utils/topo"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	name           = "p4rt:4"
	version        = "1.0.x"
	serviceAddress = "stratum-simulator"
	servicePort    = "50001"
	deviceID       = 1
	modelID        = "test"
	role           = "leaf"
)

// TestP4RTConnectivity will test the connectivity
func (s *TestSuite) TestP4RTConnectivity(t *testing.T) {
	ctx, cancel := p4rtsimulator.MakeContext()
	defer cancel()

	// Create a simulated device
	targetID := "test-topo-integration-target-1"
	simulator := p4rtsimulator.CreateSimulatorWithName(ctx, t, "stratum-simulator", true)
	assert.NotNil(t, simulator)
	defer p4rtsimulator.DeleteSimulator(t, simulator)

	// Get a topology API client
	client, err := topoutils.NewClientTopo()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	newSwitch, err := p4rtsimulator.NewSwitchEntity(name, version, serviceAddress, servicePort, deviceID, modelID, role)
	assert.NoError(t, err)
	err = client.Create(newSwitch)
	assert.NoError(t, err)

	client.WaitForTargetAvailable(ctx, t, topoapi.ID(targetID), 2*time.Minute)

	// Check the number of control relations
	relationsList, err := client.GetControlRelations()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(relationsList))

	// Check the number of switches
	switchList, err := client.GetSwitchEntities()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(switchList))

	// Check the number of controllers
	controllerList, err := client.GetControllerEntities()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(controllerList))
}
