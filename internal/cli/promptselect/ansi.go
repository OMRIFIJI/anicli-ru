package promptselect

import (
	"fmt"
)

func enterAltScreenBuf() {
	fmt.Print("\033[?1049h")
}

func exitAltScreenBuf() {
	fmt.Print("\033[?1049l")
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func hideCursor() {
	fmt.Print("\033[?25l")
}

func showCursor() {
	fmt.Print("\033[?25h")
}

func moveCursorToNewLine() {
	fmt.Print("\033[E")
}
