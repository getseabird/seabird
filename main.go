package main

import (
	"context"
	"log"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/internal/component"
	"github.com/getseabird/seabird/internal/reactive"
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

	gtk.Init()

	app := &component.App{Application: adw.NewApplication("dev.skynomads.Seabird", gio.ApplicationFlagsNone)}

	app.ConnectActivate(func() {
		tree := reactive.NewTree(context.Background(), reactive.CreateComponent(app))
		tree.(*adw.ApplicationWindow).Present()
	})
	app.Run(os.Args)

	// app, err := ui.NewApplication(version)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// app.Run()

}
