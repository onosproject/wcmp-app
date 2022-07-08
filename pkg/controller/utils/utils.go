// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/env"
	"github.com/onosproject/onos-lib-go/pkg/uri"
)

// GetControllerID gets controller URI
func GetControllerID() topoapi.ID {
	return topoapi.ID(uri.NewURI(
		uri.WithScheme("p4rt"),
		uri.WithOpaque(env.GetPodID())).String())
}
