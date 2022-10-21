package main

import (
	"context"
	"expvar"
	"flag"
	"log"
	"net/http"
	"os"
	"path"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/clly/wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
	"github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"
	"github.com/clly/wireguard-cni/pkg/ipam"
	"github.com/clly/wireguard-cni/pkg/server"
)

var (
	_ ipamv1connect.IPAMServiceHandler           = &server.Server{}
	_ wireguardv1connect.WireguardServiceHandler = &server.Server{}
)

type ClusterManagerConfig struct {
	prefix        string
	dataDirectory string
}

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	log.Println("initializing cluster-manager")
	c := config()
	log.Println("initializing server")

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if c.dataDirectory == "" {
		c.dataDirectory = path.Join(wd, "cluster-manager")
	}

	clusterIpam, err := ipam.New(context.TODO(), c.dataDirectory, c.prefix)
	if err != nil {
		log.Fatal(err)
	}

	s, err := server.NewServer(clusterIpam, server.WithDataDir(c.dataDirectory))

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
	cidrPrefix := flag.String("cidr-prefix", "10.0.0.0/8", "IPAM CIDR prefix")
	dataDir := flag.String("data-dir", "", "Data directory to store ipam and wireguard files in")
	flag.Parse()
	return ClusterManagerConfig{
		prefix:        *cidrPrefix,
		dataDirectory: *dataDir,
	}
}
