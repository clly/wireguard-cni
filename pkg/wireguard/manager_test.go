package wireguard

import (
	"bytes"
	"testing"
	"wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_WGQuick_Config(t *testing.T) {
	testcases := []struct {
		name string
	}{
		{
			name: "Happy",
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			r := require.New(t)
			wireguardM := &wireguardv1connect.MockWireguardServiceClient{}
			wgManager := &WGQuickManager{
				client: wireguardM,
			}

			wireguardM.On("Peers", mock.Anything, mock.Anything).
				Once().
				Return(nil, nil)

			b := bytes.NewBuffer(make([]byte, 0, 1024))
			err := wgManager.Config(b)
			r.NoError(err)
		})
	}
}
