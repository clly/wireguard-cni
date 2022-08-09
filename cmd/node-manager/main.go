package main

import (
	"context"
	"encoding/json"
	"expvar"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/clly/wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
	"github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"
	"github.com/clly/wireguard-cni/pkg/wireguard"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-sockaddr"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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

	ipamClient := ipamv1connect.NewIPAMServiceClient(cleanhttp.DefaultClient(), cfg.ClusterManagerAddr)
	wireguardClient := wireguardv1connect.NewWireguardServiceClient(cleanhttp.DefaultClient(), cfg.ClusterManagerAddr)

	log.Println("initializing server")
	log.Println("initializing client ipam cidr")
	svr, err := NewNodeManagerServer(ctx, cfg, ipamClient, wireguardClient)
	if err != nil {
		log.Fatal(err)
	}

	quit(svr, cfg.InterfaceName)

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
	InterfaceName      string
	ConfigDirectory    string
	ListenAddr         string
	Wireguard          wireguard.Config
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
	interfaceName := flag.String("wireguard-interface", "wg0", "wireguard interface name")
	configDirectory := flag.String("wireguard-config-directory", "/etc/wireguard", "Wireguard configuration directory")
	listenAddr := flag.String("addr", "localhost:5242", "node manager listen address")

	flag.Parse()

	clusterMgr := os.ExpandEnv(valOrEnv(*clusterMgrAddr, clusterMgrEnvKey, clusterMgrDefault))

	return NodeConfig{
		ClusterManagerAddr: clusterMgr,
		ConfigDirectory:    *configDirectory,
		InterfaceName:      *interfaceName,
		ListenAddr:         *listenAddr,
		Wireguard: wireguard.Config{
			Endpoint: *wireguardEndpoint,
		},
	}
}

func valOrEnv(v, env, defaultVal string) string {
	if v != "" {
		return v
	}
	if e := os.Getenv(env); e != "" {
		return e
	}
	return defaultVal
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
