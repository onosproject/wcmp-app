// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package p4rt

import (
	"context"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
	"io"
)

// Client P4runtime client interface
type Client interface {
	io.Closer
	WriteClient
	ReadClient
	StreamChannelClient
	PipelineConfigClient
	Capabilities(ctx context.Context, request *p4api.CapabilitiesRequest, opts ...grpc.CallOption) (*p4api.CapabilitiesResponse, error)
}

type client struct {
	p4runtimeClient      p4api.P4RuntimeClient
	writeClient          *writeClient
	readClient           *readClient
	pipelineConfigClient *pipelineConfigClient
	streamChannelClient  *streamChannelClient
}

func (c *client) SetMasterArbitration(ctx context.Context, deviceID uint64, electionID uint64) (*p4api.StreamMessageResponse_Arbitration, error) {
	arbitrationResponse, err := c.streamChannelClient.SetMasterArbitration(ctx, deviceID, electionID)
	return arbitrationResponse, errors.FromGRPC(err)
}

func (c *client) ReadEntities(ctx context.Context, request *p4api.ReadRequest, opts ...grpc.CallOption) ([]*p4api.Entity, error) {
	log.Infow("Received read entities request", "request", request)
	entities, err := c.readClient.ReadEntities(ctx, request, opts...)
	if err != nil {
		return nil, errors.FromGRPC(err)
	}
	return entities, nil

}

// Write Updates one or more P4 entities on the target.
func (c *client) Write(ctx context.Context, request *p4api.WriteRequest, opts ...grpc.CallOption) (*p4api.WriteResponse, error) {
	log.Infow("Received Write request", "request", request)
	writeResponse, err := c.writeClient.Write(ctx, request, opts...)
	return writeResponse, errors.FromGRPC(err)
}

// SetForwardingPipelineConfig  sets the P4 forwarding-pipeline config.
func (c *client) SetForwardingPipelineConfig(ctx context.Context, request *p4api.SetForwardingPipelineConfigRequest, opts ...grpc.CallOption) (*p4api.SetForwardingPipelineConfigResponse, error) {
	log.Infow("Received SetForwardingPipelineConfig request", "request", request)
	setForwardingPipelineConfigResponse, err := c.pipelineConfigClient.SetForwardingPipelineConfig(ctx, request, opts...)
	return setForwardingPipelineConfigResponse, errors.FromGRPC(err)
}

// GetForwardingPipelineConfig  gets the current P4 forwarding-pipeline config.
func (c *client) GetForwardingPipelineConfig(ctx context.Context, request *p4api.GetForwardingPipelineConfigRequest, opts ...grpc.CallOption) (*p4api.GetForwardingPipelineConfigResponse, error) {
	log.Infow("Received GetForwardingPipelineConfig request", "request", request)
	getForwardingPipelineConfigResponse, err := c.pipelineConfigClient.GetForwardingPipelineConfig(ctx, request, opts...)
	return getForwardingPipelineConfigResponse, errors.FromGRPC(err)
}

// StreamChannel Represents the bidirectional stream between the controller and the
// switch (initiated by the controller), and is managed for the following
// purposes:
// - connection initiation through client arbitration
// - indicating switch session liveness: the session is live when switch
// - sends a positive client arbitration update to the controller, and is
// - considered dead when either the stream breaks or the switch sends a
// - negative update for client arbitration
// - the controller sending/receiving packets to/from the switch
// - streaming of notifications from the switch
func (c *client) StreamChannel(ctx context.Context, opts ...grpc.CallOption) (p4api.P4Runtime_StreamChannelClient, error) {
	streamChannelClient, err := c.p4runtimeClient.StreamChannel(ctx, opts...)
	return streamChannelClient, errors.FromGRPC(err)
}

// Capabilities discovers the capabilities of the P4Runtime server implementation.
func (c *client) Capabilities(ctx context.Context, request *p4api.CapabilitiesRequest, opts ...grpc.CallOption) (*p4api.CapabilitiesResponse, error) {
	log.Infow("Received Capabilities request", "request", request)
	capabilitiesResponse, err := c.p4runtimeClient.Capabilities(ctx, request, opts...)
	return capabilitiesResponse, errors.FromGRPC(err)
}

func (c *client) Close() error {
	return nil
}

var _ Client = &client{}
