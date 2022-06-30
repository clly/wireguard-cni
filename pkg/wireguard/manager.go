package wireguard

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"strings"
	"text/template"
	wireguardv1 "wireguard-cni/gen/wgcni/wireguard/v1"

	"github.com/bufbuild/connect-go"
)

func (w *WGQuickManager) Up(device string) error {
	return run("wg-quick", "up", device)
}

func (w *WGQuickManager) Down(device string) error {
	return run("wg-quick", "down", device)
}

func shell(cmd string, args ...string) (string, error) {
	c := exec.Command(cmd, args...)
	b, err := c.CombinedOutput()
	return string(b), err
}

func run(cmd string, args ...string) error {
	output, err := shell(cmd, args...)
	log.Println("run", fmt.Sprintf("[%s %s]", cmd, strings.Join(args, " ")))
	if len(output) > 0 {
		fmt.Println(output)
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

	host, port, err := net.SplitHostPort(w.endpoint)
	if err != nil {
		return err
	}

	cfg := wgConfig{
		Address:    host,
		PrivateKey: w.key.String(),
		Port:       port,
		PostUp:     nil,
		PostDown:   nil,
		Peers:      fromPeerSlice(peers.Msg.GetPeers(), w.self()),
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
