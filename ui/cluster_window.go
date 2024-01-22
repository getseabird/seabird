package ui

import (
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/jgillich/kubegio/state"
)

type ClusterWindow struct {
	*adw.ApplicationWindow
	cluster    *state.Cluster
	prefs      *state.Preferences
	navigation *Navigation
	listView   *ListView
	detailView *DetailView
}

func NewClusterWindow(app *gtk.Application, cluster *state.Cluster, prefs *state.Preferences) *ClusterWindow {
	w := ClusterWindow{
		ApplicationWindow: adw.NewApplicationWindow(app),
		cluster:           cluster,
		prefs:             prefs,
	}
	w.SetTitle(fmt.Sprintf("%s - %s", cluster.Preferences.Name, ApplicationName))
	w.SetDefaultSize(1000, 1000)

	grid := gtk.NewGrid()
	w.SetContent(grid)

	w.detailView = NewDetailView(&w)
	grid.Attach(w.detailView, 2, 0, 1, 1)
	w.listView = NewListView(&w)
	grid.Attach(w.listView, 1, 0, 1, 1)
	w.navigation = NewNavigation(&w)
	grid.Attach(w.navigation, 0, 0, 1, 1)

	w.createActions()

	return &w
}

func (w *ClusterWindow) createActions() {
	newWindow := gio.NewSimpleAction("newWindow", nil)
	newWindow.ConnectActivate(func(_ *glib.Variant) {
		prefs, err := state.LoadPreferences()
		if err != nil {
			ShowErrorDialog(&w.Window, "Could not load preferences", err)
			return
		}
		prefs.Defaults()
		NewWelcomeWindow(w.Application(), prefs).Show()
	})
	w.AddAction(newWindow)

	disconnect := gio.NewSimpleAction("disconnect", nil)
	disconnect.ConnectActivate(func(_ *glib.Variant) {
		w.ActivateAction("newWindow", nil)
		w.Close()
	})
	w.AddAction(disconnect)

	action := gio.NewSimpleAction("prefs", nil)
	action.ConnectActivate(func(_ *glib.Variant) {
		prefs := NewPreferencesWindow(w)
		prefs.SetTransientFor(&w.Window)
		prefs.Show()
	})
	w.AddAction(action)

	action = gio.NewSimpleAction("about", nil)
	action.ConnectActivate(func(_ *glib.Variant) {
		NewAboutWindow(&w.Window).Show()
	})
	w.AddAction(action)
}
