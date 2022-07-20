// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package charts

import (
	"github.com/onosproject/helmit/pkg/helm"
	"github.com/onosproject/onos-test/pkg/onostest"
)

// CreateUmbrellaRelease creates a helm release for an onos-umbrella instance
func CreateUmbrellaRelease() *helm.HelmRelease {
	return helm.Chart("onos-umbrella", onostest.OnosChartRepo).
		Release("onos-umbrella").
		Set("import.onos-gui.enabled", false).
		Set("import.onos-config.enabled", false).
		Set("onos-topo.image.tag", "latest")
}
