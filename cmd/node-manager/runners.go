package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"wireguard-cni/pkg/wireguard"

	"oss.indeed.com/go/libtime"
)

// peerMgr will set wireguard configuration file and periodically call SetPeers which will call wg-quick down &&
// wg-quick up
func peerMgr(ctx context.Context, mgr wireguard.WireguardManager, cfgFile string) error {
	timer, cancel := libtime.SafeTimer(1 * time.Second)
	defer cancel()

	for {
		select {
		case <-timer.C:
			if err := setConfig(mgr, cfgFile); err != nil {
				log.Println(err)
				return err
			}
		case <-ctx.Done():
			log.Println("cancelling peer manager")
			return ctx.Err()
		}
		// call set peers
	}
}

func setConfig(mgr wireguard.WireguardManager, cfgFile string) error {
	f, err := os.OpenFile(cfgFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println("failed to open file", cfgFile)
		return err
	}
	err = mgr.Config(f)
	if err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	err = f.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync to disk: %w", err)
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}
