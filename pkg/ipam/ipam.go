package ipam

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"connectrpc.com/connect"
	goipam "github.com/metal-stack/go-ipam"

	ipamv1 "github.com/clly/wireguard-cni/gen/wgcni/ipam/v1"
	"github.com/clly/wireguard-cni/gen/wgcni/ipam/v1/ipamv1connect"
)

var (
// once = &sync.Once{}
)

// ClusterIpam is a wrapper around github.com/metal-stack/go-ClusterIpam.Ipamer to make persisting and loading the ClusterIpam state easier
type ClusterIpam struct {
	goipam.Ipamer
	persistFile string
	Prefix      *goipam.Prefix
	// ipamClient  ipamv1connect.IPAMServiceClient
	// TODO abstract the mode away from the server and into here
}

const IpamDataFile = "ipam.json"

func New(ctx context.Context, dataDir, cidr string) (*ClusterIpam, error) {
	ipam, err := newIPAM(ctx, dataDir)
	if err != nil {
		return nil, err
	}

	ipam.Prefix = ipam.PrefixFrom(ctx, cidr)
	if ipam.Prefix == nil {
		ipam.Prefix, err = ipam.NewPrefix(ctx, cidr)
		if err != nil {
			return nil, err
		}
	}

	if err = ipam.Save(ctx); err != nil {
		return nil, err
	}

	return ipam, nil
}

func NewRemoteIPAM(ctx context.Context, dataDir string, ipamClient ipamv1connect.IPAMServiceClient) (*ClusterIpam, error) {
	ipam, err := newIPAM(ctx, dataDir)
	if err != nil {
		return nil, err
	}

	prefixes, err := ipam.ReadAllPrefixCidrs(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Println(prefixes)
	if len(prefixes) == 0 {
		alloc, err := ipamClient.Alloc(ctx, connect.NewRequest(&ipamv1.AllocRequest{}))
		if err != nil {
			return nil, err
		}
		cidr := fmt.Sprintf("%s/%s", alloc.Msg.GetAlloc().Address, alloc.Msg.GetAlloc().Netmask)
		prefix, err := ipam.Ipamer.NewPrefix(ctx, cidr)
		if err != nil {
			return nil, err
		}
		ipam.Prefix = prefix
	} else {
		ipam.Prefix = ipam.PrefixFrom(ctx, prefixes[0])
	}

	return ipam, nil
}

func newIPAM(ctx context.Context, dataDir string) (*ClusterIpam, error) {
	ipamer := goipam.New()

	var persistFile string
	if dataDir != "" {
		persistFile = filepath.Join(dataDir, IpamDataFile)
	}
	ipam := &ClusterIpam{
		persistFile: persistFile,
		Ipamer:      ipamer,
	}
	if err := ipam.loadData(ctx); err != nil {
		return nil, err
	}

	err := ipam.Save(ctx)

	return ipam, err
}

// save will dump ipam state from memory into a file
func (i *ClusterIpam) Save(ctx context.Context) error {
	if i.persistFile == "" {
		return nil
	}

	data, err := i.Ipamer.Dump(ctx)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(path.Dir(i.persistFile), 0755); err != nil {
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
func (i *ClusterIpam) loadData(ctx context.Context) error {
	if _, err := os.Stat(i.persistFile); os.IsNotExist(err) {
		return nil
	}

	err := i.deleteAllPrefixes(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete prefixes for loading %w", err)
	}

	b, err := ioutil.ReadFile(i.persistFile)
	if err != nil {
		return fmt.Errorf("failed to read file %w", err)
	}

	return i.Ipamer.Load(ctx, string(b))
}

func (i *ClusterIpam) deleteAllPrefixes(ctx context.Context) error {
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

// func ipamUsage(i goipam.Ipamer, cidrPrefix string) func() any {
// 	return func() any {
// 		return i.PrefixFrom(context.TODO(), cidrPrefix).Usage()
// 	}
// }
