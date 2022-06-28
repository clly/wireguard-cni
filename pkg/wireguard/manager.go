package wireguard

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

func (w *WGQuickManager) Up(device string) error {
	return run("wg-quick", "up", device)
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

func (w *WGQuickManager) Config(writer io.Writer) error {
	return nil
}
