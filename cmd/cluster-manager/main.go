package main

import (
	"net/http"
	"wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
	"wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"
	"wireguard-cni/pkg/server"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	_ ipamv1connect.IPAMServiceHandler           = &server.Server{}
	_ wireguardv1connect.WireguardServiceHandler = &server.Server{}
)

func main() {
	s := server.NewIPAMServer()
	mux := http.NewServeMux()

	path, handler := ipamv1connect.NewIPAMServiceHandler(s)
	mux.Handle(path, handler)
	path, handler = wireguardv1connect.NewWireguardServiceHandler(s)
	mux.Handle(path, handler)
	http.ListenAndServe("localhost:8080", h2c.NewHandler(mux, &http2.Server{}))
}
