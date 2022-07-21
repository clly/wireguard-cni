package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"
	ipamv1 "wireguard-cni/gen/wgcni/ipam/v1"
	"wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
	"wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"
	"wireguard-cni/pkg/wireguard"

	current "github.com/containernetworking/cni/pkg/types/100"

	"github.com/bufbuild/connect-go"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/hashicorp/go-cleanhttp"
)

// type wgResult struct {}

func addWgInterface(ctx context.Context, cfg PluginConf, netnsContainer string, result *current.Result, netns ns.NetNS) error {

	return netns.Do(func(nn ns.NetNS) error {
		ip := result.IPs[0].Address.IP.String()

		ipamClient := ipamv1connect.NewIPAMServiceClient(cleanhttp.DefaultClient(), cfg.NodeManagerAddr)
		wireguardClient := wireguardv1connect.NewWireguardServiceClient(cleanhttp.DefaultClient(), cfg.NodeManagerAddr)

		resp, err := ipamClient.Alloc(ctx, connect.NewRequest(&ipamv1.AllocRequest{}))
		if err != nil {
			return err
		}

		addr := net.JoinHostPort(ip, "51820")
		log.Println("Using", ip, "as wireguard endpoint")
		log.Println("Using", addr, "as wireguard interface address")

		fmt.Fprintf(os.Stderr, "%#v\n", nn.Path())

		wgAddr := resp.Msg.Alloc.Address

		cidr := fmt.Sprintf("%s/%s", resp.Msg.Alloc.Address, resp.Msg.Alloc.Netmask)

		wgConf := wireguard.Config{
			Address:   wgAddr,
			Endpoint:  addr,
			Route:     cidr,
			Namespace: netnsContainer,
		}

		fmt.Fprintln(os.Stderr, wgConf)
		wgMgr, err := wireguard.New(ctx, wgConf, wireguardClient)
		if err != nil {
			return err
		}

		device := fmt.Sprintf("wg%s", randomString())
		readCl, err := openConfig(device)
		if err != nil {
			return err
		}
		if err = wgMgr.Config(readCl); err != nil {
			return err
		}

		return wgMgr.Up(device)
	})
}

func randomString() string {
	b := make([]byte, 4)

	rand.New(rand.NewSource(time.Now().UnixNano())).Read(b)
	return hex.EncodeToString(b)
}

func openConfig(device string) (io.WriteCloser, error) {
	filename := fmt.Sprintf("%s.conf", device)
	f := filepath.Join("/etc", "wireguard", filename)
	return os.OpenFile(f, os.O_CREATE|os.O_TRUNC|os.O_SYNC|os.O_RDWR, 0644)
}
