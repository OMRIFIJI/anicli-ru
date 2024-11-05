package main

import (
	"anicliru/internal/app"
	"log"
)

func main() {
	if err := app.StartApp(); err != nil {
		log.Fatal(err)
	}
	return
}
