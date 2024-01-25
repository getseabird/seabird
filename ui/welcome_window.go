package ui

import (
	"context"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/behavior"
	"github.com/imkira/go-observer/v2"
)

type WelcomeWindow struct {
	*adw.ApplicationWindow
	content  *adw.Bin
	behavior *behavior.Behavior
}

func NewWelcomeWindow(app *gtk.Application, behavior *behavior.Behavior) *WelcomeWindow {
	w := WelcomeWindow{
		ApplicationWindow: adw.NewApplicationWindow(app),
		content:           adw.NewBin(),
		behavior:          behavior,
	}
	w.SetDefaultSize(600, 600)
	w.SetContent(w.content)
	w.content.SetChild(w.createContent())

	return &w
}

func (w *WelcomeWindow) createContent() *adw.NavigationView {
	view := adw.NewNavigationView()
	view.ConnectPopped(func(page *adw.NavigationPage) {
		prefs := w.behavior.Preferences.Value()
		if err := prefs.Save(); err != nil {
			ShowErrorDialog(&w.Window, "Could not save preferences", err)
			return
		}
		w.content.SetChild(w.createContent())
	})

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	view.Add(adw.NewNavigationPage(box, ApplicationName))

	header := adw.NewHeaderBar()
	box.Append(header)

	page := adw.NewPreferencesPage()
	box.Append(page)

	group := adw.NewPreferencesGroup()
	group.SetTitle("Connect to Cluster")
	page.Add(group)

	add := gtk.NewButton()
	add.AddCSSClass("flat")
	add.SetIconName("list-add")
	add.ConnectClicked(func() {
		pref := NewClusterPrefPage(&w.Window, w.behavior, observer.NewProperty(behavior.ClusterPreferences{}))
		view.Push(pref.NavigationPage)
	})

	group.SetHeaderSuffix(add)

	for _, c := range w.behavior.Preferences.Value().Clusters {
		cluster := c
		row := adw.NewActionRow()
		row.SetTitle(cluster.Value().Name)
		row.SetActivatable(true)
		row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
		row.ConnectActivated(func() {
			cluster, err := w.behavior.WithCluster(context.TODO(), cluster)
			if err != nil {
				ShowErrorDialog(&w.Window, "Cluster connection failed", err)
				return
			}
			app := w.Application()
			w.Close()
			NewClusterWindow(app, cluster).Show()
		})
		group.Add(row)
	}

	return view
}
