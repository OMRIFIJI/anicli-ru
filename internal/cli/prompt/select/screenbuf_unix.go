//go:build !windows

package promptselect

import "anicliru/internal/cli/ansi"

func enterAltScreenBuf() {
	ansi.EnterAltScreenBuf()
}

func exitAltScreenBuf() {
	ansi.ExitAltScreenBuf()
}
