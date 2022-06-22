package wireguard

import (
	"context"
	"testing"
	wireguardv1 "wireguard-cni/gen/wgcni/wireguard/v1"
	"wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "Happy",
		},
	}

	for _, testcase := range tests {
		t.Run(testcase.name, func(t *testing.T) {
			r := require.New(t)
			wireguardM := &wireguardv1connect.MockWireguardServiceClient{}
			defer wireguardM.AssertExpectations(t)

			cfg := Config{
				Endpoint: "127.0.0.1:51820",
				Route:    "192.168.1.1/24",
			}

			wireguardM.On("Register", mock.Anything, mock.Anything).
				Return(nil, nil)

			err := New(context.Background(), cfg, wireguardM)
			r.NoError(err)

			_ = &wireguardv1.RegisterRequest{
				PublicKey: mock.Anything,
				Endpoint:  cfg.Endpoint,
				Route:     cfg.Route,
			}

			if testcase.err != nil {
				r.EqualError(err, testcase.err.Error())
			} else {
				r.NoError(err)
			}

		})
	}

}
