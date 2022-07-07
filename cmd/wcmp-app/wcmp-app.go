// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/wcmp-app/pkg/manager"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

var log = logging.GetLogger()

// The main entry point
func main() {
	if err := getRootCommand().Execute(); err != nil {
		println(err)
		os.Exit(1)
	}
}

func getRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wcmp-app",
		Short: "wcmp-app",
		RunE:  runRootCommand,
	}
	cmd.Flags().String("caPath", "", "path to CA certificate")
	cmd.Flags().String("keyPath", "", "path to client private key")
	cmd.Flags().String("certPath", "", "path to client certificate")
	cmd.Flags().String("topoEndpoint", "onos-topo:5150", "topology service endpoint")
	return cmd
}

func runRootCommand(cmd *cobra.Command, args []string) error {
	caPath, _ := cmd.Flags().GetString("caPath")
	keyPath, _ := cmd.Flags().GetString("keyPath")
	certPath, _ := cmd.Flags().GetString("certPath")
	topoEndpoint, _ := cmd.Flags().GetString("topoEndpoint")

	log.Infow("Starting wcmp-app",
		"CAPath", caPath,
		"KeyPath", keyPath,
		"CertPath", certPath,
		"TopoAddress", topoEndpoint,
	)

	cfg := manager.Config{
		CAPath:      caPath,
		KeyPath:     keyPath,
		CertPath:    certPath,
		TopoAddress: topoEndpoint,
		GRPCPort:    5150,
	}

	mgr := manager.NewManager(cfg)

	mgr.Run()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	mgr.Close()
	return nil
}
