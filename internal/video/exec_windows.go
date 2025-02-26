//go:build windows

package video

import (
	"fmt"
	"os/exec"
	"syscall"
)

func execCommand(mpvCmd string) *exec.Cmd {
	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: fmt.Sprintf(`/c "%s"`, mpvCmd)}
	return cmd
}
