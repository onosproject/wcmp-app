// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package topo

import (
	"context"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	toposdk "github.com/onosproject/onos-ric-sdk-go/pkg/topo"
	"time"
)

// ClientTopo R-NIB client interface
type ClientTopo interface {
	GetControlRelations() ([]topoapi.Object, error)
	GetSwitchEntities() ([]topoapi.Object, error)
	GetControllerEntities() ([]topoapi.Object, error)
}

// NewClientTopo creates a new topo SDK client
func NewClientTopo() (Client, error) {
	sdkClient, err := toposdk.NewClient()
	if err != nil {
		return Client{}, err
	}
	cl := Client{
		client: sdkClient,
	}
	return cl, nil
}

// Client topo SDK client
type Client struct {
	client toposdk.Client
}

func getFilter(kind string) *topoapi.Filters {
	controlRelationFilter := &topoapi.Filters{
		KindFilter: &topoapi.Filter{
			Filter: &topoapi.Filter_Equal_{
				Equal_: &topoapi.EqualFilter{
					Value: kind,
				},
			},
		},
	}
	return controlRelationFilter
}

// GetControlRelationFilter gets control relation filter
func GetControlRelationFilter() *topoapi.Filters {
	return getFilter(topoapi.ControlsKind)
}

// GetControlRelations returns a list of the control relations
func (c *Client) GetControlRelations() ([]topoapi.Object, error) {
	filter := GetControlRelationFilter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	relationsList, err := c.client.List(ctx, toposdk.WithListFilters(filter))
	cancel()
	return relationsList, err
}

// GetSwitchFilter gets switch filter
func GetSwitchFilter() *topoapi.Filters {
	return getFilter(topoapi.SwitchKind)
}

// GetSwitchEntities returns a list of the switches
func (c *Client) GetSwitchEntities() ([]topoapi.Object, error) {
	filter := GetSwitchFilter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	switchList, err := c.client.List(ctx, toposdk.WithListFilters(filter))
	cancel()
	return switchList, err
}

// GetControllerFilter gets controller filter
func GetControllerFilter() *topoapi.Filters {
	return getFilter(topoapi.ControllerKind)
}

// GetControllerEntities returns a list of the controllers
func (c *Client) GetControllerEntities() ([]topoapi.Object, error) {
	filter := GetControllerFilter()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	controllerList, err := c.client.List(ctx, toposdk.WithListFilters(filter))
	cancel()
	return controllerList, err
}
