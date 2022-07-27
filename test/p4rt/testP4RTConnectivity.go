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

// TestP4RTConnectivity will test the connectivity
func (s *TestSuite) TestP4RTConnectivity(t *testing.T) {
	ctx, cancel := p4rtsimulator.MakeContext()
	defer cancel()

	// Create a simulated device
	targetID := "test-topo-integration-target-1"
	simulator := p4rtsimulator.CreateSimulatorWithName(ctx, t, "stratum-simulator")
	assert.NotNil(t, simulator)
	topoutils.WaitForTargetAvailable(ctx, t, topoapi.ID(targetID), 2*time.Minute)
	defer p4rtsimulator.DeleteSimulator(t, simulator)

	// Get a topology API client
	client, err := topoutils.NewClientTopo()
	assert.NoError(t, err)
	assert.NotNil(t, client)

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
