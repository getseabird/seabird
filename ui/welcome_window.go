package ui

import (
	"context"
	"os"
	"runtime"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
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

	banner := adw.NewBanner("Your free trial expires in ∞ days")
	// banner.SetRevealed(true)
	banner.SetButtonLabel("Purchase")
	banner.ConnectButtonClicked(func() {
		view.Push(w.createPurchasePage())
	})
	box.Append(banner)

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

	return view
}

func (w *WelcomeWindow) createPurchasePage() *adw.NavigationPage {
	content := gtk.NewBox(gtk.OrientationVertical, 0)
	navPage := adw.NewNavigationPage(content, "Purchase")

	header := adw.NewHeaderBar()
	content.Append(header)

	prefPage := adw.NewPreferencesPage()
	content.Append(prefPage)

	group := adw.NewPreferencesGroup()
	group.SetTitle("Purchase Seabird")
	group.SetDescription("There is no time limit for testing Seabird. When you buy a license, you not only get priority support but also help secure the future development of Seabird.")
	prefPage.Add(group)

	action := adw.NewActionRow()
	action.SetTitle("Purchase now")
	action.SetActivatable(true)
	action.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
	action.ConnectActivated(func() {
		gtk.ShowURI(&w.Window, "https://seabird.lemonsqueezy.com/checkout/buy/7f6c107c-b8e8-4a28-b99e-35d18669ad37", gdk.CURRENT_TIME)
	})
	group.Add(action)

	entry := adw.NewEntryRow()
	entry.SetTitle("License key")
	entry.SetShowApplyButton(true)
	group.Add(entry)

	return navPage
}
