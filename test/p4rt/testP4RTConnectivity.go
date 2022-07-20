// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"
	"time"
	"github.com/onosproject/onos-config/test/utils/p4rtsimulator"
)

func (s *TestSuite) testSinglePath(t *testing.T) {
		// Create a simulated device
		simulator := p4rtsimulator.CreateSimulator(ctx, t)
		defer p4rtsimulator.DeleteSimulator(t, simulator)
}
