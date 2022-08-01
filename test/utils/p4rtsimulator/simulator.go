// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package p4rtsimulator

import (
	"context"
	"github.com/onosproject/helmit/pkg/helm"
	"github.com/onosproject/helmit/pkg/kubernetes"
	"github.com/onosproject/helmit/pkg/util/random"
	"github.com/onosproject/onos-api/go/onos/topo"
	toposdk "github.com/onosproject/onos-ric-sdk-go/pkg/topo"
	"github.com/onosproject/onos-test/pkg/onostest"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

const (
	// Maximum time for an entire test to complete
	defaultTestTimeout = 3 * time.Minute

	// SimulatorTargetType type for simulated target
	SimulatorTargetType = "devicesim"
	// SimulatorTargetVersion default version for the simulated target
	SimulatorTargetVersion = "1.0.0"

	defaultTimeout = time.Second * 30
)

// NewSimulatorTargetEntity creates a topo entity for a device simulator target
func NewSimulatorTargetEntity(ctx context.Context, simulator *helm.HelmRelease, targetType string, targetVersion string) (*topo.Object, error) {
	simulatorClient := kubernetes.NewForReleaseOrDie(simulator)
	services, err := simulatorClient.CoreV1().Services().List(ctx)
	if err != nil {
		return nil, err
	}
	service := services[0]
	return NewTargetEntity(simulator.Name(), targetType, targetVersion, service.Ports()[0].Address(true))
}

// NewTargetEntity creates a topo entity with the specified target name, type, version and service address
func NewTargetEntity(name string, targetType string, targetVersion string, serviceAddress string) (*topo.Object, error) {
	o := topo.Object{
		ID:   topo.ID(name),
		Type: topo.Object_ENTITY,
		Obj: &topo.Object_Entity{
			Entity: &topo.Entity{
				KindID: topo.ID(targetType),
			},
		},
	}

	if err := o.SetAspect(&topo.TLSOptions{Insecure: true, Plain: true}); err != nil {
		return nil, err
	}

	return &o, nil
}

// NewSwitchEntity creates a switch entity with specified information
func NewSwitchEntity(name string, version string, serviceAddress string, servicePort string, deviceID uint64, modelID string, role string) (*topo.Object, error) {
	o := topo.Object{
		ID:   topo.ID(name),
		Type: topo.Object_ENTITY,
		Obj: &topo.Object_Entity{
			Entity: &topo.Entity{
				KindID: topo.ID(topo.SwitchKind),
			},
		},
	}

	if err := o.SetAspect(&topo.TLSOptions{Insecure: true, Plain: true}); err != nil {
		return nil, err
	}

	serverEndpoint := serviceAddress + ":" + servicePort

	timeout := defaultTimeout
	if err := o.SetAspect(&topo.Configurable{
		Type:                 topo.SwitchKind,
		Address:              serverEndpoint,
		Version:              version,
		Timeout:              &timeout,
		ValidateCapabilities: true,
	}); err != nil {
		return nil, err
	}

	port, _ := strconv.Atoi(servicePort)
	portNum := uint32(port)
	endpoint := topo.Endpoint{
		Address: serviceAddress,
		Port:    portNum,
	}

	if err := o.SetAspect(&topo.P4RTServerInfo{
		ControlEndpoint: &endpoint,
		DeviceID:        deviceID,
	}); err != nil {
		return nil, err
	}

	if err := o.SetAspect(&topo.Switch{
		ModelID: modelID,
		Role:    role,
	}); err != nil {
		return nil, err
	}

	return &o, nil
}

// AddTargetToTopo adds a new target to topo
func AddTargetToTopo(ctx context.Context, targetEntity *topo.Object) error {
	client, err := toposdk.NewClient()
	if err != nil {
		return err
	}
	err = client.Create(ctx, targetEntity)
	return err
}

// MakeContext returns a new context for use in GNMI requests
func MakeContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	return context.WithTimeout(ctx, defaultTestTimeout)
}

// CreateSimulator creates a device simulator
func CreateSimulator(ctx context.Context, t *testing.T) *helm.HelmRelease {
	return CreateSimulatorWithName(ctx, t, random.NewPetName(2), true)
}

// CreateSimulatorWithName creates a device simulator
func CreateSimulatorWithName(ctx context.Context, t *testing.T, name string, createTopoEntity bool) *helm.HelmRelease {
	simulator := helm.
		Chart("stratum-simulator", onostest.OnosChartRepo).
		Release(name).
		Set("image.tag", "latest")
	err := simulator.Install(true)
	assert.NoError(t, err, "could not install device simulator %v", err)

	time.Sleep(2 * time.Second)

	if createTopoEntity {
		simulatorTarget, err := NewSimulatorTargetEntity(ctx, simulator, SimulatorTargetType, SimulatorTargetVersion)
		assert.NoError(t, err, "could not make target for simulator %v", err)

		err = AddTargetToTopo(ctx, simulatorTarget)
		assert.NoError(t, err, "could not add target to topo for simulator %v", err)
	}

	return simulator
}

// DeleteSimulator shuts down the simulator pod and removes the target from topology
func DeleteSimulator(t *testing.T, simulator *helm.HelmRelease) {
	assert.NoError(t, simulator.Uninstall())
}
