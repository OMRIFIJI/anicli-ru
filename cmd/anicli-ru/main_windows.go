//go:build windows

package main

import (
	"fmt"
	"github.com/OMRIFIJI/anicli-ru/internal/app"
	"os"

	"golang.org/x/sys/windows"
)

// Необходимо для ANSI escape codes
func prepareTerminal() uint32 {
	stdout := windows.Handle(os.Stdout.Fd())
	var originalMode uint32

	windows.GetConsoleMode(stdout, &originalMode)
	windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	return originalMode
}

func restoreTerminal(originalMode uint32) {
	stdout := windows.Handle(os.Stdout.Fd())
	windows.SetConsoleMode(stdout, originalMode)
}

func main() {
	originalMode := prepareTerminal()
	defer restoreTerminal(originalMode)

	if err := app.RunApp(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
