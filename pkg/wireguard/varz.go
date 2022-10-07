package wireguard

import "expvar"

var (
	expWireguardPublicKey = &expvar.String{}
)

func init() {
	expvar.Publish("wireguard-public-key", expWireguardPublicKey)
}
