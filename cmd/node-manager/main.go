package main

import (
	"context"
	"encoding/json"
	"expvar"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/hashicorp/go-sockaddr"
	socktemplate "github.com/hashicorp/go-sockaddr/template"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/clly/wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
	"github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"
)

func main() {
	ctx := context.Background()
	log.Println("initializing node-manager")
	cfg := config()

	b, err := json.MarshalIndent(cfg, "==>", "  ")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(b))

	log.Println("initializing server")
	log.Println("initializing client ipam cidr")
	svr, err := NewNodeManagerServer(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	quit(svr, cfg.Wireguard.InterfaceName)

	log.Println("initializing serve mux")
	mux := http.NewServeMux()

	path, handler := ipamv1connect.NewIPAMServiceHandler(svr)
	log.Println("Registering IPAM Handler on", path)
	mux.Handle(path, handler)

	path, handler = wireguardv1connect.NewWireguardServiceHandler(svr)
	log.Println("Registering Wireguard Handler on ", path)
	mux.Handle(path, handler)

	mux.Handle("/debug/varz", expvar.Handler())
	log.Println("listening", cfg.ListenAddr, "...")
	log.Fatal(http.ListenAndServe("localhost:5242", h2c.NewHandler(mux, &http2.Server{})))

}

type NodeConfig struct {
	ClusterManagerAddr string
	ConfigDirectory    string
	ListenAddr         string
	DataDirectory      string
	Wireguard          WireguardNodeConfig
}

// WireguardNodeConfig contains all the information that gets fed into the wireguard manager at some point
type WireguardNodeConfig struct {
	Endpoint      string
	InterfaceName string
}

const clusterMgrEnvKey = "CLUSTER_MANAGER_ADDR"
const clusterMgrDefault = "http://localhost:8080"

func config() NodeConfig {
	ip, err := sockaddr.GetPrivateIP()
	addr := net.JoinHostPort(ip, "51820")
	if err != nil {
		log.Println("failed to discover default address")
		addr = ""
	}
	clusterMgrAddr := flag.String("cluster-manager-url", "", "CNI Cluster Manager address")
	wireguardEndpoint := flag.String("wireguard-endpoint", addr, "endpoint:port for the wireguard socket")
	wireguardParse := flag.String("wireguard-sockaddr-network", "", "use sockaddr to parse the subnet on the interface for the wireguard endpoint")
	interfaceName := flag.String("wireguard-interface", "wg0", "wireguard interface name")
	configDirectory := flag.String("wireguard-config-directory", "/etc/wireguard", "Wireguard configuration directory")
	listenAddr := flag.String("addr", "localhost:5242", "node manager listen address")
	dataDir := flag.String("data-dir", "", "Data directory to store ipam and wireguard files in")

	flag.Parse()
	addr = *wireguardEndpoint

	if *wireguardParse != "" {
		log.Println(*wireguardParse)
		socktmpl := fmt.Sprintf("{{ GetPrivateInterfaces|include \"network\" \"%s\" | attr \"address\" }}", *wireguardParse)
		log.Println("parsing wireguard endpoint using sockaddr", socktmpl)
		addr, err = socktemplate.Parse(socktmpl)
		if err != nil {
			log.Fatal("failed to parse sockaddr tempalte")
		}
		addr = net.JoinHostPort(addr, "51820")
	}

	// later we can use flag.Visit to see if the clusterMgrAddr was visited
	clusterMgr := os.ExpandEnv(first(*clusterMgrAddr, os.Getenv(clusterMgrEnvKey), clusterMgrDefault))

	return NodeConfig{
		ClusterManagerAddr: clusterMgr,
		ConfigDirectory:    *configDirectory,
		ListenAddr:         *listenAddr,
		DataDirectory:      *dataDir,
		Wireguard: WireguardNodeConfig{
			Endpoint:      addr,
			InterfaceName: *interfaceName,
		},
	}
}

func first(s ...string) string {
	for _, v := range s {
		fmt.Fprintln(os.Stderr, v)
		if v != "" {
			return v
		}
	}
	return ""
}

func quit(mgr *NodeManagerServer, device string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		log.Println("shutting down...")
		logOnErr(mgr.wgManager.Down(device))
		for _, cancel := range mgr.cancelers {
			cancel()
		}
		os.Exit(1)
	}()
}

func logOnErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
