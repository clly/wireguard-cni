package wireguard

import (
	"context"
	"io"
	"log"

	wireguardv1 "wireguard-cni/gen/wgcni/wireguard/v1"
	"wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"

	"github.com/bufbuild/connect-go"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Config struct {
	Endpoint  string
	Route     string
	Namespace string
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
}

// WGQuickManager implements WireguardManager using shell scripts and wg-quick
type WGQuickManager struct {
	client    wireguardv1connect.WireguardServiceClient
	key       wgtypes.Key
	endpoint  string
	namespace string
}

func New(ctx context.Context, cfg Config, client wireguardv1connect.WireguardServiceClient) (WireguardManager, error) {
	log.Println("generating public keys")
	key, err := generateKeys()
	if err != nil {
		return nil, err
	}

	expWireguardPublicKey.Set(key.PublicKey().String())

	req := &wireguardv1.RegisterRequest{
		PublicKey: key.PublicKey().String(),
		Endpoint:  cfg.Endpoint,
		Route:     cfg.Route,
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
	}
	return mgr, err
}

func generateKeys() (wgtypes.Key, error) {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return wgtypes.Key{}, err
	}
	return key, nil
}
