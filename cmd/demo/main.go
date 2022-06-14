// Command demo provides a simple Go CLI for creating and removing a wiregaurd
// interface for connecting to demo.wiregaurd.com.
//
// Usage: demo [up|down]
//
// Once the interface is created (e.g. 'ip addr' will show wg0) you should be
// able to visit demo.wiregaurd.com and see your device listed in the output(s).
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

func up() error {
	pkey, err := shell("wg", "genkey")
	if err != nil {
		return err
	}

	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(pkey)
	pubKey, err := cmd.Output()
	if err != nil {
		return err
	}

	w, err := demoPublicKey(string(pubKey))
	if err != nil {
		return err
	}

	log.Println("wg status", w.status)
	log.Println("wg pubkey", w.server_pubkey)
	log.Println("wg server port", w.server_port)
	internalIP := strings.TrimRightFunc(w.internal_ip, unicode.IsControl)
	log.Println("wg internal ip", string(internalIP[0:]))

	if err := run("ip", "link", "add", "dev", "wg0", "type", "wireguard"); err != nil {
		return err
	}

	if err := ioutil.WriteFile("pkey", []byte(pkey), 0444); err != nil {
		return err
	}

	if err := run(
		"wg", "set", "wg0", "private-key", "pkey", "peer", w.server_pubkey,
		"allowed-ips", "0.0.0.0/0", "endpoint", fmt.Sprintf("demo.wireguard.com:%s", w.server_port),
		"persistent-keepalive", "25",
	); err != nil {
		return err
	}

	if err := run("ip", "address", "add", fmt.Sprintf("%s/24", internalIP), "dev", "wg0"); err != nil {
		return err
	}

	if err := run("ip", "link", "set", "up", "wg0"); err != nil {
		return err
	}

	return nil
}

func down() error {
	return run("ip", "link", "del", "dev", "wg0")
}

func main() {
	mode := "up"
	if len(os.Args) == 2 {
		mode = os.Args[1]
	}

	var err error

	switch mode {
	case "up":
		err = up()
	case "down":
		err = down()
	default:
		err = errors.New("usage: demo [up|down]")
	}

	if err != nil {
		log.Fatal("failed:", err)
	}
}

type wg struct {
	status        string
	server_pubkey string
	server_port   string
	internal_ip   string
}

func (w wg) String() string {
	return fmt.Sprintf("(%s %s %s %s)", w.status, w.server_pubkey, w.server_port, w.internal_ip)
}

func demoPublicKey(publicKey string) (*wg, error) {
	endpoint := "demo.wireguard.com:42912"
	conn, err := net.Dial("tcp4", endpoint)
	if err != nil {
		return nil, err
	}
	conn.Write([]byte(publicKey))
	b := make([]byte, 1024)
	_, err = conn.Read(b)
	if err != nil {
		return nil, err
	}

	pieces := strings.Split(strings.TrimSpace(string(b)), ":")
	return &wg{
		status:        pieces[0],
		server_pubkey: pieces[1],
		server_port:   pieces[2],
		internal_ip:   pieces[3],
	}, nil
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
