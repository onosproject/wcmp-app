// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package p4rt

import (
	"context"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
	"io"
)

type StreamChannelClient interface {
	SetMasterArbitration(ctx context.Context, deviceID uint64, electionID uint64) (*p4api.StreamMessageResponse_Arbitration, error)
}

type streamChannelClient struct {
	p4runtimeClient p4api.P4RuntimeClient
}

func (s *streamChannelClient) SetMasterArbitration(ctx context.Context, deviceID uint64, electionID uint64) (*p4api.StreamMessageResponse_Arbitration, error) {
	streamChannel, err := s.p4runtimeClient.StreamChannel(ctx)
	if err != nil {
		return nil, err
	}
	defer streamChannel.CloseSend()
	errCh := make(chan error)
	var result *p4api.StreamMessageResponse_Arbitration
	go func() {
		for {
			in, err := streamChannel.Recv()
			if err == io.EOF {
				errCh <- nil
			}
			if err != nil {
				errCh <- err
			}
			arbitration, ok := in.Update.(*p4api.StreamMessageResponse_Arbitration)
			if !ok {
				continue
			}
			result = arbitration
			errCh <- nil
		}

	}()

	request := &p4api.StreamMessageRequest{
		Update: &p4api.StreamMessageRequest_Arbitration{Arbitration: &p4api.MasterArbitrationUpdate{
			DeviceId: deviceID,
			ElectionId: &p4api.Uint128{
				Low:  electionID,
				High: 0,
			},
		}},
	}
	err = streamChannel.Send(request)
	if err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return result, err

	}
}

var _ StreamChannelClient = &streamChannelClient{}
