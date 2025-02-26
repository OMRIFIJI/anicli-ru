//go:build windows

package promptselect

import (
	"anicliru/internal/cli/ansi"
	"os"
)

var isCmd = os.Getenv("COMSPEC")[len(os.Getenv("COMSPEC"))-7:] == "cmd.exe"

func enterAltScreenBuf() {
	if !isCmd {
		ansi.EnterAltScreenBuf()
	}
}

func exitAltScreenBuf() {
	if !isCmd {
		ansi.ExitAltScreenBuf()
	}
}
