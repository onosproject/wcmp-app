// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package pipeliner

import (
	"context"
	p4rtapi "github.com/onosproject/onos-api/go/onos/p4rt/v1"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/controller"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/wcmp-app/pkg/pluginregistry"
	"github.com/onosproject/wcmp-app/pkg/store/pipelineconfig"
	"github.com/onosproject/wcmp-app/pkg/store/topo"
	"google.golang.org/protobuf/proto"
	"time"
)

var log = logging.GetLogger()

const (
	defaultTimeout = 30 * time.Second
)

// NewController returns a new P4RT target  controller
func NewController(topo topo.Store, pipelineConfigs pipelineconfig.Store, p4PluginRegistry pluginregistry.P4PluginRegistry) *controller.Controller {
	c := controller.NewController("pipeliner")
	c.Watch(&TopoWatcher{
		topo: topo,
	})

	c.Watch(&PipelineConfigWatcher{
		pipelineConfigs: pipelineConfigs,
	})

	c.Reconcile(&Reconciler{
		topo:             topo,
		pipelineConfigs:  pipelineConfigs,
		p4PluginRegistry: p4PluginRegistry,
	})
	return c
}

// Reconciler reconciles P4RT connections
type Reconciler struct {
	topo             topo.Store
	pipelineConfigs  pipelineconfig.Store
	p4PluginRegistry pluginregistry.P4PluginRegistry
}

// Reconcile reconciles setting pipeline configuration
func (r *Reconciler) Reconcile(id controller.ID) (controller.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	targetID := id.Value.(topoapi.ID)
	target, err := r.topo.Get(ctx, targetID)
	if err != nil {
		if !errors.IsNotFound(err) {
			return controller.Result{}, err
		}
		return controller.Result{}, nil
	}
	p4rtServerInfo := &topoapi.P4RTServerInfo{}
	err = target.GetAspect(p4rtServerInfo)
	if err != nil {
		log.Errorw("Failed creating device pipeline config for target", "targetID", targetID, "error", err)
		return controller.Result{}, err
	}
	pipelinesInfo := p4rtServerInfo.Pipelines
	if len(pipelinesInfo) == 0 {
		log.Warnw("Failed creating device pipeline config for target", "targetID", targetID, "error", err)
		return controller.Result{}, err
	}
	pipelineInfo := p4rtServerInfo.Pipelines[0]
	pipelineName := pipelineInfo.Name
	pipelineVersion := pipelineInfo.Version
	pipelineArch := pipelineInfo.Architecture
	pipelineID := pipelineconfig.NewPipelineConfigID(p4rtapi.TargetID(targetID), pipelineName, pipelineVersion, pipelineArch)
	pluginID := p4rtapi.NewP4PluginID(pipelineName, pipelineVersion, pipelineArch)
	p4Plugin, err := r.p4PluginRegistry.GetPlugin(pluginID)
	if err != nil {
		log.Errorw("Failed creating device pipeline config for target", "pipelineConfigID", pipelineID, "targetID", targetID, "error", err)
		return controller.Result{}, err
	}

	deviceConfig, err := p4Plugin.GetP4DeviceConfig()
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Errorw("Failed Reconciling device pipeline config for target", "pipelineConfigID", pipelineID, "targetID", targetID, "error", err)
			return controller.Result{}, err
		}
		log.Warnw("Failed Reconciling device pipeline config for target; device config not found", "pipelineConfigID", pipelineID, "targetID", targetID, "error", err)
		return controller.Result{}, nil
	}
	// If device config is nil, we can initialize it with an empty byte array
	if deviceConfig == nil {
		deviceConfig = []byte{}
	}
	p4Info, err := p4Plugin.GetP4Info()
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Errorw("Failed creating device pipeline config for target", "pipelineConfigID", pipelineID, "targetID", targetID, "error", err)
			return controller.Result{}, err
		}
		return controller.Result{}, nil
	}
	p4InfoBytes, err := proto.Marshal(p4Info)
	if err != nil {
		log.Errorw("Failed creating device pipeline config for target", "pipelineConfigID", pipelineID, "targetID", targetID, "error", err)
		return controller.Result{}, err
	}

	err = r.pipelineConfigs.Create(ctx, &p4rtapi.PipelineConfig{
		ID:       pipelineID,
		TargetID: p4rtapi.TargetID(targetID),
		Action:   p4rtapi.ConfigurationAction_VERIFY_AND_COMMIT,
		Spec: &p4rtapi.PipelineConfigSpec{
			P4DeviceConfig: deviceConfig,
			P4Info:         p4InfoBytes,
		},
	})
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			log.Errorw("Failed Reconciling creating pipeline config for target", "targetID", targetID, "error", err)
			return controller.Result{}, err
		}
		return controller.Result{}, nil
	}
	log.Infow("Device Pipeline config is created successfully in pipeline config data store", "pipelineConfigID", pipelineID, "target ID", targetID)
	return controller.Result{}, nil
}
