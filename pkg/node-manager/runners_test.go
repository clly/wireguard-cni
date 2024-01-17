package nodemanager

import (
	"context"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	wireguardv1 "github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1"
	"github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"
	"github.com/clly/wireguard-cni/pkg/wireguard"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const cfgFile = "hack/wg0.conf"

type TestManager struct {
	wireguard.WireguardManager
}

func (t *TestManager) SetPeers(device string, peers []*wireguard.Peer) error {
	return nil
}

var _ wireguard.WireguardManager = (*TestManager)(nil)

func Test_PeerManagerRunner(t *testing.T) {
	r := require.New(t)
	clientM := &wireguardv1connect.MockWireguardServiceClient{}
	wireguardM := &wireguard.MockWGClient{}

	t.Cleanup(func() {
		os.Remove(cfgFile)
	})

	peers := []*wireguardv1.Peer{
		{
			PublicKey: "abc=",
			Endpoint:  "192.168.1.2:51820",
			Route:     "10.0.1.0/24",
		},
	}

	clientM.On("Register", mock.Anything, mock.Anything).
		Once().
		Return(nil, nil)

	clientM.On("Peers", mock.Anything, mock.Anything).
		Once().
		Return(connect.NewResponse(&wireguardv1.PeersResponse{
			Peers: peers,
		}), nil)

	clientM.On("Peers", mock.Anything, mock.Anything).
		Maybe().
		Return(connect.NewResponse(&wireguardv1.PeersResponse{
			Peers: peers,
		}), nil)

	wireguardM.On("Device", mock.Anything).
		Once().
		Return(&wgtypes.Device{}, nil)

	wireguardM.On("Device", mock.Anything).
		Once().
		Return(&wgtypes.Device{}, nil)

	defer clientM.AssertExpectations(t)

	// how do I eventually make it so that SetPeers doesn't call wg-quick??
	mgr, err := wireguard.New(context.Background(), wireguard.Config{
		Endpoint: "192.168.1.1:51820",
		Route:    "10.0.0.0/24",
	}, wireguardM, clientM)
	r.NoError(err)

	// tmgr := &TestManager{
	// 	WireguardManager: mgr,
	// }

	r.NoError(setConfig(mgr, cfgFile))
	// go func() {
	// 	r.NoError(peerMgr(context.Background(), tmgr, cfgFile))
	// }()

	time.Sleep(2 * time.Second)
	_, err = os.Stat(cfgFile)
	r.NoError(err)
}

func Test_deviceFromFile(t *testing.T) {
	device := deviceFromConf(cfgFile)
	require.Equal(t, "wg0", device)
}
