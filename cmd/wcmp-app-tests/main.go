// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
		"github.com/onosproject/helmit/pkg/registry"
		"github.com/onosproject/helmit/pkg/test"
		"github.com/onosproject/wcmp-app/test/p4rt"
		_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	registry.RegisterTestSuite("p4rt", &p4rt.TestSuite{})
	test.Main()
}
