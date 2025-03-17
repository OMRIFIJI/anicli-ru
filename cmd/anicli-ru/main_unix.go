//go:build !windows

package main

import (
	"fmt"
	"github.com/OMRIFIJI/anicli-ru/internal/app"
	"os"
)

func main() {
	if err := app.RunApp(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
