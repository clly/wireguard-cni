package wireguard

import (
	"context"
	"testing"

	"github.com/bufbuild/connect-go"
	wireguardv1 "github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1"
	"github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var _ WireguardManager = (*WGQuickManager)(nil)

func Test_New(t *testing.T) {
	tests := map[string]struct {
		err       error
		namespace string
	}{
		"Happy": {},
		"WithNamespace": {
			namespace: "/ns/namespace/name",
		},
	}

	for name, testcase := range tests {
		t.Run(name, func(t *testing.T) {
			r := require.New(t)
			wireguardM := &wireguardv1connect.MockWireguardServiceClient{}
			wgclientM := &MockWGClient{}
			defer wireguardM.AssertExpectations(t)

			cfg := Config{
				Endpoint: "127.0.0.1:51820",
				Route:    "192.168.1.1/24",
			}

			req := &wireguardv1.RegisterRequest{
				PublicKey: mock.Anything,
				Endpoint:  cfg.Endpoint,
				Route:     cfg.Route,
			}

			if testcase.namespace != "" {
				req.Namespace = testcase.namespace
				cfg.Namespace = testcase.namespace
			}

			wireguardM.On("Register", mock.Anything, mock.MatchedBy(func(req *connect.Request[wireguardv1.RegisterRequest]) bool {
				t := req.Msg.Endpoint == cfg.Endpoint &&
					req.Msg.Namespace == cfg.Namespace &&
					req.Msg.Route == cfg.Route
				return t
			})).
				Return(nil, nil)

			wgclientM.On("Device", mock.Anything).
				Once().
				Return(&wgtypes.Device{}, nil)

			// wgclientM.On("Device", mock.Anything).
			// 	Maybe().Twice().
			// 	Return(&wgtypes.Device{}, nil)

			wm, err := New(context.Background(), cfg, wgclientM, wireguardM)
			r.NoError(err)

			wm.stopCh <- struct{}{}

			if testcase.err != nil {
				r.EqualError(err, testcase.err.Error())
			} else {
				r.NoError(err)
			}

		})
	}

}
