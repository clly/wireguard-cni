package main

import (
	"context"
	"expvar"
	"flag"
	"fmt"
	"log"
	"net/http"
	"wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
	"wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"
	"wireguard-cni/pkg/wireguard"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-sockaddr"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	ctx := context.Background()
	log.Println("initializing node-manager")
	cfg := config()

	ipamClient := ipamv1connect.NewIPAMServiceClient(cleanhttp.DefaultClient(), cfg.ClusterManagerAddr)
	wireguardClient := wireguardv1connect.NewWireguardServiceClient(cleanhttp.DefaultClient(), cfg.ClusterManagerAddr)

	log.Println("initializing server")
	log.Println("initializing client ipam cidr")
	svr, err := NewNodeManagerServer(ctx, cfg, ipamClient, wireguardClient)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("initializing serve mux")
	mux := http.NewServeMux()

	path, handler := ipamv1connect.NewIPAMServiceHandler(svr)
	log.Println("Registering IPAM Handler on", path)
	mux.Handle(path, handler)

	path, handler = wireguardv1connect.NewWireguardServiceHandler(svr)
	log.Println("Registering Wireguard Handler on ", path)
	mux.Handle(path, handler)

	mux.Handle("/debug/varz", expvar.Handler())
	log.Println("listening localhost:5242 ...")
	log.Fatal(http.ListenAndServe("localhost:5242", h2c.NewHandler(mux, &http2.Server{})))

}

type NodeConfig struct {
	ClusterManagerAddr string
	Wireguard          wireguard.Config
}

func config() NodeConfig {
	ip, err := sockaddr.GetPrivateIP()
	ip = fmt.Sprintf("%s:51820", ip)
	if err != nil {
		log.Println("failed to discover default address")
		ip = ""
	}
	clusterMgrAddr := flag.String("cluster-manager-url", "http://localhost:8080", "CNI Cluster Manager address")
	wireguardEndpoint := flag.String("wireguard-endpoint", ip, "endpoint:port for the wireguard socket")

	flag.Parse()

	return NodeConfig{
		ClusterManagerAddr: *clusterMgrAddr,
		Wireguard: wireguard.Config{
			Endpoint: *wireguardEndpoint,
			Route:    "",
		},
	}
}
