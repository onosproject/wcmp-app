// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package p4rt

import (
	"context"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
)

type WriteClient interface {
	Write(ctx context.Context, request *p4api.WriteRequest, opts ...grpc.CallOption) (*p4api.WriteResponse, error)
}

type writeClient struct {
	p4runtimeClient p4api.P4RuntimeClient
}

func (w *writeClient) Write(ctx context.Context, request *p4api.WriteRequest, opts ...grpc.CallOption) (*p4api.WriteResponse, error) {
	response, err := w.p4runtimeClient.Write(ctx, request, opts...)
	return response, errors.FromGRPC(err)
}

var _ WriteClient = &writeClient{}
