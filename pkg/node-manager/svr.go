package nodemanager

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/hashicorp/go-cleanhttp"
	"golang.zx2c4.com/wireguard/wgctrl"

	"connectrpc.com/connect"

	ipamv1 "github.com/clly/wireguard-cni/gen/wgcni/ipam/v1"
	"github.com/clly/wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
	wireguardv1 "github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1"
	"github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"
	"github.com/clly/wireguard-cni/pkg/ipam"
	"github.com/clly/wireguard-cni/pkg/server"
	"github.com/clly/wireguard-cni/pkg/wireguard"
)

func init() {
	expvar.Publish("ipam-cidr", &wgCidrPrefix)
	expvar.Publish("iptables-post-up", &postUpVar)
	expvar.Publish("iptables-post-down", &postDownVar)
}

var (
	wgCidrPrefix expvar.String
	postUpVar    expvar.String
	postDownVar  expvar.String
)

const (
	PostUp   = "iptables -A FORWARD -i %%i -j ACCEPT; iptables -A FORWARD -o %%i -j ACCEPT; iptables -t nat -A POSTROUTING -j MASQUERADE -s %s"
	PostDown = "iptables -D FORWARD -i %%i -j ACCEPT; iptables -D FORWARD -o %%i -j ACCEPT; iptables -t nat -D POSTROUTING -s %s -j MASQUERADE"
)

type NodeConfig struct {
	ClusterManagerAddr string
	ConfigDirectory    string
	ListenAddr         string
	DataDirectory      string
	Wireguard          WireguardNodeConfig
}

// WireguardNodeConfig contains all the information that gets fed into the wireguard manager at some point
type WireguardNodeConfig struct {
	Endpoint      string
	InterfaceName string
}
type NodeManagerServer struct {
	*server.Server
	wgManager wireguard.WireguardManager
	cancelers []context.CancelFunc
}

func NewNodeManagerServer(ctx context.Context, cfg NodeConfig) (*NodeManagerServer, error) {
	ipamClient := ipamv1connect.NewIPAMServiceClient(cleanhttp.DefaultClient(), cfg.ClusterManagerAddr)
	wireguardClient := wireguardv1connect.NewWireguardServiceClient(cleanhttp.DefaultClient(), cfg.ClusterManagerAddr)

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.DataDirectory == "" {
		cfg.DataDirectory = path.Join(wd, "node-manager")
	}

	clusterIpam, err := ipam.NewRemoteIPAM(context.TODO(), cfg.DataDirectory, ipamClient)
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to alloc ip range from cluster-manager %w", err)
	}

	cidr := clusterIpam.Prefix.Cidr

	wireguardConfig := wireguard.Config{
		Route:    cidr,
		Endpoint: cfg.Wireguard.Endpoint,
	}

	postUpCmd := fmt.Sprintf(PostUp, cidr)
	postUpVar.Set(postUpCmd)
	postDownCmd := fmt.Sprintf(PostDown, cidr)
	postDownVar.Set(postDownCmd)
	wgCidrPrefix.Set(cidr)

	wgclient, err := wgctrl.New()
	if err != nil {
		return nil, err
	}

	// This is a shitty circular dependency I've created. We need the self for the server to include ourselves in the
	// peers response but we also need the server to set our own configs, so now it's eventually consistent and I'm sad.
	// We can refactor it but probably later
	wgManager, err := wireguard.New(ctx, wireguardConfig, wgclient, wireguardClient, wireguard.WithPost(postUpCmd, postDownCmd))
	if err != nil {
		return nil, fmt.Errorf("failed to start wireguard manager %w", err)
	}

	wgSelf := wgManager.Self()

	self := &wireguardv1.Peer{
		PublicKey: wgSelf.PublicKey,
		Endpoint:  wgSelf.Endpoint,
		Route:     wgSelf.AllowedIPs,
	}

	svr, err := server.NewServer(clusterIpam, server.WithNodeConfig(self, ipamClient), server.WithDataDir(cfg.DataDirectory))

	if err != nil {
		return nil, err
	}

	selfAlloc, err := svr.Alloc(ctx, &connect.Request[ipamv1.AllocRequest]{})
	if err != nil {
		return nil, err
	}
	wgManager.SetAddress(selfAlloc.Msg.Alloc.Address)
	wgManager.SetPeerRegistry(svr)

	configFile := filepath.Join(cfg.ConfigDirectory, fmt.Sprintf("%s.conf", cfg.Wireguard.InterfaceName))

	if err = setConfig(wgManager, configFile); err != nil {
		log.Println("failed to write config file")
		return nil, err
	}

	if err = wgManager.Up(cfg.Wireguard.InterfaceName); err != nil {
		log.Println("failed to bring interface", cfg.Wireguard.InterfaceName, "up")
		return nil, err
	}

	peerCtx, cancel := context.WithCancel(ctx)
	go func() {
		err = peerMgr(peerCtx, wgManager, configFile)
		if !errors.Is(err, context.Canceled) {
			panic(err)
		}
	}()

	return &NodeManagerServer{
		Server:    svr,
		wgManager: wgManager,
		cancelers: []context.CancelFunc{cancel},
	}, nil
}

func (m *NodeManagerServer) Down(device string) error {
	for _, cancel := range m.cancelers {
		cancel()
	}
	return m.wgManager.Down(device)
}
