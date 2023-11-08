package wireguard

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	wireguardv1 "github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1"
	"github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"
)

func Test_WGQuick_Config(t *testing.T) {
	testcases := []struct {
		name  string
		peers []*wireguardv1.Peer
	}{
		{
			name:  "NilPeers",
			peers: nil,
		},
		{
			name:  "EmptyPeers",
			peers: []*wireguardv1.Peer{},
		},
		{
			name: "OnePeer",
			peers: []*wireguardv1.Peer{
				{
					PublicKey: "abc=",
					Endpoint:  "192.168.1.2:51820",
					Route:     "10.0.0.0/24",
				},
			},
		},
		{
			name: "TwoPeer",
			peers: []*wireguardv1.Peer{
				{
					PublicKey: "abc=",
					Endpoint:  "192.168.1.2:51820",
					Route:     "10.0.0.0/24",
				},
				{
					PublicKey: "def=",
					Endpoint:  "192.168.1.3:51820",
					Route:     "10.0.0.1/24",
				},
			},
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
				addr:     "10.0.0.1",
			}

			pbPeer := &wireguardv1.Peer{
				PublicKey: wgManager.self().PublicKey,
				Endpoint:  wgManager.endpoint,
				Route:     "",
			}
			testcase.peers = append(testcase.peers, pbPeer)

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
			r.Equal(b.String(), strings.TrimSpace(string(golden)))
		})
	}
}
