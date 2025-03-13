//go:build !windows

package main

import (
	"anicliru/internal/app"
	"fmt"
)

func main() {
	if err := app.RunApp(); err != nil {
		fmt.Println(err)
	}
}
