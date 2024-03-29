package wireguard

import (
	"context"
	"expvar"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"connectrpc.com/connect"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	wireguardv1 "github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1"
	"github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"
)

type Config struct {
	// Endpoint is the address that other wireguard nodes should dial
	Endpoint string
	// Address is the ip address that should be set on the wireguard interface
	Address   string
	Route     string
	Namespace string
	// device is the device name for the wireguard interface
	Device string
}

// WireguardManager creates and deletes Wireguard interfaces, generates wireguard configuration and can update peers on
// a wireguard interface
type WireguardManager interface {
	// Config should write the wireguard configuration file to the provided writer
	Config(w io.Writer) error
	// Up should bring the provided device up
	Up(device string) error
	// Down should print the provided device down
	Down(device string) error
	// SetPeers should add the provided peers to the provided device. It may manage routes.
	SetPeers(device string, peers []*Peer) error
	// Device will return the underlying wireguard device configuration
	Device() *wgtypes.Device
}

// WGQuickManager implements WireguardManager using shell scripts and wg-quick
type WGQuickManager struct {
	client wireguardv1connect.WireguardServiceClient
	key    wgtypes.Key
	// endpoint is the address that other wireguard servers should dial
	endpoint string
	// addr is the address that should be set on the wireguard interface
	addr         string
	namespace    string
	peerRegistry Peers

	logOutput io.Writer
	postup    *string
	postdown  *string

	wgclient WGClient
	device   *wgtypes.Device

	stopCh chan struct{}
}

// wireguard client interface
type WGClient interface {
	io.Closer
	Devices() ([]*wgtypes.Device, error)
	Device(name string) (*wgtypes.Device, error)
	ConfigureDevice(name string, cfg wgtypes.Config) error
}

func (w *WGQuickManager) SetPeerRegistry(p Peers) {
	w.peerRegistry = p
}

func (w *WGQuickManager) SetAddress(addr string) {
	w.addr = addr
}

func (w *WGQuickManager) Device() *wgtypes.Device {
	return w.device
}

type Peers interface {
	ListPeers() ([]*wireguardv1.Peer, error)
}

func (w *WGQuickManager) Self() Peer {
	return Peer{
		Endpoint:   w.endpoint,
		PublicKey:  w.key.PublicKey().String(),
		AllowedIPs: "0.0.0.0/0",
	}
}

func (w *WGQuickManager) expvar() any {
	if w.device == nil {
		return map[string]any{}
	}

	return map[string]any{
		"name":          w.device.Name,
		"listen-port":   w.device.ListenPort,
		"firewall-mark": w.device.FirewallMark,
		"type":          w.device.Type.String(),
	}
}

var once = &sync.Once{}

func New(ctx context.Context, cfg Config, wgClient WGClient, client wireguardv1connect.WireguardServiceClient, opts ...WGOption) (*WGQuickManager, error) {
	if cfg.Device == "" {
		cfg.Device = "wg0"
	}

	d, err := deviceIfExists(wgClient, cfg.Device)
	if err != nil {
		return nil, err
	}

	var key wgtypes.Key
	if d != nil {
		key = d.PrivateKey
	} else {
		log.Println("generating public keys")
		key, err = generateKeys()
		if err != nil {
			return nil, err
		}
	}

	expWireguardPublicKey.Set(key.PublicKey().String())

	req := &wireguardv1.RegisterRequest{
		PublicKey: key.PublicKey().String(),
		Endpoint:  cfg.Endpoint,
		Route:     cfg.Route,
		Namespace: cfg.Namespace,
	}
	log.Println("registering with public key with upstream", "endpoint", cfg.Endpoint, "route", cfg.Route)
	_, err = client.Register(ctx, connect.NewRequest(req))
	if err != nil {
		log.Println("failed to register with upstream", err)
	}

	mgr := &WGQuickManager{
		client:    client,
		key:       key,
		endpoint:  cfg.Endpoint,
		namespace: cfg.Namespace,
		addr:      cfg.Address,
		logOutput: os.Stdout,
		device:    d,
		wgclient:  wgClient,
		stopCh:    make(chan struct{}),
	}

	for _, opt := range opts {
		opt(mgr)
	}

	once.Do(func() {
		expvar.Publish("wireguard-device-wg0", expvar.Func(mgr.expvar))
	})

	go func(wm *WGQuickManager, deviceName string) {
		t := time.NewTicker(30 * time.Second)
		log.Println("Starting device sync")
		// sleep for a second to see if we bring up the device quick enough. run it first then periodically
		time.Sleep(time.Second)
		d, err := deviceIfExists(wm.wgclient, deviceName)
		if err != nil {
			log.Println("failed to read device", err)
		}
		if d != nil {
			wm.device = d
		}
		for {
			select {
			case <-t.C:
				log.Println("Syncing device", deviceName)
				d, err := deviceIfExists(wm.wgclient, deviceName)
				if err != nil {
					log.Println("failed to read device", err)
				}
				if d != nil {
					wm.device = d
				}
			case <-wm.stopCh:
				t.Stop()
				return
			}

		}
	}(mgr, cfg.Device)

	return mgr, err
}

func generateKeys() (wgtypes.Key, error) {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return wgtypes.Key{}, err
	}
	return key, nil
}

func deviceIfExists(c WGClient, name string) (*wgtypes.Device, error) {
	d, err := c.Device(name)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if os.IsNotExist(err) {
		return nil, nil
	}
	return d, nil
}
