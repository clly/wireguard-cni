package wireguard

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	wireguardv1 "wireguard-cni/gen/wgcni/wireguard/v1"
	"wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"

	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_WGQuick_Config(t *testing.T) {
	testcases := []struct {
		name  string
		peers []*wireguardv1.Peer
	}{
		{
			name: "EmptyPeers",
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			r := require.New(t)
			wireguardM := &wireguardv1connect.MockWireguardServiceClient{}
			wgManager := &WGQuickManager{
				client:   wireguardM,
				key:      [32]byte{},
				endpoint: "192.168.1.1:51820",
			}

			wireguardM.On("Peers", mock.Anything, mock.Anything).
				Once().
				Return(connect.NewResponse(&wireguardv1.PeersResponse{
					Peers: testcase.peers,
				}), nil)

			b := bytes.NewBuffer(make([]byte, 0, 1024))
			err := wgManager.Config(b)
			r.NoError(err)
			golden, err := ioutil.ReadFile(fmt.Sprintf("%s/%s.conf", "hack", testcase.name))
			r.NoError(err)
			r.Equal(b.String(), string(golden))
		})
	}
}
