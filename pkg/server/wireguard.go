package server

import (
	"context"
	wireguardv1 "wireguard-cni/gen/wgcni/wireguard/v1"
	"wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"

	"github.com/bufbuild/connect-go"
)

var (
	_ wireguardv1connect.WireguardServiceHandler = &Server{}
)

func (s *Server) Register(ctx context.Context,
	req *connect.Request[wireguardv1.RegisterRequest],
) (*connect.Response[wireguardv1.RegisterResponse], error) {
	return nil, nil
}

func (s *Server) Peers(ctx context.Context,
	req *connect.Request[wireguardv1.PeersRequest],
) (*connect.Response[wireguardv1.PeersResponse], error) {
	p := &wireguardv1.PeersResponse{}
	return connect.NewResponse(p), nil
}
