package server

import (
	"context"
	"testing"
	ipamv1 "wireguard-cni/gen/wgcni/ipam/v1"

	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/require"
)

func Test_Alloc(t *testing.T) {
	testcases := []struct {
		name string
		req  *ipamv1.AllocRequest
		resp *ipamv1.AllocResponse
		err  error
	}{
		{
			name: "HappyPath",
			req:  &ipamv1.AllocRequest{},
			resp: &ipamv1.AllocResponse{
				Alloc: &ipamv1.IPAlloc{
					Address: "10.0.0.0",
					Netmask: "24",
					Version: ipamv1.IPVersion_IP_VERSION_V4,
				},
			},
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			r := require.New(t)
			s := NewServer()
			expectedResponse := connect.NewResponse(testcase.resp)
			req := connect.NewRequest(testcase.req)
			resp, err := s.Alloc(context.Background(), req)
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
