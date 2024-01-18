package main

import (
	"log"

	"github.com/jgillich/kubegio/ui"
)

func main() {
	app, err := ui.NewApplication()
	if err != nil {
		log.Fatal(err)
	}
	app.Run()
}
