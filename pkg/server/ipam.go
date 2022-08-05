package server

import (
	"context"
	"expvar"
	"log"
	"net/http"
	"sync"

	"github.com/bufbuild/connect-go"
	ipamv1 "github.com/clly/wireguard-cni/gen/wgcni/ipam/v1"
	"github.com/clly/wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
	wireguardv1 "github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1"
	"github.com/hashicorp/go-uuid"
	goipam "github.com/metal-stack/go-ipam"
)

var _ ipamv1connect.IPAMServiceHandler = new(Server)

var (
	wireguardExpvar = new(expvar.Map).Init()
	once            = new(sync.Once)
)

type ModeIPAM int

const (
	ClusterMode ModeIPAM = iota
	NodeMode
)

func init() {
	expvar.Publish("wireguard", wireguardExpvar)
}

func (s *Server) IPAMServiceHandler() (string, http.Handler) {
	return ipamv1connect.NewIPAMServiceHandler(s)
}

func NewServer(cidr string, ipamMode ModeIPAM, self *wireguardv1.Peer) (*Server, error) {
	wireguardExpvar.Init()

	ipam := goipam.New()

	prefix, err := ipam.NewPrefix(cidr)
	if err != nil {
		return nil, err
	}

	once.Do(func() {
		expvar.Publish("ipam-usage", expvar.Func(ipamUsage(ipam, prefix.Cidr)))
	})

	return &Server{
		wgKey:     newMapDB(),
		expvarMap: wireguardExpvar,
		prefix:    prefix,
		mode:      ipamMode,
		ipam:      ipam,
		self:      self,
	}, nil
}

func ipamUsage(i goipam.Ipamer, cidrPrefix string) func() any {
	return func() any {
		return i.PrefixFrom(cidrPrefix).Usage()
	}
}

func newUUID() string {
	id, err := uuid.GenerateUUID()
	if err != nil {
		panic(err)
	}
	return id
}

func (s *Server) Alloc(
	ctx context.Context,
	req *connect.Request[ipamv1.AllocRequest],
) (*connect.Response[ipamv1.AllocResponse], error) {

	id := newUUID() // todo, record this

	alloc := &ipamv1.IPAlloc{
		Netmask: "24",
		Version: ipamv1.IPVersion_IP_VERSION_V4,
		Id:      id,
	}

	switch s.mode {
	case ClusterMode:
		prefix, err := s.ipam.AcquireChildPrefix(s.prefix.Cidr, 24)
		if err != nil {
			return nil, err
		}
		ip, err := prefix.Network()
		if err != nil {
			return nil, err
		}
		alloc.Address = ip.String()
	case NodeMode:
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

	log.Printf("Allocated new /%s CIDR: %s, ID: %s\n", alloc.Netmask, alloc.Address, alloc.Id)
	return connect.NewResponse(response), nil
}

func (s *Server) DeAlloc(ctx context.Context, req *connect.Request[ipamv1.DeAllocRequest]) (*connect.Response[ipamv1.DeAllocResponse], error) {
	id := req.Msg.Id
	log.Printf("DeAlloc id: %s", id)
	response := new(ipamv1.DeAllocResponse)
	return connect.NewResponse(response), nil
}
