// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package pipelineconfig

import (
	"context"
	p4rtapi "github.com/onosproject/onos-api/go/onos/p4rt/v1"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/controller"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/wcmp-app/pkg/controller/utils"
	"github.com/onosproject/wcmp-app/pkg/pluginregistry"
	"github.com/onosproject/wcmp-app/pkg/southbound/p4rt"
	pipelineConfigStore "github.com/onosproject/wcmp-app/pkg/store/pipelineconfig"
	"github.com/onosproject/wcmp-app/pkg/store/topo"
	p4configapi "github.com/p4lang/p4runtime/go/p4/config/v1"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/protobuf/proto"
	"time"
)

var log = logging.GetLogger()

const (
	defaultTimeout = 30 * time.Second
)

// NewController returns a new device pipeline pipelineconfig controller
func NewController(topo topo.Store, conns p4rt.ConnManager, p4PluginRegistry pluginregistry.P4PluginRegistry, pipelineConfigStore pipelineConfigStore.Store) *controller.Controller {
	c := controller.NewController("pipelineconfig")
	c.Watch(&TopoWatcher{
		topo: topo,
	})
	c.Watch(&Watcher{
		pipelineConfigs: pipelineConfigStore,
	})
	c.Reconcile(&Reconciler{
		conns:               conns,
		topo:                topo,
		p4PluginRegistry:    p4PluginRegistry,
		pipelineConfigStore: pipelineConfigStore,
	})
	return c
}

// Reconciler reconciles device pipeline config
type Reconciler struct {
	conns               p4rt.ConnManager
	topo                topo.Store
	p4PluginRegistry    pluginregistry.P4PluginRegistry
	pipelineConfigStore pipelineConfigStore.Store
}

// Reconcile reconciles pipeline configuration
func (r *Reconciler) Reconcile(id controller.ID) (controller.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	pipelineConfigID := id.Value.(p4rtapi.PipelineConfigID)
	log.Infow("Reconcile pipeline configuration", "pipelineConfigID", pipelineConfigID)
	pipelineConfig, err := r.pipelineConfigStore.Get(ctx, pipelineConfigID)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Warnw("Failed to reconcile Pipeline Configuration", "pipelineConfig ID", pipelineConfigID, "error", err)
			return controller.Result{}, err
		}
		log.Warnw("Failed to reconcile pipeline configuration; Pipeline Configuration not found", "pipelineConfigID", pipelineConfigID)
		return controller.Result{}, nil
	}
	if pipelineConfig.Status.State != p4rtapi.PipelineConfigStatus_PENDING {
		log.Warnw("Failed to reconcile Pipeline Configuration", "pipelineConfig ID", pipelineConfigID, "state", pipelineConfig.Status.State)
		return controller.Result{}, nil
	}
	if pipelineConfig.Spec == nil || len(pipelineConfig.Spec.P4Info) == 0 || len(pipelineConfig.Spec.P4DeviceConfig) == 0 {
		log.Errorw("Failed Reconciling device pipeline config for target", "pipelineConfig ID", pipelineConfig.ID, "targetID", pipelineConfig.TargetID, "error", err)
		return controller.Result{}, nil
	}

	log.Infow("Reconciling Device Pipeline Config for target", "pipelineConfig ID", pipelineConfig.ID, "targetID", pipelineConfig.TargetID)

	switch pipelineConfig.Action {
	case p4rtapi.ConfigurationAction_VERIFY_AND_COMMIT:
		log.Infow("Reconciling device pipeline config for VERIFY_AND_COMMIT action", "pipelineConfig ID", pipelineConfig.ID)
		return r.reconcileVerifyAndCommitAction(ctx, pipelineConfig)
	}

	return controller.Result{}, nil

}
func (r *Reconciler) reconcileVerifyAndCommitAction(ctx context.Context, pipelineConfig *p4rtapi.PipelineConfig) (controller.Result, error) {
	targetID := topoapi.ID(pipelineConfig.TargetID)
	target, err := r.topo.Get(ctx, targetID)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Errorw("Failed Reconciling device pipeline config for target", "targetID", targetID, "error", err)
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

	mastership := topoapi.P4RTMastershipState{}
	_ = target.GetAspect(&mastership)

	// If the master node ID is not set, skip reconciliation.
	if mastership.NodeId == "" {
		log.Infow("No master for target", "targetID", targetID)
		return controller.Result{}, nil
	}

	if len(p4rtServerInfo.Pipelines) == 0 {
		log.Errorw("Failed Reconciling device pipeline config for target", "targetID", targetID, "error", err)
		return controller.Result{}, errors.NewNotFound("Device pipeline config info is not found", targetID)
	}
	// If we've made it this far, we know there's a master relation.
	// Get the relation and check whether this node is the source
	relation, err := r.topo.Get(ctx, topoapi.ID(mastership.NodeId))
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Errorw("Failed fetching master Relation  from topo", "mastership node ID", mastership.NodeId, "error", err)
			return controller.Result{}, err
		}
		log.Warnw("Master relation not found for target", "targetID", targetID)
		return controller.Result{}, nil
	}
	if relation.GetRelation().SrcEntityID != utils.GetControllerID() {
		log.Debugw("Not the master for target", "targetID", targetID)
		return controller.Result{}, nil
	}

	// Get the master connection
	conn, ok := r.conns.Get(ctx, p4rt.ConnID(relation.ID))
	if !ok {
		log.Warnw("P4RT Connection not found for target", "targetID", targetID)
		return controller.Result{}, nil
	}

	p4InfoBytes := pipelineConfig.Spec.P4Info
	p4DeviceConfig := pipelineConfig.Spec.P4DeviceConfig

	p4Info := &p4configapi.P4Info{}
	err = proto.Unmarshal(p4InfoBytes, p4Info)
	if err != nil {
		return controller.Result{}, err
	}

	config := &p4api.ForwardingPipelineConfig{
		P4Info:         p4Info,
		P4DeviceConfig: p4DeviceConfig,
	}

	_, err = conn.SetForwardingPipelineConfig(ctx, &p4api.SetForwardingPipelineConfigRequest{
		DeviceId: p4rtServerInfo.DeviceID,
		ElectionId: &p4api.Uint128{
			Low:  mastership.Term,
			High: 0,
		},
		Config: config,
		Action: p4api.SetForwardingPipelineConfigRequest_VERIFY_AND_COMMIT,
	})

	if err != nil {
		log.Errorw("Failed Reconciling device pipeline config for target", "targetID", targetID, "error", err)
		pipelineConfig.Status.State = p4rtapi.PipelineConfigStatus_FAILED
		err = r.pipelineConfigStore.Update(ctx, pipelineConfig)
		if err != nil {
			if !errors.IsNotFound(err) || !errors.IsConflict(err) {
				log.Errorw("Failed Reconciling device pipeline config for target", "targetID", targetID, "error", err)
				return controller.Result{}, err
			}
			return controller.Result{}, nil
		}
		return controller.Result{}, nil
	}
	pipelineConfig.Status.State = p4rtapi.PipelineConfigStatus_COMPLETE
	err = r.pipelineConfigStore.Update(ctx, pipelineConfig)
	if err != nil {
		if !errors.IsNotFound(err) || !errors.IsConflict(err) {
			log.Errorw("Failed Reconciling device pipeline config for target", "targetID", targetID, "error", err)
			return controller.Result{}, err
		}
		return controller.Result{}, nil
	}
	log.Infow("Device pipelineConfig is Set Successfully", "targetID", targetID)
	return controller.Result{}, nil
}
