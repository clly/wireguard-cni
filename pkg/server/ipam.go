package server

import (
	"context"
	"expvar"
	"log"
	"net/http"
	ipamv1 "wireguard-cni/gen/wgcni/ipam/v1"
	"wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"

	"github.com/bufbuild/connect-go"
)

var (
	_ ipamv1connect.IPAMServiceHandler = &Server{}
)

func (s *Server) IPAMServiceHandler() (string, http.Handler) {
	return ipamv1connect.NewIPAMServiceHandler(s)
}

func NewServer() *Server {
	m := new(expvar.Map).Init()
	expvar.Publish("wireguard", m)
	return &Server{
		wgKey:     newMapDB(),
		expvarMap: m,
	}
}

func (s *Server) Alloc(
	ctx context.Context,
	req *connect.Request[ipamv1.AllocRequest],
) (*connect.Response[ipamv1.AllocResponse], error) {
	log.Println("Headers", req.Header())
	response := &ipamv1.AllocResponse{
		Alloc: &ipamv1.IPAlloc{
			Address: "10.0.0.0",
			Netmask: "24",
			Version: ipamv1.IPVersion_IP_VERSION_V4,
		},
	}

	return connect.NewResponse(response), nil
}