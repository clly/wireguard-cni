package server

import (
	"context"
	"log"

	"github.com/bufbuild/connect-go"
	ipamv1 "github.com/clly/wireguard-cni/gen/wgcni/ipam/v1"
	"github.com/clly/wireguard-cni/pkg/ipam"
)

func (s *Server) Alloc(
	ctx context.Context,
	req *connect.Request[ipamv1.AllocRequest],
) (*connect.Response[ipamv1.AllocResponse], error) {

	alloc := &ipamv1.IPAlloc{
		Netmask: "24",
		Version: ipamv1.IPVersion_IP_VERSION_V4,
	}

	switch s.mode {
	case ipam.CLUSTER_MODE:
		prefix, err := s.ipam.AcquireChildPrefix(ctx, s.ipam.Prefix.Cidr, 24)
		if err != nil {
			return nil, err
		}
		ip, err := prefix.Network()
		if err != nil {
			return nil, err
		}
		alloc.Address = ip.String()
	case ipam.NODE_MODE:
		alloc.Netmask = "32"
		ip, err := s.ipam.AcquireIP(ctx, s.ipam.Prefix.Cidr)
		if err != nil {
			return nil, err
		}
		alloc.Address = ip.IP.String()
	}
	response := &ipamv1.AllocResponse{
		Alloc: alloc,
	}

	if err := s.ipam.Save(ctx); err != nil {
		return nil, err
	}
	log.Printf("Allocated new /%s CIDR %s\n", alloc.Netmask, alloc.Address)

	return connect.NewResponse(response), nil
}
