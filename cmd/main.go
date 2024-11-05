package main

import (
	"anicliru/internal/app"
	"log"
)

func main() {
	if err := app.RunApp(); err != nil {
		log.Fatal(err)
	}
	return
}
