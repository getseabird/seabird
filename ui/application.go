package ui

import (
	"context"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/jgillich/kubegio/state"
)

var application Application

type Application struct {
	*adw.Application
	window       *gtk.ApplicationWindow
	navigation   *Navigation
	resourceView *ResourceView
	cluster      *state.Cluster
}

func NewApplication() (*Application, error) {
	gtk.Init()

	cluster, err := state.NewCluster(context.TODO())
	if err != nil {
		return nil, err
	}

	application = Application{
		Application: adw.NewApplication("com.github.diamondburned.gotk4-examples.gtk4.simple", gio.ApplicationFlagsNone),
		cluster:     cluster,
	}

	application.ConnectActivate(func() {
		application.window = Window(&application.Application.Application)
		application.navigation = NewNavigation()
		application.resourceView = NewResourceView()

		paned := gtk.NewPaned(gtk.OrientationHorizontal)
		paned.SetStartChild(application.navigation)
		paned.SetEndChild(application.resourceView)
		application.window.SetChild(paned)
		application.window.Show()
	})

	provider := gtk.NewCSSProvider()
	provider.LoadFromPath("theme.css")
	gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), provider, 0)

	return &application, nil
}

func (a *Application) Run() {
	if code := a.Application.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
