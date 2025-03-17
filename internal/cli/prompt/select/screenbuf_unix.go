//go:build !windows

package promptselect

import "github.com/OMRIFIJI/anicli-ru/internal/cli/ansi"

func enterAltScreenBuf() {
	ansi.EnterAltScreenBuf()
}

func exitAltScreenBuf() {
	ansi.ExitAltScreenBuf()
}
