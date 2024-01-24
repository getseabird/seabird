package ui

import (
	"context"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/state"
)

type WelcomeWindow struct {
	*adw.ApplicationWindow
	content *adw.Bin
	prefs   *state.Preferences
}

func NewWelcomeWindow(app *gtk.Application, prefs *state.Preferences) *WelcomeWindow {
	w := WelcomeWindow{
		ApplicationWindow: adw.NewApplicationWindow(app),
		content:           adw.NewBin(),
		prefs:             prefs,
	}
	w.SetDefaultSize(600, 600)
	w.SetContent(w.content)
	w.content.SetChild(w.createContent())

	return &w
}

func (w *WelcomeWindow) createContent() *adw.NavigationView {
	view := adw.NewNavigationView()
	view.ConnectPopped(func(page *adw.NavigationPage) {
		if err := w.prefs.Save(); err != nil {
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
		pref := NewClusterPrefPage(&w.Window, w.prefs, nil)
		view.Push(pref.NavigationPage)
	})

	group.SetHeaderSuffix(add)

	for _, c := range w.prefs.Clusters {
		cluster := c
		row := adw.NewActionRow()
		row.SetTitle(cluster.Name)
		row.SetActivatable(true)
		row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
		row.ConnectActivated(func() {
			cluster, err := state.NewCluster(context.TODO(), cluster)
			if err != nil {
				ShowErrorDialog(&w.Window, "Cluster connection failed", err)
				return
			}
			app := w.Application()
			w.Close()
			NewClusterWindow(app, cluster, w.prefs).Show()
		})
		group.Add(row)
	}

	return view
}
