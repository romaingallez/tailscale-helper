package main

import (
	"log"
	"os"

	"tailscale_helper/cmd"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("tailscale_helper: ")
	log.SetOutput(os.Stderr)
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
