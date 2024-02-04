package ui

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/behavior"
	"github.com/getseabird/seabird/widget"
)

type ClusterWindow struct {
	*widget.UniversalApplicationWindow
	behavior     *behavior.ClusterBehavior
	prefs        *behavior.Preferences
	navigation   *Navigation
	listView     *ListView
	detailView   *DetailView
	toastOverlay *adw.ToastOverlay
}

func NewClusterWindow(app *gtk.Application, behavior *behavior.ClusterBehavior) *ClusterWindow {
	w := ClusterWindow{
		UniversalApplicationWindow: widget.NewUniversalApplicationWindow(app),
		behavior:                   behavior,
	}
	w.SetIconName("seabird")
	w.SetTitle(fmt.Sprintf("%s - %s", behavior.ClusterPreferences.Value().Name, ApplicationName))
	w.SetDefaultSize(1300, 700)
	if runtime.GOOS == "windows" {
		w.SetDecorated(true) // https://gitlab.gnome.org/GNOME/gtk/-/issues/3749
	}

	var h glib.SignalHandle
	h = w.ConnectCloseRequest(func() bool {
		prefs := behavior.Preferences.Value()
		if err := prefs.Save(); err != nil {
			d := ShowErrorDialog(&w.Window, "Could not save preferences", err)
			d.ConnectUnrealize(func() {
				w.Close()
			})
			w.HandlerDisconnect(h)
			return true
		}
		return false
	})

	w.toastOverlay = adw.NewToastOverlay()
	w.SetContent(w.toastOverlay)

	grid := gtk.NewGrid()
	w.toastOverlay.SetChild(grid)

	w.detailView = NewDetailView(&w.Window, behavior.NewRootDetailBehavior())
	nav := adw.NewNavigationView()
	nav.Add(w.detailView.NavigationPage)
	grid.Attach(nav, 2, 0, 1, 1)
	w.listView = NewListView(&w.Window, behavior.NewListBehavior())
	grid.Attach(w.listView, 1, 0, 1, 1)
	w.navigation = NewNavigation(behavior)
	grid.Attach(w.navigation, 0, 0, 1, 1)

	w.createActions()

	return &w
}

func (w *ClusterWindow) createActions() {
	newWindow := gio.NewSimpleAction("newWindow", nil)
	newWindow.ConnectActivate(func(_ *glib.Variant) {
		prefs, err := behavior.LoadPreferences()
		if err != nil {
			ShowErrorDialog(&w.Window, "Could not load preferences", err)
			return
		}
		prefs.Defaults()
		NewWelcomeWindow(w.Application(), w.behavior.Behavior).Show()
	})
	w.AddAction(newWindow)
	w.Application().SetAccelsForAction("win.newWindow", []string{"<Ctrl>N"})

	disconnect := gio.NewSimpleAction("disconnect", nil)
	disconnect.ConnectActivate(func(_ *glib.Variant) {
		w.ActivateAction("newWindow", nil)
		w.Close()
	})
	w.AddAction(disconnect)
	w.Application().SetAccelsForAction("win.disconnect", []string{"<Ctrl>Q"})

	action := gio.NewSimpleAction("prefs", nil)
	action.ConnectActivate(func(_ *glib.Variant) {
		prefs := NewPreferencesWindow(w.behavior)
		prefs.SetTransientFor(&w.Window)
		prefs.Show()
	})
	w.AddAction(action)

	action = gio.NewSimpleAction("about", nil)
	action.ConnectActivate(func(_ *glib.Variant) {
		NewAboutWindow(&w.Window).Show()
	})
	w.AddAction(action)

	filterNamespace := gio.NewSimpleAction("filterNamespace", glib.NewVariantType("s"))
	filterNamespace.ConnectActivate(func(parameter *glib.Variant) {
		text := strings.Trim(fmt.Sprintf("%s ns:%s", w.behavior.SearchText.Value(), parameter.String()), " ")
		w.behavior.SearchText.Update(text)
	})
	actionGroup := gio.NewSimpleActionGroup()
	actionGroup.AddAction(filterNamespace)
	w.InsertActionGroup("list", actionGroup)
}
