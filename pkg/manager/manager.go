// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-lib-go/pkg/northbound"
	"github.com/onosproject/wcmp-app/pkg/controller/connection"
	"github.com/onosproject/wcmp-app/pkg/controller/mastership"
	"github.com/onosproject/wcmp-app/pkg/controller/node"
	"github.com/onosproject/wcmp-app/pkg/controller/target"
	"github.com/onosproject/wcmp-app/pkg/pluginregistry"
	"github.com/onosproject/wcmp-app/pkg/southbound/p4rt"
	"github.com/onosproject/wcmp-app/pkg/store/topo"
)

var log = logging.GetLogger()

// Config is a manager configuration
type Config struct {
	CAPath      string
	KeyPath     string
	CertPath    string
	TopoAddress string
	GRPCPort    int
	P4Plugins   []string
}

// Manager single point of entry for the wcmp-app
type Manager struct {
	Config Config
}

// NewManager initializes the application manager
func NewManager(cfg Config) *Manager {
	log.Infow("Creating manager")
	p4PluginRegistry := pluginregistry.NewP4PluginRegistry()
	for _, smp := range cfg.P4Plugins {
		if err := p4PluginRegistry.RegisterPlugin(smp); err != nil {
			log.Fatal(err)
		}
	}
	mgr := Manager{
		Config: cfg,
	}
	return &mgr
}

// Run runs manager
func (m *Manager) Run() {
	log.Infow("Starting Manager")

	if err := m.Start(); err != nil {
		log.Fatalw("Unable to run Manager", "error", err)
	}
}

// Start initializes and starts controllers, stores, southbound modules.
func (m *Manager) Start() error {
	opts, err := certs.HandleCertPaths(m.Config.CAPath, m.Config.KeyPath, m.Config.CertPath, true)
	if err != nil {
		return err
	}

	// Create new topo store
	topoStore, err := topo.NewStore(m.Config.TopoAddress, opts...)
	if err != nil {
		return err
	}
	conns := p4rt.NewConnManager()
	// Starts NB server
	err = m.startNorthboundServer()
	if err != nil {
		return err
	}
	// Starts node controller
	err = m.startNodeController(topoStore)
	if err != nil {
		return err
	}
	// Starts connection controller
	err = m.startConnController(topoStore, conns)
	if err != nil {
		return err
	}

	// Starts target controller
	err = m.startTargetController(topoStore, conns)
	if err != nil {
		return err
	}
	// Starts mastership controller
	err = m.startMastershipController(topoStore, conns)
	if err != nil {
		return err
	}

	return nil
}

// startNodeController starts node controller
func (m *Manager) startNodeController(topo topo.Store) error {
	nodeController := node.NewController(topo)
	return nodeController.Start()
}

// startConnController starts connection controller
func (m *Manager) startConnController(topo topo.Store, conns p4rt.ConnManager) error {
	connController := connection.NewController(topo, conns)
	return connController.Start()
}

// startTargetController starts target controller
func (m *Manager) startTargetController(topo topo.Store, conns p4rt.ConnManager) error {
	targetController := target.NewController(topo, conns)
	return targetController.Start()
}

// startMastershipController starts mastership controller
func (m *Manager) startMastershipController(topo topo.Store, conns p4rt.ConnManager) error {
	mastershipController := mastership.NewController(topo, conns)
	return mastershipController.Start()
}

// startSouthboundServer starts the northbound gRPC server
func (m *Manager) startNorthboundServer() error {
	s := northbound.NewServer(northbound.NewServerCfg(
		m.Config.CAPath,
		m.Config.KeyPath,
		m.Config.CertPath,
		int16(m.Config.GRPCPort),
		true,
		northbound.SecurityConfig{}))
	s.AddService(logging.Service{})

	doneCh := make(chan error)
	go func() {
		err := s.Serve(func(started string) {
			log.Info("Started NBI on ", started)
			close(doneCh)
		})
		if err != nil {
			doneCh <- err
		}
	}()
	return <-doneCh
}

// Close kills the manager
func (m *Manager) Close() {
	log.Infow("Closing Manager")
}
