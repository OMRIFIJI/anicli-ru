//go:build !windows

package ansi

func EnterAltScreenBuf() {
	enterAltScreenBufCommon()
}

func ExitAltScreenBuf() {
	exitAltScreenBufCommon()
}
