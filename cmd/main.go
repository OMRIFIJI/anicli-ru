package main

import (
	"anicliru/internal/app"
	"fmt"
)

func main() {
    a := app.App{}
	if err := a.RunApp(); err != nil {
		fmt.Println(err)
	}
	return
}
