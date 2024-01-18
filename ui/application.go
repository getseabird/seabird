package ui

import (
	"context"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/jgillich/kubegio/state"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var application Application

type Application struct {
	*adw.Application
	window       *gtk.ApplicationWindow
	grid         *gtk.Grid
	navigation   *Navigation
	resourceView *ListView
	detailView   *DetailView
	cluster      *state.Cluster
	config       *state.Config
}

func NewApplication() (*Application, error) {
	gtk.Init()

	cluster, err := state.NewCluster(context.TODO())
	if err != nil {
		return nil, err
	}

	config, err := state.LoadConfig()
	if err != nil {
		return nil, err
	}

	application = Application{
		Application: adw.NewApplication("com.github.diamondburned.gotk4-examples.gtk4.simple", gio.ApplicationFlagsNone),
		cluster:     cluster,
		config:      config,
	}

	application.ConnectActivate(func() {
		application.window = Window(&application.Application.Application)
		application.navigation = NewNavigation()
		application.resourceView = NewListView()
		application.grid = gtk.NewGrid()

		application.grid.Attach(application.navigation, 0, 0, 1, 1)
		application.grid.Attach(application.resourceView, 1, 0, 1, 1)
		application.window.SetChild(application.grid)
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

func (a *Application) DetailView(object client.Object) {
	if application.detailView != nil {
		a.grid.Remove(application.detailView)
	}
	application.detailView = NewDetailView(object)
	a.grid.Attach(application.detailView, 2, 0, 1, 1)
}
