package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type PrefsWindow struct {
	*adw.PreferencesWindow
	navigationView *adw.NavigationView
	root           *ClusterWindow
}

func NewPreferencesWindow(root *ClusterWindow) *PrefsWindow {
	w := PrefsWindow{PreferencesWindow: adw.NewPreferencesWindow(), root: root}

	content := gtk.NewBox(gtk.OrientationVertical, 0)
	w.navigationView = adw.NewNavigationView()
	w.navigationView.Add(adw.NewNavigationPage(content, "main"))
	w.SetContent(w.navigationView)

	header := adw.NewHeaderBar()
	view := adw.NewViewSwitcher()
	view.SetPolicy(adw.ViewSwitcherPolicyWide)
	header.SetTitleWidget(view)
	content.Append(header)

	stack := adw.NewViewStack()
	generalPage := adw.NewBin()
	generalPage.SetChild(w.createGeneralPage())
	stack.AddTitledWithIcon(generalPage, "general", "General", "document-properties-symbolic")
	content.Append(stack)
	view.SetStack(stack)

	w.ConnectUnrealize(func() {
		if err := root.prefs.Save(); err != nil {
			ShowErrorDialog(&w.Window.Window, "Could not save preferences", err)
			return
		}
	})

	w.navigationView.ConnectPopped(func(page *adw.NavigationPage) {
		generalPage.SetChild(w.createGeneralPage())
	})

	return &w
}

func (w *PrefsWindow) createGeneralPage() gtk.Widgetter {
	page := adw.NewPreferencesPage()

	general := adw.NewPreferencesGroup()
	colorScheme := adw.NewComboRow()
	colorScheme.SetTitle("Color Scheme")
	themes := gtk.NewStringList([]string{"Default", "Light", "Dark"})
	colorScheme.SetModel(themes)
	colorScheme.SetSelected(uint(w.root.prefs.ColorScheme))
	colorScheme.Connect("notify::selected-item", func() {
		w.root.prefs.ColorScheme = adw.ColorScheme(colorScheme.Selected())
		if w.root.prefs.ColorScheme == adw.ColorSchemePreferLight {
			w.root.prefs.ColorScheme = adw.ColorSchemeForceDark
		}
		adw.StyleManagerGetDefault().SetColorScheme(adw.ColorScheme(w.root.prefs.ColorScheme))
	})
	general.Add(colorScheme)

	clusters := adw.NewPreferencesGroup()
	clusters.SetTitle("Clusters")
	addCluster := gtk.NewButton()
	addCluster.AddCSSClass("flat")
	addCluster.SetIconName("list-add")
	addCluster.ConnectClicked(func() {
		w.navigationView.Push(NewClusterPrefPage(&w.Window.Window, w.root.prefs, nil).NavigationPage)
	})

	clusters.SetHeaderSuffix(addCluster)
	for _, c := range w.root.prefs.Clusters {
		cluster := c
		row := adw.NewActionRow()
		row.SetActivatable(true)
		row.ConnectActivated(func() {
			w.navigationView.Push(NewClusterPrefPage(&w.Window.Window, w.root.prefs, cluster).NavigationPage)
		})
		row.SetTitle(cluster.Name)
		row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
		clusters.Add(row)
	}

	page.Add(general)
	page.Add(clusters)

	return page
}

func (p *PrefsWindow) other() gtk.Widgetter {
	return gtk.NewLabel("other")
}
