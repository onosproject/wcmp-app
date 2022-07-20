// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"
	"github.com/onosproject/wcmp-app/test/utils/p4rtsimulator"
)

func (s *TestSuite) testSinglePath(t *testing.T) {
	ctx, cancel := p4rtsimulator.MakeContext()
	defer cancel()

	// Create a simulated device
	simulator := p4rtsimulator.CreateSimulator(ctx, t)
	defer p4rtsimulator.DeleteSimulator(t, simulator)
}
