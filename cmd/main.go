package main

import (
	"anicliru/internal/cli"
	"log"
)

func main() {
	if err := cli.StartApp(); err != nil {
		log.Fatal(err)
	}
	return
}
