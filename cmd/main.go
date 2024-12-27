package main

import (
	"anicliru/internal/app"
	"fmt"
    "log"
)

func main() {
    a, err := app.NewApp()
    if err != nil {
		log.Fatal(err)
    }

	if err := a.RunApp(); err != nil {
		fmt.Println(err)
	}

	return
}
