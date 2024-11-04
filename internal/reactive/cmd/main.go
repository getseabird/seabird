package main

import (
	"context"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	r "github.com/getseabird/seabird/internal/reactive"
)

func main() {
	gtk.Init()

	app := gtk.NewApplication("dev.skynomads.Seabird", gio.ApplicationFlagsNone)

	window := r.AdwApplicationWindow{
		ApplicationWindow: r.ApplicationWindow{
			Application: app,
			Window: r.Window{
				Title:         "Hello World",
				DefaultHeight: 300,
				DefaultWidth:  400,
			},
		},
		Content: r.CreateComponent(&SampleComponent{}),
	}

	app.ConnectActivate(func() {
		w := r.NewTree(context.Background(), &window).(*adw.ApplicationWindow)
		w.Present()
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
