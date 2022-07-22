// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	p4rtapi "github.com/onosproject/onos-api/go/onos/p4rt/v1"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/controller"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/wcmp-app/pkg/store/pipelineconfig"
	"github.com/onosproject/wcmp-app/pkg/store/topo"
	"time"
)

var log = logging.GetLogger()

const (
	defaultTimeout = 30 * time.Second
)

// NewController returns a new P4RT target  controller
func NewController(topo topo.Store, pipelineConfigs pipelineconfig.Store) *controller.Controller {
	c := controller.NewController("wcmp")
	c.Watch(&TopoWatcher{
		topo: topo,
	})

	c.Watch(&PipelineConfigWatcher{
		pipelineConfigs: pipelineConfigs,
	})

	c.Reconcile(&Reconciler{
		topo:            topo,
		pipelineConfigs: pipelineConfigs,
	})
	return c
}

// Reconciler reconciles P4RT connections
type Reconciler struct {
	topo            topo.Store
	pipelineConfigs pipelineconfig.Store
}

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
		log.Errorw("Failed Reconciling device pipeline config for target", "targetID", targetID, "error", err)
		return controller.Result{}, err
	}
	pipelinesInfo := p4rtServerInfo.Pipelines
	if len(pipelinesInfo) == 0 {
		log.Errorw("Failed Reconciling creating pipeline config for target", "targetID", targetID, "error", err)
		return controller.Result{}, err
	}
	pipelineInfo := p4rtServerInfo.Pipelines[0]
	pipelineName := pipelineInfo.Name
	pipelineVersion := pipelineInfo.Version
	pipelineArch := pipelineInfo.Architecture
	pipelineID := pipelineconfig.NewPipelineConfigID(p4rtapi.TargetID(targetID), pipelineName, pipelineVersion, pipelineArch)
	err = r.pipelineConfigs.Create(ctx, &p4rtapi.PipelineConfig{
		ID:       pipelineID,
		TargetID: p4rtapi.TargetID(targetID),
	})
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			log.Errorw("Failed Reconciling creating pipeline config for target", "targetID", targetID, "error", err)

			return controller.Result{}, err
		}
		return controller.Result{}, nil
	}
	return controller.Result{}, nil
}
