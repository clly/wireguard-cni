package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"strings"
	"unicode"

	"github.com/bitfield/script"
)

func main() {
	pkey, err := script.Exec("wg genkey").String()
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(pkey)
	pubKey, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	wgstat, err := demoPublicKey(string(pubKey))
	if err != nil {
		panic(err)
	}
	fmt.Println(wgstat.status)
	fmt.Println(wgstat.server_pubkey)
	fmt.Println(wgstat.server_port)
	internalIP := strings.TrimRightFunc(wgstat.internal_ip, unicode.IsControl)
	fmt.Println(string(internalIP[0:]))

	//noErr(sh("ip link del dev wg0"))
	noErr(sh("pwd"))
	//noErr(sh("ip link add dev wg0 type wireguard"))
	//checkPipe(script.Exec("ip link add dev wg0 type wireguard"))
	ioutil.WriteFile("pkey", []byte(pkey), 0444)
	noErr(sh(fmt.Sprintf("wg set wg0 private-key pkey peer %s allowed-ips 0.0.0.0/0 endpoint demo.wireguard.com:%s persistent-keepalive 25", wgstat.server_pubkey, wgstat.server_port)))
	noErr(sh(fmt.Sprintf("ip address add %s/24 dev wg0", internalIP)))
	noErr(sh("ip link set up dev wg0"))
}

func noErr(err error) {
	if err != nil {
		panic(err)
	}
}
func checkPipe(p *script.Pipe) {
	fmt.Println(p.ExitStatus())
	if p.ExitStatus() != 0 {
		p.Stdout()
		panic(p.Error())
	}
}

type wg struct {
	status        string
	server_pubkey string
	server_port   string
	internal_ip   string
}

func demoPublicKey(publicKey string) (wg, error) {
	endpoint := "demo.wireguard.com:42912"
	conn, err := net.Dial("tcp4", endpoint)
	if err != nil {
		return wg{}, err
	}
	conn.Write([]byte(publicKey))
	b := make([]byte, 1024)
	_, err = conn.Read(b)
	if err != nil {
		return wg{}, err
	}

	pieces := strings.Split(strings.TrimSpace(string(b)), ":")
	w := wg{
		status:        pieces[0],
		server_pubkey: pieces[1],
		server_port:   pieces[2],
		internal_ip:   pieces[3],
	}
	return w, nil
}

func sh(c string) error {

	args := strings.Split(c, " ")
	exec.LookPath(args[0])
	cmd := exec.Command(args[0], args[1:]...)
	b, err := cmd.Output()
	fmt.Printf("%s\n", b)
	if err != nil {
		return fmt.Errorf("failed to execute %s %w", c, err)
	}
	return nil
}
