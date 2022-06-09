package server

import (
	"context"
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
			s := NewIPAMServer()
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
			req:  &wireguardv1.RegisterRequest{},
			resp: &wireguardv1.RegisterResponse{},
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			r := require.New(t)
			s := NewIPAMServer()
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
