package server

import (
	"context"
	"expvar"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/bufbuild/connect-go"
	goipam "github.com/metal-stack/go-ipam"

	ipamv1 "github.com/clly/wireguard-cni/gen/wgcni/ipam/v1"
	"github.com/clly/wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
)

var (
	_               ipamv1connect.IPAMServiceHandler = &Server{}
	wireguardExpvar                                  = new(expvar.Map).Init()
	once                                             = &sync.Once{}
)

// clusterIpam is a wrapper around github.com/metal-stack/go-clusterIpam.Ipamer to make persisting and loading the clusterIpam state easier
type clusterIpam struct {
	goipam.Ipamer
	persistFile string
	prefix      *goipam.Prefix
}

const IpamDataFile = "ipam.json"

func newIPAM(ctx context.Context, dataDir, cidr string) (*clusterIpam, error) {
	ipamer := goipam.New()

	prefix, err := ipamer.NewPrefix(ctx, cidr)
	if err != nil {
		return nil, err
	}

	var persistFile string
	if dataDir != "" {
		persistFile = filepath.Join(dataDir, IpamDataFile)
	}
	ipam := &clusterIpam{
		persistFile: persistFile,
		Ipamer:      ipamer,
		prefix:      prefix,
	}
	if err := ipam.loadData(ctx); err != nil {
		return nil, err
	}
	return ipam, nil
}

// save will dump ipam state from memory into a file
func (i *clusterIpam) save(ctx context.Context) error {
	if i.persistFile == "" {
		return nil
	}
	data, err := i.Ipamer.Dump(ctx)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(i.persistFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_SYNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(data)
	return err
}

// loadData will load ipam state from the data directory
func (i *clusterIpam) loadData(ctx context.Context) error {
	if _, err := os.Stat(i.persistFile); os.IsNotExist(err) {
		return nil
	}

	err := i.deleteAllPrefixes(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete prefixes for loading %w", err)
	}
	b, err := ioutil.ReadFile(i.persistFile)
	if err != nil {
		return err
	}
	return i.Ipamer.Load(ctx, string(b))
}

func (i *clusterIpam) deleteAllPrefixes(ctx context.Context) error {
	prefixes, err := i.Ipamer.ReadAllPrefixCidrs(ctx)
	if err != nil {
		return fmt.Errorf("failed to read prefixes for deletion %w", err)
	}
	for _, prefix := range prefixes {
		_, err := i.Ipamer.DeletePrefix(ctx, prefix)
		if err != nil {
			return err
		}
	}
	return nil
}

type IPAM_MODE int

const (
	CLUSTER_MODE IPAM_MODE = iota
	NODE_MODE
)

func init() {
	expvar.Publish("wireguard", wireguardExpvar)
}

func (s *Server) IPAMServiceHandler() (string, http.Handler) {
	return ipamv1connect.NewIPAMServiceHandler(s)
}

func ipamUsage(i goipam.Ipamer, cidrPrefix string) func() any {
	return func() any {
		return i.PrefixFrom(context.TODO(), cidrPrefix).Usage()
	}
}

func (s *Server) Alloc(
	ctx context.Context,
	req *connect.Request[ipamv1.AllocRequest],
) (*connect.Response[ipamv1.AllocResponse], error) {

	alloc := &ipamv1.IPAlloc{
		Netmask: "24",
		Version: ipamv1.IPVersion_IP_VERSION_V4,
	}

	switch s.mode {
	case CLUSTER_MODE:
		prefix, err := s.ipam.AcquireChildPrefix(ctx, s.ipam.prefix.Cidr, 24)
		if err != nil {
			return nil, err
		}
		ip, err := prefix.Network()
		if err != nil {
			return nil, err
		}
		alloc.Address = ip.String()
	case NODE_MODE:
		alloc.Netmask = "32"
		ip, err := s.ipam.AcquireIP(ctx, s.ipam.prefix.Cidr)
		if err != nil {
			return nil, err
		}
		alloc.Address = ip.IP.String()
	}
	response := &ipamv1.AllocResponse{
		Alloc: alloc,
	}

	if err := s.ipam.save(ctx); err != nil {
		return nil, err
	}
	log.Printf("Allocated new /%s CIDR %s\n", alloc.Netmask, alloc.Address)

	return connect.NewResponse(response), nil
}
