package wireguard

import (
	"context"
	"log"

	wireguardv1 "wireguard-cni/gen/wgcni/wireguard/v1"
	"wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"

	"github.com/bufbuild/connect-go"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Config struct {
	Endpoint string
	Route    string
}

// deviceName string,
func New(ctx context.Context, cfg Config, client wireguardv1connect.WireguardServiceClient) error {
	log.Println("generating public keys")
	key, err := generateKeys()
	if err != nil {
		return err
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

	return err
}

func generateKeys() (wgtypes.Key, error) {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return wgtypes.Key{}, err
	}
	return key, nil
}
