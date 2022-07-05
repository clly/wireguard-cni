package main

import (
	"log"
	"os"
	"time"
	"wireguard-cni/pkg/wireguard"
)

// peerMgr will set wireguard configuration file and periodically call SetPeers which will call wg-quick down &&
// wg-quick up
func peerMgr(mgr wireguard.WireguardManager, cfgFile string) error {
	ticker := time.NewTicker(1 * time.Second)

	for {
		<-ticker.C
		if err := setConfig(mgr, cfgFile); err != nil {
			return err
		}
		// call set peers
	}
}

func setConfig(mgr wireguard.WireguardManager, cfgFile string) error {
	f, err := os.OpenFile(cfgFile, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		log.Println("failed to open file", cfgFile)
		return err
	}
	err = mgr.Config(f)
	if err != nil {
		log.Println("failed to write config")
		return err
	}
	err = f.Sync()
	if err != nil {
		log.Println("failed to sync to disk")
		return err
	}
	err = f.Close()
	if err != nil {
		log.Println("failed to close file")
	}
	return nil
}
