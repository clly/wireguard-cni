package main

import (
	"context"
	"fmt"
	ipamv1 "wireguard-cni/gen/wgcni/ipam/v1"
	"wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
	"wireguard-cni/pkg/server"

	"github.com/bufbuild/connect-go"
)

type NodeManagerServer struct {
	*server.Server
}

func NewNodeManagerServer(ipamClient ipamv1connect.IPAMServiceClient) (*NodeManagerServer, error) {
	alloc, err := ipamClient.Alloc(context.Background(), connect.NewRequest(&ipamv1.AllocRequest{}))
	if err != nil {
		return nil, err
	}
	cidr := fmt.Sprintf("%s/%s", alloc.Msg.GetAlloc().Address, alloc.Msg.GetAlloc().Netmask)

	svr, err := server.NewServer(cidr, server.NODE_MODE)
	if err != nil {
		return nil, err
	}

	return &NodeManagerServer{
		Server: svr,
	}, nil
}
