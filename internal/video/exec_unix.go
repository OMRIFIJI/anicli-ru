//go:build !windows

package video

import "os/exec"

func execCommand(mpvCmd string) *exec.Cmd {
	return exec.Command("/bin/bash", "-c", mpvCmd)
}
