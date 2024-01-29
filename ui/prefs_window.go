package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/behavior"
	"github.com/imkira/go-observer/v2"
)

type PrefsWindow struct {
	*adw.PreferencesWindow
	behavior       *behavior.ClusterBehavior
	navigationView *adw.NavigationView
}

func NewPreferencesWindow(behavior *behavior.ClusterBehavior) *PrefsWindow {
	w := PrefsWindow{PreferencesWindow: adw.NewPreferencesWindow(), behavior: behavior}

	content := gtk.NewBox(gtk.OrientationVertical, 0)
	w.navigationView = adw.NewNavigationView()
	w.navigationView.Add(adw.NewNavigationPage(content, "Preferences"))
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
	colorScheme.SetSelected(uint(w.behavior.Preferences.Value().ColorScheme))
	colorScheme.Connect("notify::selected-item", func() {
		prefs := w.behavior.Preferences.Value()
		prefs.ColorScheme = adw.ColorScheme(colorScheme.Selected())
		if prefs.ColorScheme == adw.ColorSchemePreferLight {
			prefs.ColorScheme = adw.ColorSchemeForceDark
		}
		w.behavior.Preferences.Update(prefs)
	})
	general.Add(colorScheme)

	clusters := adw.NewPreferencesGroup()
	clusters.SetTitle("Clusters")
	addCluster := gtk.NewButton()
	addCluster.AddCSSClass("flat")
	addCluster.SetIconName("list-add")
	addCluster.ConnectClicked(func() {
		page := NewClusterPrefPage(&w.Window.Window, w.behavior.Behavior, observer.NewProperty(behavior.ClusterPreferences{}))
		w.navigationView.Push(page.NavigationPage)
	})

	clusters.SetHeaderSuffix(addCluster)
	for _, c := range w.behavior.Preferences.Value().Clusters {
		cluster := c
		row := adw.NewActionRow()
		row.SetActivatable(true)
		row.ConnectActivated(func() {
			w.navigationView.Push(NewClusterPrefPage(&w.Window.Window, w.behavior.Behavior, cluster).NavigationPage)
		})
		row.SetTitle(cluster.Value().Name)
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
