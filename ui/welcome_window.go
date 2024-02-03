package ui

import (
	"context"
	"os"
	"runtime"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/behavior"
	"github.com/getseabird/seabird/widget"
	"github.com/imkira/go-observer/v2"
)

type WelcomeWindow struct {
	*widget.UniversalApplicationWindow
	content  *adw.Bin
	behavior *behavior.Behavior
}

func NewWelcomeWindow(app *gtk.Application, behavior *behavior.Behavior) *WelcomeWindow {
	w := WelcomeWindow{
		UniversalApplicationWindow: widget.NewUniversalApplicationWindow(app),
		content:                    adw.NewBin(),
		behavior:                   behavior,
	}
	w.SetApplication(app)
	w.SetIconName("seabird")
	w.SetDefaultSize(600, 600)
	w.SetContent(w.content)
	w.content.SetChild(w.createContent())
	w.SetTitle(ApplicationName)

	return &w
}

func (w *WelcomeWindow) createContent() *adw.NavigationView {
	view := adw.NewNavigationView()
	view.ConnectPopped(func(page *adw.NavigationPage) {
		w.content.SetChild(w.createContent())
	})

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	view.Add(adw.NewNavigationPage(box, ApplicationName))

	if runtime.GOOS != "windows" {
		header := gtk.NewHeaderBar()
		box.Append(header)
	}

	page := adw.NewPreferencesPage()
	box.Append(page)

	if clusters := w.behavior.Preferences.Value().Clusters; len(clusters) > 0 {
		group := adw.NewPreferencesGroup()
		group.SetTitle("Connect to Cluster")
		page.Add(group)

		add := gtk.NewButton()
		add.AddCSSClass("flat")
		add.SetIconName("list-add")
		add.ConnectClicked(func() {
			pref := NewClusterPrefPage(&w.ApplicationWindow.Window, w.behavior, observer.NewProperty(behavior.ClusterPreferences{}))
			view.Push(pref.NavigationPage)
		})

		group.SetHeaderSuffix(add)

		for i, c := range w.behavior.Preferences.Value().Clusters {
			cluster := c
			row := adw.NewActionRow()
			row.SetTitle(cluster.Value().Name)
			row.SetActivatable(true)
			spinner := widget.NewFallbackSpinner(gtk.NewImageFromIconName("go-next-symbolic"))
			row.AddSuffix(spinner)
			row.ConnectActivated(func() {
				spinner.Start()
				go func() {
					behavior, err := w.behavior.WithCluster(context.TODO(), cluster)
					glib.IdleAdd(func() {
						spinner.Stop()
						if err != nil {
							ShowErrorDialog(&w.ApplicationWindow.Window, "Cluster connection failed", err)
							return
						}
						app := w.Application()
						w.Close()
						NewClusterWindow(app, behavior).Show()
					})
				}()
			})
			group.Add(row)
			if os.Getenv("SEABIRD_DEV") == "1" && i == 0 {
				defer row.Activate()
			}
		}
	} else {
		status := adw.NewStatusPage()
		status.SetIconName("seabird")
		status.SetTitle("No Clusters Found")
		status.SetDescription("Connect to a cluster to get started.")
		btn := gtk.NewButton()
		btn.ConnectClicked(func() {
			pref := NewClusterPrefPage(&w.ApplicationWindow.Window, w.behavior, observer.NewProperty(behavior.ClusterPreferences{}))
			view.Push(pref.NavigationPage)
		})
		btn.SetHAlign(gtk.AlignCenter)
		btn.SetLabel("New Cluster")
		btn.AddCSSClass("pill")
		btn.AddCSSClass("suggested-action")
		status.SetChild(btn)
		box.Append(status)
	}

	// term := vte.NewTerminal()
	// pty, err := term.PtyNewSync(context.TODO(), vte.PtyDefault)
	// if err != nil {
	// 	panic(err)
	// }

	// 	box.Append(term)

	return view
}
