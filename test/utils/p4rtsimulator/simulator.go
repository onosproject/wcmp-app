// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package p4rtsimulator

import (
	"context"
	"testing"
	"time"
	"github.com/onosproject/onos-test/pkg/onostest"
	"github.com/onosproject/helmit/pkg/helm"
	"github.com/stretchr/testify/assert"
	"github.com/onosproject/helmit/pkg/util/random"
)

const (
	// Maximum time for an entire test to complete
	defaultTestTimeout = 3 * time.Minute
)


// MakeContext returns a new context for use in GNMI requests
func MakeContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	return context.WithTimeout(ctx, defaultTestTimeout)
}

// CreateSimulator creates a device simulator
func CreateSimulator(ctx context.Context, t *testing.T) *helm.HelmRelease {
	return CreateSimulatorWithName(ctx, t, random.NewPetName(2))
}

// CreateSimulatorWithName creates a device simulator
func CreateSimulatorWithName(ctx context.Context, t *testing.T, name string) *helm.HelmRelease {
	simulator := helm.
		Chart("stratum-simulator", onostest.OnosChartRepo).
		Release(name).
		Set("image.tag", "latest")
	err := simulator.Install(true)
	assert.NoError(t, err, "could not install device simulator %v", err)

	time.Sleep(2 * time.Second)

	return simulator
}

// DeleteSimulator shuts down the simulator pod and removes the target from topology
func DeleteSimulator(t *testing.T, simulator *helm.HelmRelease) {
	assert.NoError(t, simulator.Uninstall())
}

