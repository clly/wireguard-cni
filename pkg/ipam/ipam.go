package ipam

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	goipam "github.com/metal-stack/go-ipam"
)

var (
// once = &sync.Once{}
)

// ClusterIpam is a wrapper around github.com/metal-stack/go-ClusterIpam.Ipamer to make persisting and loading the ClusterIpam state easier
type ClusterIpam struct {
	goipam.Ipamer
	persistFile string
	Prefix      *goipam.Prefix
	// TODO abstract the mode away from the server and into here
}

const IpamDataFile = "ipam.json"

func New(ctx context.Context, dataDir, cidr string) (*ClusterIpam, error) {
	ipamer := goipam.New()

	prefix, err := ipamer.NewPrefix(ctx, cidr)
	if err != nil {
		return nil, err
	}

	var persistFile string
	if dataDir != "" {
		persistFile = filepath.Join(dataDir, IpamDataFile)
	}
	ipam := &ClusterIpam{
		persistFile: persistFile,
		Ipamer:      ipamer,
		Prefix:      prefix,
	}
	if err := ipam.loadData(ctx); err != nil {
		return nil, err
	}
	return ipam, nil
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
		return err
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
