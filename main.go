package main

import (
	"log"

	"github.com/getseabird/seabird/ui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	ui.Version = version
	app, err := ui.NewApplication(version)
	if err != nil {
		log.Fatal(err)
	}
	app.Run()
}
