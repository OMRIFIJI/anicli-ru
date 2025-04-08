package ansi

import (
	"fmt"
)

func HideCursor() {
	fmt.Print("\033[?25l")
}

func ShowCursor() {
	fmt.Print("\033[?25h")
}

func ClearLine() {
	fmt.Print("\r\033[2K")
}

const (
	ClearScreen = "\033[H\033[2J"
	ColorReset  = "\033[0m"
	ColorPrompt = "\033[34m"
	ColorErr    = "\033[31m"
)

func enterAltScreenBufCommon() {
	fmt.Print("\033[?1049h")
}

func exitAltScreenBufCommon() {
	fmt.Print("\033[?1049l")
}
