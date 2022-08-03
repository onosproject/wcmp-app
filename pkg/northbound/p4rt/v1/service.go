// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"github.com/onosproject/onos-lib-go/pkg/northbound"
	"github.com/onosproject/wcmp-app/pkg/pluginregistry"
	"github.com/onosproject/wcmp-app/pkg/southbound/p4rt"
	"github.com/onosproject/wcmp-app/pkg/store/pipelineconfig"
	"github.com/onosproject/wcmp-app/pkg/store/topo"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
	"sync"
)

// Service implements service for P4Runtime
type Service struct {
	northbound.Service
	p4PluginRegistry    pluginregistry.P4PluginRegistry
	pipelineConfigStore pipelineconfig.Store
	topo                topo.Store
	conns               p4rt.ConnManager
}

func NewService(
	p4PluginRegistry pluginregistry.P4PluginRegistry,
	pipelineConfigStore pipelineconfig.Store,
	topo topo.Store,
	conns p4rt.ConnManager) Service {
	return Service{
		p4PluginRegistry:    p4PluginRegistry,
		pipelineConfigStore: pipelineConfigStore,
		topo:                topo,
		conns:               conns,
	}

}

type Server struct {
	mu sync.RWMutex
	p4api.UnimplementedP4RuntimeServer
	p4PluginRegistry    pluginregistry.P4PluginRegistry
	pipelineConfigStore pipelineconfig.Store
	topo                topo.Store
	conns               p4rt.ConnManager
}

func (s *Server) Capabilities(ctx context.Context, request *p4api.CapabilitiesRequest) (*p4api.CapabilitiesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s Service) Register(r *grpc.Server) {
	p4api.RegisterP4RuntimeServer(r, &Server{
		p4PluginRegistry:    s.p4PluginRegistry,
		topo:                s.topo,
		conns:               s.conns,
		pipelineConfigStore: s.pipelineConfigStore,
	})
}
