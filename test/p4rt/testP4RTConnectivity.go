// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"
	"time"
	"github.com/onosproject/wcmp-app/test/utils/p4rtsimulator"
)

func (s *TestSuite) testSinglePath(t *testing.T) {
	ctx, cancel := p4rtsimulator.MakeContext()
	defer cancel()

	// Create a simulated device
	simulator := p4rtsimulator.CreateSimulatorWithName(ctx, t, "stratum-simulator")
	defer p4rtsimulator.DeleteSimulator(t, simulator)

	time.Sleep(2 * time.Second)
}
