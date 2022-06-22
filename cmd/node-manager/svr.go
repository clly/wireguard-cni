package main

import (
	"context"
	"expvar"
	"fmt"
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
}

func NewNodeManagerServer(ctx context.Context, cfg NodeConfig, ipamClient ipamv1connect.IPAMServiceClient, wireguardClient wireguardv1connect.WireguardServiceClient) (*NodeManagerServer, error) {
	alloc, err := ipamClient.Alloc(context.Background(), connect.NewRequest(&ipamv1.AllocRequest{}))
	if err != nil {
		return nil, err
	}
	cidr := fmt.Sprintf("%s/%s", alloc.Msg.GetAlloc().Address, alloc.Msg.GetAlloc().Netmask)
	cfg.Wireguard.Route = cidr
	wgCidrPrefix.Set(cidr)

	err = wireguard.New(ctx, cfg.Wireguard, wireguardClient)
	if err != nil {
		return nil, fmt.Errorf("failed to start wireguard manager %w", err)
	}

	svr, err := server.NewServer(cidr, server.NODE_MODE)
	if err != nil {
		return nil, err
	}

	return &NodeManagerServer{
		Server: svr,
	}, nil
}
