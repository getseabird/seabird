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
	window     *adw.ApplicationWindow
	mainGrid   *gtk.Grid
	navigation *Navigation
	listView   *ListView
	detailView *DetailView
	cluster    *state.Cluster
	prefs      *state.Preferences
}

func NewApplication() (*Application, error) {
	gtk.Init()

	prefs, err := state.LoadPreferences()
	if err != nil {
		return nil, err
	}

	prefs.Clusters = append(prefs.Clusters, state.ClusterPreferences{Name: "minikube"})
	prefs.Defaults()

	cluster, err := state.NewCluster(context.TODO(), prefs.Clusters[0])
	if err != nil {
		return nil, err
	}

	application = Application{
		Application: adw.NewApplication("com.github.diamondburned.gotk4-examples.gtk4.simple", gio.ApplicationFlagsNone),
		cluster:     cluster,
		prefs:       prefs,
	}

	application.ConnectActivate(func() {
		application.window = application.newWindow()
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

func (a *Application) newWindow() *adw.ApplicationWindow {
	window := adw.NewApplicationWindow(&a.Application.Application)
	window.SetTitle("kubegtk")
	window.SetDefaultSize(1000, 800)
	a.mainGrid = gtk.NewGrid()
	window.SetContent(a.mainGrid)

	application.navigation = NewNavigation()
	a.mainGrid.Attach(application.navigation, 0, 0, 1, 1)

	application.detailView = NewDetailView()
	a.mainGrid.Attach(application.detailView, 2, 0, 1, 1)

	application.listView = NewListView()
	a.mainGrid.Attach(application.listView, 1, 0, 1, 1)

	window.Show()

	return window
}
