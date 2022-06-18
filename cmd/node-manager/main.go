package main

import (
	"expvar"
	"flag"
	"log"
	"net/http"
	"wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
	"wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"

	"github.com/hashicorp/go-cleanhttp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	log.Println("initializing node-manager")
	cfg := config()

	ipamClient := ipamv1connect.NewIPAMServiceClient(cleanhttp.DefaultClient(), cfg.ClusterManagerAddr)

	log.Println("initializing server")
	log.Println("initializing client ipam cidr")
	svr, err := NewNodeManagerServer(ipamClient)
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
	log.Println("listening localhost:8080 ...")
	log.Fatal(http.ListenAndServe("localhost:8080", h2c.NewHandler(mux, &http2.Server{})))

}

type NodeConfig struct {
	ClusterManagerAddr string
}

func config() NodeConfig {
	clusterMgrAddr := flag.String("cluster-manager-addr", "localhost:8080", "CNI Cluster Manager address")
	flag.Parse()

	return NodeConfig{
		ClusterManagerAddr: *clusterMgrAddr,
	}
}
