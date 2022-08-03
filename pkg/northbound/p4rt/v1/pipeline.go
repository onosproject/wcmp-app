// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
)

// SetForwardingPipelineConfig Sets the P4 forwarding-pipeline config.
func (s *Server) SetForwardingPipelineConfig(ctx context.Context, request *p4api.SetForwardingPipelineConfigRequest) (*p4api.SetForwardingPipelineConfigResponse, error) {
	//TODO implement me
	panic("implement me")
}

// GetForwardingPipelineConfig the current P4 forwarding-pipeline config
func (s *Server) GetForwardingPipelineConfig(ctx context.Context, request *p4api.GetForwardingPipelineConfigRequest) (*p4api.GetForwardingPipelineConfigResponse, error) {
	//TODO implement me
	panic("implement me")
}
