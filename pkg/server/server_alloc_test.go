package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"

	ipamv1 "github.com/clly/wireguard-cni/gen/wgcni/ipam/v1"
	"github.com/clly/wireguard-cni/pkg/ipam"
)

func Test_Alloc(t *testing.T) {
	testcases := []struct {
		name    string
		req     *ipamv1.AllocRequest
		resp    *ipamv1.AllocResponse
		dataDir func(s string) newServerOpt
		err     error
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
		{
			name: "HappyWithDataFile",
			req:  &ipamv1.AllocRequest{},
			resp: &ipamv1.AllocResponse{
				Alloc: &ipamv1.IPAlloc{
					Address: "10.0.0.0",
					Netmask: "24",
					Version: ipamv1.IPVersion_IP_VERSION_V4,
				},
			},
			dataDir: func(s string) newServerOpt {
				return WithDataDir(s)
			},
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			r := require.New(t)
			ctx := context.Background()

			// setup serverOpt
			var opts = make([]newServerOpt, 0)

			var dir string
			if testcase.dataDir != nil {
				opts = append(opts, WithDataDir(dir))

				dir = t.TempDir()
				t.Cleanup(func() {
					r.NoError(os.RemoveAll(dir))
				})
			}
			clusterIpam, err := ipam.New(ctx, dir, "10.0.0.0/8")
			r.NoError(err)

			s, err := NewServer(clusterIpam, opts...)
			r.NoError(err)
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

			if testcase.dataDir != nil {
				dataFile := filepath.Join(dir, ipam.IpamDataFile)
				r.FileExists(dataFile)

				// load data file
				ipam, err := ipam.New(ctx, dir, "10.0.0.0/8")
				r.NoError(err)

				prefixes, err := ipam.ReadAllPrefixCidrs(ctx)
				r.NoError(err)
				r.Contains(prefixes, "10.0.0.0/8")
				r.Contains(prefixes, "10.0.0.0/24")
			}
		})
	}
}
