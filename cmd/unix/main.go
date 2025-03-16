//go:build !windows

package main

import (
	"anicliru/internal/app"
	"fmt"
	"os"
)

func main() {
	if err := app.RunApp(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
