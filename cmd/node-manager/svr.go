package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-cleanhttp"

	"github.com/bufbuild/connect-go"

	ipamv1 "github.com/clly/wireguard-cni/gen/wgcni/ipam/v1"
	"github.com/clly/wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
	wireguardv1 "github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1"
	"github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"
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

type NodeManagerServer struct {
	*server.Server
	wgManager wireguard.WireguardManager
	cancelers []context.CancelFunc
}

func NewNodeManagerServer(ctx context.Context, cfg NodeConfig) (*NodeManagerServer, error) {
	ipamClient := ipamv1connect.NewIPAMServiceClient(cleanhttp.DefaultClient(), cfg.ClusterManagerAddr)
	wireguardClient := wireguardv1connect.NewWireguardServiceClient(cleanhttp.DefaultClient(), cfg.ClusterManagerAddr)

	alloc, err := ipamClient.Alloc(context.Background(), connect.NewRequest(&ipamv1.AllocRequest{}))
	if err != nil {
		return nil, err
	}

	cidr := fmt.Sprintf("%s/%s", alloc.Msg.GetAlloc().Address, alloc.Msg.GetAlloc().Netmask)

	wireguardConfig := wireguard.Config{
		Route:    cidr,
		Endpoint: cfg.Wireguard.Endpoint,
	}

	postUpCmd := fmt.Sprintf(PostUp, cidr)
	postUpVar.Set(postUpCmd)
	postDownCmd := fmt.Sprintf(PostDown, cidr)
	postDownVar.Set(postDownCmd)
	wgCidrPrefix.Set(cidr)

	// This is a shitty circular dependency I've created. We need the self for the server to include ourselves in the
	// peers response but we also need the server to set our own configs, so now it's eventually consistent and I'm sad.
	// We can refactor it but probably later
	wgManager, err := wireguard.New(ctx, wireguardConfig, wireguardClient, wireguard.WithPost(postUpCmd, postDownCmd))
	if err != nil {
		return nil, fmt.Errorf("failed to start wireguard manager %w", err)
	}

	wgSelf := wgManager.Self()

	self := &wireguardv1.Peer{
		PublicKey: wgSelf.PublicKey,
		Endpoint:  wgSelf.Endpoint,
		Route:     wgSelf.AllowedIPs,
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	svr, err := server.NewServer(cidr, server.WithNodeConfig(self), server.WithDataDir(wd))
	if err != nil {
		return nil, err
	}

	selfAlloc, err := svr.Alloc(ctx, &connect.Request[ipamv1.AllocRequest]{})
	if err != nil {
		return nil, err
	}
	wgManager.SetAddress(selfAlloc.Msg.Alloc.Address)
	wgManager.SetPeerRegistry(svr)

	configFile := filepath.Join(cfg.ConfigDirectory, fmt.Sprintf("%s.conf", cfg.InterfaceName))

	if err = setConfig(wgManager, configFile); err != nil {
		log.Println("failed to write config file")
		return nil, err
	}

	if err = wgManager.Up(cfg.InterfaceName); err != nil {
		log.Println("failed to bring interface", cfg.InterfaceName, "up")
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
