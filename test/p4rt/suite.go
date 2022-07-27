// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/onosproject/helmit/pkg/input"
	"github.com/onosproject/helmit/pkg/test"
	"github.com/onosproject/wcmp-app/test/utils/charts"
)

type testSuite struct {
	test.Suite
}

// TestSuite is the wcmp-app P4RT test suite
type TestSuite struct {
	testSuite
}

// SetupTestSuite sets up the wcmp-app P4RT test suite
func (s *TestSuite) SetupTestSuite(c *input.Context) error {
	registry := c.GetArg("registry").String("")
	umbrella := charts.CreateUmbrellaRelease()
	r := umbrella.
		Set("global.image.registry", registry).
		Set("import.onos-cli.enabled", true). // not needed - can be enabled by adding '--set onos-umbrella.import.onos-cli.enabled=true' to helmit args for investigation
		Set("import.wcmp-app.enabled", true).
		Install(true)
	return r
}
