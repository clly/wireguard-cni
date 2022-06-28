package wireguard

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
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
