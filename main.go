package main

import (
	"log"
	"os"

	"github.com/getseabird/seabird/internal/ui"

	"net/http"
	_ "net/http/pprof"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if os.Getenv("SEABIRD_DEV") == "1" {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	ui.Version = version
	app, err := ui.NewApplication(version)
	if err != nil {
		log.Fatal(err)
	}
	app.Run()
}
