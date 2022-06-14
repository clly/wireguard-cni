package main

import (
	"expvar"
	"log"
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
	log.Println("initializing cluster-manager")
	log.Println("initializing server")
	s := server.NewServer()
	log.Println("initializing serve mux")
	mux := http.NewServeMux()

	path, handler := ipamv1connect.NewIPAMServiceHandler(s)
	log.Println("Registering IPAM Handler on", path)
	mux.Handle(path, handler)

	path, handler = wireguardv1connect.NewWireguardServiceHandler(s)
	log.Println("Registering Wireguard Handler on ", path)
	mux.Handle(path, handler)

	mux.Handle("/debug/varz", expvar.Handler())
	log.Println("listening localhost:8080 ...")
	http.ListenAndServe("localhost:8080", h2c.NewHandler(mux, &http2.Server{}))
}