package ui

import (
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/jgillich/kubegio/state"
)

type PreferencesWindow struct {
	*adw.PreferencesWindow
	NavigationView *adw.NavigationView
}

func NewPreferencesWindow() *PreferencesWindow {
	p := PreferencesWindow{PreferencesWindow: adw.NewPreferencesWindow()}

	mainPage := gtk.NewBox(gtk.OrientationVertical, 0)
	p.NavigationView = adw.NewNavigationView()
	p.NavigationView.Add(adw.NewNavigationPage(mainPage, "asd"))
	p.SetContent(p.NavigationView)

	headerBar := adw.NewHeaderBar()
	viewSwitcher := adw.NewViewSwitcher()
	viewSwitcher.SetPolicy(adw.ViewSwitcherPolicyWide)
	headerBar.SetTitleWidget(viewSwitcher)
	mainPage.Append(headerBar)

	viewStack := adw.NewViewStack()
	viewStack.AddTitledWithIcon(p.general(), "general", "Clusters", "document-properties-symbolic")
	viewStack.AddTitledWithIcon(p.other(), "other", "other", "accessories-text-editor-symbolic")
	mainPage.Append(viewStack)
	viewSwitcher.SetStack(viewStack)

	btn := gtk.NewButton()
	btn.SetLabel("fooooo")
	btn.ConnectClicked(func() {
		clusterPage := gtk.NewBox(gtk.OrientationVertical, 0)
		headerBar := adw.NewHeaderBar()
		clusterPage.Append(headerBar)
		navigationPage := adw.NewNavigationPage(clusterPage, "Cluster")
		p.NavigationView.Add(navigationPage)
		p.NavigationView.Push(navigationPage)
	})
	mainPage.Append(btn)

	return &p
}

func (p *PreferencesWindow) general() gtk.Widgetter {
	page := adw.NewPreferencesPage()

	general := adw.NewPreferencesGroup()
	general.SetTitle("General")
	theme := adw.NewComboRow()
	theme.SetTitle("Theme")
	themes := gtk.NewStringList([]string{"Dark", "Light"})
	theme.SetModel(themes)
	general.Add(theme)

	clusters := adw.NewPreferencesGroup()
	clusters.SetTitle("Clusters")
	addCluster := gtk.NewButton()
	addCluster.SetLabel("New Cluster")
	clusters.SetHeaderSuffix(addCluster)
	for _, c := range application.prefs.Clusters {
		cluster := c
		row := adw.NewActionRow()
		row.SetActivatable(true)
		row.ConnectActivated(func() {
			p.NavigationView.Push(p.newClusterPrefPage(cluster))
		})
		row.SetTitle(cluster.Name)
		row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
		clusters.Add(row)
	}

	page.Add(general)
	page.Add(clusters)

	return page
}

func (p *PreferencesWindow) other() gtk.Widgetter {
	return gtk.NewLabel("other")
}

func (p *PreferencesWindow) newClusterPrefPage(pref state.ClusterPreferences) *adw.NavigationPage {
	content := gtk.NewBox(gtk.OrientationVertical, 0)
	headerBar := adw.NewHeaderBar()
	content.Append(headerBar)

	general := adw.NewPreferencesGroup()
	general.SetTitle("General")
	name := adw.NewEntryRow()
	name.SetTitle("Name")
	general.Add(name)

	resources := adw.NewPreferencesGroup()
	resources.SetTitle("Favourites")
	for _, fav := range pref.Navigation.Favourites {
		ar := adw.NewActionRow()
		ar.AddCSSClass("property")
		if fav.Group == "" {
			fav.Group = "core"
		}
		ar.SetTitle(fmt.Sprintf("%s/%s", fav.Group, fav.Version))
		ar.SetSubtitle(fav.Resource)
		resources.Add(ar)
	}

	prefPage := adw.NewPreferencesPage()
	prefPage.Add(general)
	prefPage.Add(resources)
	content.Append(prefPage)

	page := adw.NewNavigationPage(content, "Cluster")
	return page
}
