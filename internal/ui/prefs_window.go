package ui

import (
	"context"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/behavior"
	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/imkira/go-observer/v2"
)

type PrefsWindow struct {
	*adw.PreferencesWindow
	ctx            context.Context
	behavior       *behavior.ClusterBehavior
	navigationView *adw.NavigationView
}

func NewPreferencesWindow(ctx context.Context, behavior *behavior.ClusterBehavior) *PrefsWindow {
	window := adw.NewPreferencesWindow()
	ctx = ctxt.With[*gtk.Window](ctx, &window.Window.Window)
	w := PrefsWindow{PreferencesWindow: window, behavior: behavior, ctx: ctx}

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
		page := NewClusterPrefPage(w.ctx, w.behavior.Behavior, observer.NewProperty(api.ClusterPreferences{}))
		w.navigationView.Push(page.NavigationPage)
	})

	clusters.SetHeaderSuffix(addCluster)
	for _, cluster := range w.behavior.Preferences.Value().Clusters {
		row := adw.NewActionRow()
		row.SetActivatable(true)
		row.ConnectActivated(func() {
			w.navigationView.Push(NewClusterPrefPage(w.ctx, w.behavior.Behavior, cluster).NavigationPage)
		})
		row.SetTitle(cluster.Value().Name)
		if kubeconfig := cluster.Value().Kubeconfig; kubeconfig != nil {
			label := gtk.NewLabel(kubeconfig.Path)
			label.AddCSSClass("dim-label")
			label.SetHAlign(gtk.AlignStart)
			row.AddSuffix(label)
		}
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
