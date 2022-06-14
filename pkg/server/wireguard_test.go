package server

import (
	"context"
	"errors"
	"testing"
	wireguardv1 "wireguard-cni/gen/wgcni/wireguard/v1"

	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/require"
)

func Test_Peers(t *testing.T) {
	testcases := []struct {
		name string
		req  *wireguardv1.PeersRequest
		resp *wireguardv1.PeersResponse
		err  error
	}{
		{
			name: "HappyPath",
			req:  &wireguardv1.PeersRequest{},
			resp: &wireguardv1.PeersResponse{},
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			r := require.New(t)
			s := NewServer()
			expectedResponse := connect.NewResponse(testcase.resp)
			req := connect.NewRequest(testcase.req)
			resp, err := s.Peers(context.Background(), req)
			if testcase.err != nil {
				r.Error(err)
				r.EqualError(testcase.err, err.Error())
			} else {
				r.Nil(err)
				r.Equal(expectedResponse, resp)
			}
		})
	}
}

func Test_Register(t *testing.T) {
	testcases := []struct {
		name string
		req  *wireguardv1.RegisterRequest
		resp *wireguardv1.RegisterResponse
		err  error
	}{
		{
			name: "HappyPath",
			req: &wireguardv1.RegisterRequest{
				PublicKey: "abc123",
				Endpoint:  "192.168.1.1:51820",
				Route:     "10.0.0.9/24",
			},
			resp: &wireguardv1.RegisterResponse{},
		},
		{
			name: "MissingPK",
			req: &wireguardv1.RegisterRequest{
				PublicKey: "",
				Endpoint:  "192.168.1.1:51820",
				Route:     "10.0.0.0/24",
			},
			resp: nil,
			err:  connect.NewError(connect.CodeFailedPrecondition, errors.New("public_key: cannot be blank.")),
		},
		{
			name: "MissingEndpoint",
			req: &wireguardv1.RegisterRequest{
				PublicKey: "abc123",
				Endpoint:  "",
				Route:     "10.0.0.0/24",
			},
			resp: nil,
			err:  connect.NewError(connect.CodeFailedPrecondition, errors.New("endpoint: cannot be blank.")),
		},
		{
			name: "MissingRoute",
			req: &wireguardv1.RegisterRequest{
				PublicKey: "abc123",
				Endpoint:  "192.168.1.1:51820",
				Route:     "",
			},
			resp: nil,
			err:  connect.NewError(connect.CodeFailedPrecondition, errors.New("route: cannot be blank.")),
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			r := require.New(t)
			s := NewServer()
			expectedResponse := connect.NewResponse(testcase.resp)
			req := connect.NewRequest(testcase.req)
			resp, err := s.Register(context.Background(), req)
			if testcase.err != nil {
				r.Error(err)
				r.EqualError(testcase.err, err.Error())
			} else {
				r.Nil(err)
				r.Equal(expectedResponse, resp)
			}
		})
	}
}
