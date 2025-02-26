//go:build windows

package main

import (
	"anicliru/internal/app"
	"fmt"
	"log"
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

	a, err := app.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := a.RunApp(); err != nil {
		fmt.Println(err)
	}
}
