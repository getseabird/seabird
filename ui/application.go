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

const ApplicationName = "kubegtk"

// TODO remove this
var application Application

type Application struct {
	*adw.Application
	version    string
	content    *adw.Bin
	window     *adw.ApplicationWindow
	mainGrid   *gtk.Grid
	navigation *Navigation
	listView   *ListView
	detailView *DetailView
	cluster    *state.Cluster
	prefs      *state.Preferences
}

func NewApplication(version string) (*Application, error) {
	gtk.Init()

	prefs, err := state.LoadPreferences()
	if err != nil {
		return nil, err
	}
	prefs.Defaults()

	application = Application{
		Application: adw.NewApplication("io.github.jgillich.kubegtk", gio.ApplicationFlagsNone),
		prefs:       prefs,
		version:     version,
	}
	application.ConnectActivate(func() {
		application.window = adw.NewApplicationWindow(&application.Application.Application)
		application.window.SetTitle(ApplicationName)

		application.content = adw.NewBin()
		application.content.SetChild(application.createWelcomeContent())
		application.window.SetContent(application.content)
		application.window.Show()
	})

	provider := gtk.NewCSSProvider()
	provider.LoadFromPath("theme.css")
	gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	return &application, nil
}

func (a *Application) Run() {
	if code := a.Application.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func (a *Application) createMainContent(cluster *state.Cluster) *gtk.Grid {
	a.cluster = cluster

	a.window.SetDefaultSize(1000, 1000)
	a.mainGrid = gtk.NewGrid()

	application.detailView = NewDetailView()
	a.mainGrid.Attach(application.detailView, 2, 0, 1, 1)

	application.listView = NewListView()
	a.mainGrid.Attach(application.listView, 1, 0, 1, 1)

	application.navigation = NewNavigation()
	a.mainGrid.Attach(application.navigation, 0, 0, 1, 1)

	return a.mainGrid
}

func (a *Application) createWelcomeContent() *adw.NavigationView {
	a.window.SetDefaultSize(600, 600)

	view := adw.NewNavigationView()
	view.ConnectPopped(func(page *adw.NavigationPage) {
		if err := application.prefs.Save(); err != nil {
			ShowErrorDialog(&a.window.Window, "Could not save preferences", err)
			return
		}
		application.content.SetChild(a.createWelcomeContent())
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
		pref := NewClusterPrefPage(&a.window.Window, nil)
		view.Push(pref.NavigationPage)
	})

	group.SetHeaderSuffix(add)

	for _, c := range application.prefs.Clusters {
		cluster := c
		row := adw.NewActionRow()
		row.SetTitle(cluster.Name)
		row.SetActivatable(true)
		row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
		row.ConnectActivated(func() {
			cluster, err := state.NewCluster(context.TODO(), cluster)
			if err != nil {
				ShowErrorDialog(&a.window.Window, "Cluster connection failed", err)
				return
			}
			application.content.SetChild(a.createMainContent(cluster))
		})
		group.Add(row)
	}

	return view
}
