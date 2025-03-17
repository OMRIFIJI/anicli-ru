//go:build !windows

package main

import (
	"github.com/OMRIFIJI/anicli-ru/internal/app"
	"fmt"
	"os"
)

func main() {
	if err := app.RunApp(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
