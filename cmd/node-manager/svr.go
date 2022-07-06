package main

import (
	"context"
	"expvar"
	"fmt"
	"log"
	"path/filepath"

	ipamv1 "wireguard-cni/gen/wgcni/ipam/v1"
	"wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
	"wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"
	"wireguard-cni/pkg/server"
	"wireguard-cni/pkg/wireguard"

	"github.com/bufbuild/connect-go"
)

func init() {
	expvar.Publish("ipam-cidr", &wgCidrPrefix)
}

var (
	wgCidrPrefix expvar.String
)

type NodeManagerServer struct {
	*server.Server
	wgManager wireguard.WireguardManager
	cancelers []context.CancelFunc
}

func NewNodeManagerServer(ctx context.Context, cfg NodeConfig, ipamClient ipamv1connect.IPAMServiceClient, wireguardClient wireguardv1connect.WireguardServiceClient) (*NodeManagerServer, error) {
	alloc, err := ipamClient.Alloc(context.Background(), connect.NewRequest(&ipamv1.AllocRequest{}))
	if err != nil {
		return nil, err
	}
	cidr := fmt.Sprintf("%s/%s", alloc.Msg.GetAlloc().Address, alloc.Msg.GetAlloc().Netmask)
	cfg.Wireguard.Route = cidr
	wgCidrPrefix.Set(cidr)

	wgManager, err := wireguard.New(ctx, cfg.Wireguard, wireguardClient)
	if err != nil {
		return nil, fmt.Errorf("failed to start wireguard manager %w", err)
	}

	svr, err := server.NewServer(cidr, server.NODE_MODE)
	if err != nil {
		return nil, err
	}

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
		panic(err)
	}()

	return &NodeManagerServer{
		Server:    svr,
		wgManager: wgManager,
		cancelers: []context.CancelFunc{cancel},
	}, nil
}
