package ansi

import (
	"fmt"
)

func EnterAltScreenBuf() {
	fmt.Print("\033[?1049h")
}

func ExitAltScreenBuf() {
	fmt.Print("\033[?1049l")
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

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
	ColorReset  = "\033[0m"
	ColorPrompt = "\033[34m"
	ColorErr    = "\033[31m"
)
