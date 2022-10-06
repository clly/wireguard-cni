package wireguard

import "expvar"

var (
	expWireguardPublicKey = &expvar.String{}
	expDeviceFunc         expvar.Func
)

func init() {
	expvar.Publish("wireguard-public-key", expWireguardPublicKey)
	expvar.Publish("wireguard-device-wg0", expDeviceFunc)
}
