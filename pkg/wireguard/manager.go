package wireguard

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/template"

	"github.com/bufbuild/connect-go"
	wireguardv1 "github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1"
)

func (w *WGQuickManager) Up(device string) error {
	var cmd []string
	if w.namespace != "" {
		// nsenter --net=/run/docker/netns/f1cffea8d447
		cmd = append(cmd, "nsenter", fmt.Sprintf("--net=%s", w.namespace))
	}
	cmd = append(cmd, "wg-quick", "up", device)
	_, _ = fmt.Fprintln(os.Stderr, cmd)
	return run(w.logOutput, cmd[0], cmd[1:]...)
}

func (w *WGQuickManager) Down(device string) error {
	return run(w.logOutput, "wg-quick", "down", device)
}

// SetPeers will set peers by bringing the device down and up. The configuration file must be written before calling
// SetPeers.
func (w *WGQuickManager) SetPeers(device string, peers []*Peer) error {
	if err := w.Down(device); err != nil {
		return err
	}
	return w.Up(device)
}

func shell(cmd string, args ...string) (string, error) {
	c := exec.Command(cmd, args...)
	b, err := c.CombinedOutput()
	return string(b), err
}

func run(w io.Writer, cmd string, args ...string) error {
	output, err := shell(cmd, args...)
	log.Println("run", fmt.Sprintf("[%s %s]", cmd, strings.Join(args, " ")))
	if len(output) > 0 {
		_, _ = fmt.Fprintln(w, output)
	}
	return err
}

//go:embed tmpl/wireguard.conf.tmpl
var wgConfigTemplate string

type wgConfig struct {
	Address    string
	PrivateKey string
	Port       string
	PostUp     *string
	PostDown   *string
	Peers      []Peer
}

type Peer struct {
	Endpoint   string
	PublicKey  string
	AllowedIPs string
}

func (w *WGQuickManager) Config(writer io.Writer) error {
	t, err := template.New("wg-config").Parse(wgConfigTemplate)
	if err != nil {
		return err
	}

	// We need to feed the context in for tracing in the future

	// This could also be cached and updated in the background in the future or we can add streaming which would
	// probably be more efficient and everyone could get updated at the same time
	peers, err := w.client.Peers(context.Background(), connect.NewRequest(&wireguardv1.PeersRequest{}))
	if err != nil {
		return err
	}

	// get the peers that are connected to ourself. Maybe this should be feed differently but this seems easiest right now
	var selfPbPeers = []*wireguardv1.Peer{}
	if w.peerRegistry != nil {
		registryPeers, err := w.peerRegistry.ListPeers()
		if err != nil {
			return err
		}
		selfPbPeers = registryPeers
	}

	cfgPeer := fromPeerSlice(selfPbPeers, w.self())

	_, port, err := net.SplitHostPort(w.endpoint)
	if err != nil {
		return err
	}

	cfgPeer = append(cfgPeer, fromPeerSlice(peers.Msg.GetPeers(), w.self())...)

	sort.SliceStable(cfgPeer, func(i, j int) bool {
		return cfgPeer[i].AllowedIPs < cfgPeer[j].AllowedIPs
	})

	cfg := wgConfig{
		Address:    w.addr,
		PrivateKey: w.key.String(),
		Port:       port,
		PostUp:     w.postup,
		PostDown:   w.postdown,
		Peers:      cfgPeer,
	}

	return t.Execute(writer, cfg)
}

// we could preconstruct this to lower allocs
func (w *WGQuickManager) self() Peer {
	return Peer{
		Endpoint:  w.endpoint,
		PublicKey: w.key.PublicKey().String(),
	}
}

func fromPeerSlice(pbPeers []*wireguardv1.Peer, self Peer) []Peer {
	peers := make([]Peer, 0, len(pbPeers))
	for _, pbPeer := range pbPeers {
		peer := fromPeer(pbPeer)
		if peer.PublicKey == self.PublicKey {
			continue
		}
		peers = append(peers, peer)
	}
	return peers
}

func fromPeer(p *wireguardv1.Peer) Peer {
	return Peer{
		Endpoint:   p.GetEndpoint(),
		PublicKey:  p.GetPublicKey(),
		AllowedIPs: p.GetRoute(),
	}
}
