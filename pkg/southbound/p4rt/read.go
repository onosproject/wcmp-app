// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package p4rt

import (
	"context"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
	"io"
)

type ReadClient interface {
	ReadEntities(ctx context.Context, request *p4api.ReadRequest, opts ...grpc.CallOption) ([]*p4api.Entity, error)
}

type readClient struct {
	p4runtimeClient p4api.P4RuntimeClient
}

func (r readClient) ReadEntities(ctx context.Context, request *p4api.ReadRequest, opts ...grpc.CallOption) ([]*p4api.Entity, error) {
	stream, err := r.p4runtimeClient.Read(ctx, request, opts...)
	if err != nil {
		return nil, err
	}
	var entities []*p4api.Entity
	for {
		rep, err := stream.Recv()
		if err == io.EOF || err == context.Canceled {
			break
		}
		if err != nil {
			return nil, err
		}
		for _, e := range rep.Entities {
			entities = append(entities, e)
		}
	}
	return entities, nil
}

var _ ReadClient = &readClient{}
