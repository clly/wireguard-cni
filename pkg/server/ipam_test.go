package server

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/bufbuild/connect-go"
	goipam "github.com/metal-stack/go-ipam"
	"github.com/stretchr/testify/require"

	ipamv1 "github.com/clly/wireguard-cni/gen/wgcni/ipam/v1"
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
			s, err := NewServer("10.0.0.0/8")
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

		})
	}
}

func Test_NewIPAM(t *testing.T) {
	const parentPrefix = "10.0.0.0/8"
	testcases := map[string]struct {
		withDataDir bool
		prefixs     []string
	}{
		"EmptyDataDir": {
			withDataDir: false,
		},
		"WithDataDir": {
			withDataDir: true,
		},
		"WithStoredAllocs": {
			withDataDir: true,
			prefixs:     []string{"10.0.0.0/24", "10.0.1.0/24", "10.0.2.0/24"},
		},
	}
	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			r := require.New(t)
			ctx := context.Background()
			d := t.TempDir()
			ipam, err := newIPAM(context.Background(), d, parentPrefix)
			r.NoError(err)
			for _, s := range testcase.prefixs {
				_, err := ipam.AcquireSpecificChildPrefix(ctx, parentPrefix, s)
				r.NoError(err)
			}
			r.NoError(ipam.save(ctx))
			ipam, err = newIPAM(context.Background(), d, "10.0.0.0/8")
			r.NoError(err)
			r.NoError(ipam.save(ctx))
			if !testcase.withDataDir {
				r.NoFileExists(IpamDataFile)
			} else {
				r.FileExists(filepath.Join(d, IpamDataFile))
			}

			if testcase.prefixs != nil {
				actual, err := ipam.ReadAllPrefixCidrs(ctx)
				r.NoError(err)
				sort.Strings(actual)
				expectedPrefixes := append(testcase.prefixs, parentPrefix)
				sort.Strings(expectedPrefixes)
				r.EqualValues(expectedPrefixes, actual)
			}

		})
	}
}
