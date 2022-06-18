package server

import (
	"context"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"sync"
	ipamv1 "wireguard-cni/gen/wgcni/ipam/v1"
	"wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"

	"github.com/bufbuild/connect-go"
	goipam "github.com/metal-stack/go-ipam"
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

func NewServer(cidr string, ipamMode IPAM_MODE) (*Server, error) {
	wireguardExpvar.Init()

	ipam := goipam.New()

	prefix, err := ipam.NewPrefix(cidr)
	if err != nil {
		return nil, err
	}

	once.Do(func() {
		expvar.Publish("foo", expvar.Func(ipamUsage(prefix)))
	})

	return &Server{
		wgKey:     newMapDB(),
		expvarMap: wireguardExpvar,
		prefix:    prefix,
		mode:      ipamMode,
		ipam:      ipam,
	}, nil
}

func ipamUsage(p *goipam.Prefix) func() any {
	var prefix = p
	return func() any {
		return prefix.Usage()
	}
}

func (s *Server) Alloc(
	ctx context.Context,
	req *connect.Request[ipamv1.AllocRequest],
) (*connect.Response[ipamv1.AllocResponse], error) {
	log.Println("Headers", req.Header())

	alloc := &ipamv1.IPAlloc{
		Netmask: "24",
		Version: ipamv1.IPVersion_IP_VERSION_V4,
	}

	switch s.mode {
	case CLUSTER_MODE:
		fmt.Println(s.prefix.Cidr)
		prefix, err := s.ipam.AcquireChildPrefix(s.prefix.Cidr, 24)
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
		ip, err := s.ipam.AcquireIP(s.prefix.Cidr)
		if err != nil {
			return nil, err
		}
		alloc.Address = ip.IP.String()
	}
	response := &ipamv1.AllocResponse{
		Alloc: alloc,
	}

	return connect.NewResponse(response), nil
}
