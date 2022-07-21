package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"wireguard-cni/pkg/wireguard"
)

// peerMgr will set wireguard configuration file and periodically call SetPeers which will call wg-quick down &&
// wg-quick up
func peerMgr(ctx context.Context, mgr wireguard.WireguardManager, cfgFile string) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	cfgHash := []byte{}
	device := deviceFromConf(cfgFile)
	log.Println("Starting config sync...")

	for {
		select {
		case <-ticker.C:
			if err := setConfig(mgr, cfgFile); err != nil {
				log.Println(err)
				return err
			}
			sha, err := hashFile(cfgFile)
			if err != nil {
				return err
			}
			if bytes.Equal(sha, cfgHash) {
				// this should be debug
				continue
			}
			cfgHash = sha
			log.Println("New config processed. Setting peers")
			if err := mgr.SetPeers(device, nil); err != nil {
				log.Println("failed to set peers", err)
			}
		case <-ctx.Done():
			log.Println("cancelling peer manager")
			return ctx.Err()
		}
	}
}

func setConfig(mgr wireguard.WireguardManager, cfgFile string) error {
	f, err := os.OpenFile(cfgFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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

func hashFile(cfgFile string) ([]byte, error) {
	b, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		log.Println("failed to read config file")
		return nil, err
	}

	d := sha256.Sum256(b)
	return d[:], nil
}

func deviceFromConf(cfgFile string) string {
	fileName := filepath.Base(cfgFile)
	device := strings.TrimRight(fileName, filepath.Ext(fileName))
	return device
}
