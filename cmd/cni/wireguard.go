package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/bufbuild/connect-go"
	ipamv1 "github.com/clly/wireguard-cni/gen/wgcni/ipam/v1"
	"github.com/clly/wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
	"github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"
	"github.com/clly/wireguard-cni/pkg/wireguard"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/hashicorp/go-cleanhttp"
)

// type wgResult struct {}

func addWgInterface(ctx context.Context, cfg PluginConf, netnsContainer string, result *current.Result, netns ns.NetNS) (string, error) {

	f, err := ioutil.TempFile("/tmp", "wireguard")
	if err != nil {
		f.Close()
	}

	var device string
	err = netns.Do(func(nn ns.NetNS) error {
		ip := result.IPs[0].Address.IP.String()

		ipamClient := ipamv1connect.NewIPAMServiceClient(cleanhttp.DefaultClient(), cfg.NodeManagerAddr)
		wireguardClient := wireguardv1connect.NewWireguardServiceClient(cleanhttp.DefaultClient(), cfg.NodeManagerAddr)

		resp, err := ipamClient.Alloc(ctx, connect.NewRequest(&ipamv1.AllocRequest{}))
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to request alloc")
			return err
		}

		addr := net.JoinHostPort(ip, "51820")
		// log.Println("Using", ip, "as wireguard endpoint")
		// log.Println("Using", addr, "as wireguard interface address")

		fmt.Fprintf(os.Stderr, "Namespace Path: %#v\n", nn.Path())

		wgAddr := resp.Msg.Alloc.Address

		cidr := fmt.Sprintf("%s/%s", resp.Msg.Alloc.Address, resp.Msg.Alloc.Netmask)

		wgConf := wireguard.Config{
			Address:   wgAddr,
			Endpoint:  addr,
			Route:     cidr,
			Namespace: nn.Path(),
		}

		netns.Path()

		fmt.Fprintln(os.Stderr, wgConf)
		wgMgr, err := wireguard.New(ctx, wgConf, wireguardClient, wireguard.WithOutput(os.Stderr))
		if err != nil {
			log.Println("failed to create wireguard manager")
			return err
		}

		device = fmt.Sprintf("wg%s", randomString())
		readCl, err := openConfig(device)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to open configuration file for interface", device)
			return err
		}
		if err = wgMgr.Config(readCl); err != nil {
			fmt.Fprintln(os.Stderr, "failed to write config")
			return err
		}

		fmt.Fprintln(os.Stderr, "Bringing up device", device)
		return wgMgr.Up(device)
	})

	return device, err
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

func getResult(device string, sandbox string) (current.IPConfig, current.Interface, error) {
	intf, err := net.InterfaceByName(device)
	if err != nil {
		return current.IPConfig{}, current.Interface{}, fmt.Errorf("could not retrieve interface by name: %s %w", device, err)
	}
	currentInterface := current.Interface{
		Name:    device,
		Mac:     intf.HardwareAddr.String(),
		Sandbox: sandbox,
	}
	addrs, err := intf.Addrs()
	if err != nil {
		return current.IPConfig{}, current.Interface{}, fmt.Errorf("failed to get addrs %w", err)
	}
	_, ipNet, err := net.ParseCIDR(addrs[0].String())
	if err != nil {
		return current.IPConfig{}, current.Interface{}, fmt.Errorf("failed to parse cidr %w", err)
	}

	ipCfg := current.IPConfig{
		Interface: nil,
		Address:   *ipNet,
		Gateway:   ipNet.IP,
	}
	return ipCfg, currentInterface, nil
}
