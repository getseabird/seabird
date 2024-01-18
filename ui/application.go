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
	navigation   *Navigation
	resourceView *ListView
	detailView   *DetailView
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
		application.resourceView = NewListView()

		box := gtk.NewBox(gtk.OrientationHorizontal, 0)
		box.Append(application.navigation)
		box.Append(application.resourceView)
		application.window.SetChild(box)
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
	box := application.window.Child().(*gtk.Box)
	if application.detailView != nil {
		box.Remove(application.detailView)
	}
	application.detailView = NewDetailView(object)
	box.Append(application.detailView)
}
