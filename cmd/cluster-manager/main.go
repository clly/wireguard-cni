package main

import (
	"expvar"
	"flag"
	"log"
	"net/http"

	"github.com/clly/wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
	"github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"
	"github.com/clly/wireguard-cni/pkg/server"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	_ ipamv1connect.IPAMServiceHandler           = &server.Server{}
	_ wireguardv1connect.WireguardServiceHandler = &server.Server{}
)

type ClusterManagerConfig struct {
	ipamMode server.ModeIPAM
	prefix   string
}

func main() {
	log.Println("initializing cluster-manager")
	c := config()
	log.Println("initializing server")
	s, err := server.NewServer(c.prefix, c.ipamMode, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("initializing serve mux")
	mux := http.NewServeMux()

	path, handler := ipamv1connect.NewIPAMServiceHandler(s)
	log.Println("Registering IPAM Handler on", path)
	mux.Handle(path, handler)

	path, handler = wireguardv1connect.NewWireguardServiceHandler(s)
	log.Println("Registering Wireguard Handler on ", path)
	mux.Handle(path, handler)

	mux.Handle("/debug/varz", expvar.Handler())
	log.Println("listening :8080 ...")
	log.Fatal(http.ListenAndServe(":8080", h2c.NewHandler(mux, &http2.Server{})))
}

func config() ClusterManagerConfig {
	cidrPrefix := flag.String("cidr-prefix", "10.0.0.0/8", "Ipam CIDR prefix")
	flag.Parse()
	return ClusterManagerConfig{
		ipamMode: server.ClusterMode,
		prefix:   *cidrPrefix,
	}
}
