package server

import (
	"context"
	"expvar"
	"log"
	"net/http"
	"sync"

	"github.com/bufbuild/connect-go"
	goipam "github.com/metal-stack/go-ipam"

	ipamv1 "github.com/clly/wireguard-cni/gen/wgcni/ipam/v1"
	"github.com/clly/wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
)

var (
	_               ipamv1connect.IPAMServiceHandler = &Server{}
	wireguardExpvar                                  = new(expvar.Map).Init()
	once                                             = &sync.Once{}
)

type IPAM_MODE int

const (
	CLUSTER_MODE IPAM_MODE = iota
	NODE_MODE
)

func init() {
	expvar.Publish("wireguard", wireguardExpvar)
}

func (s *Server) IPAMServiceHandler() (string, http.Handler) {
	return ipamv1connect.NewIPAMServiceHandler(s)
}

func ipamUsage(i goipam.Ipamer, cidrPrefix string) func() any {
	return func() any {
		return i.PrefixFrom(context.TODO(), cidrPrefix).Usage()
	}
}

func (s *Server) Alloc(
	ctx context.Context,
	req *connect.Request[ipamv1.AllocRequest],
) (*connect.Response[ipamv1.AllocResponse], error) {

	alloc := &ipamv1.IPAlloc{
		Netmask: "24",
		Version: ipamv1.IPVersion_IP_VERSION_V4,
	}

	switch s.mode {
	case CLUSTER_MODE:
		prefix, err := s.ipam.AcquireChildPrefix(ctx, s.prefix.Cidr, 24)
		if err != nil {
			return nil, err
		}
		ip, err := prefix.Network()
		if err != nil {
			return nil, err
		}
		alloc.Address = ip.String()
	case NODE_MODE:
		alloc.Netmask = "32"
		ip, err := s.ipam.AcquireIP(ctx, s.prefix.Cidr)
		if err != nil {
			return nil, err
		}
		alloc.Address = ip.IP.String()
	}
	response := &ipamv1.AllocResponse{
		Alloc: alloc,
	}

	log.Printf("Allocated new /%s CIDR %s\n", alloc.Netmask, alloc.Address)

	return connect.NewResponse(response), nil
}
