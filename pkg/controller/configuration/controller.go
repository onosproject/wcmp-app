package configuration

import (
	"context"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/controller"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/wcmp-app/pkg/southbound/p4rt"
	"github.com/onosproject/wcmp-app/pkg/store/topo"
	"time"
)

var log = logging.GetLogger()

const (
	defaultTimeout = 30 * time.Second
)

// NewController returns a new device pipeline configuration controller
func NewController(topo topo.Store, conns p4rt.ConnManager) *controller.Controller {
	c := controller.NewController("configuration")
	c.Watch(&TopoWatcher{
		topo: topo,
	})
	c.Reconcile(&Reconciler{
		conns: conns,
		topo:  topo,
	})
	return c
}

// Reconciler reconciles device pipeline configuration
type Reconciler struct {
	conns p4rt.ConnManager
	topo  topo.Store
}

func (r *Reconciler) Reconcile(id controller.ID) (controller.Result, error) {

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	targetID := id.Value.(topoapi.ID)

	log.Infow("Reconciling device pipeline configuration for target", "targetID")

	return r.reconcileConfiguration(ctx, targetID)

}

func (r *Reconciler) reconcileConfiguration(ctx context.Context, targetID topoapi.ID) (controller.Result, error) {

	target, err := r.topo.Get(ctx, targetID)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Errorw("Failed fetching target Entity from topo", "targetID", targetID, "error", err)
			return controller.Result{}, err
		}

		return controller.Result{}, nil
	}

	return controller.Result{}, nil

}
