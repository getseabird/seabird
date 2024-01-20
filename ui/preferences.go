package ui

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/jgillich/kubegio/state"
	"k8s.io/client-go/tools/clientcmd"
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
	addCluster.AddCSSClass("flat")
	addCluster.SetIconName("list-add")
	addCluster.ConnectClicked(func() {
		p.NavigationView.Push(NewClusterPrefPage(nil, &p.Window.Window).NavigationPage)
	})

	clusters.SetHeaderSuffix(addCluster)
	for _, c := range application.prefs.Clusters {
		cluster := c
		row := adw.NewActionRow()
		row.SetActivatable(true)
		row.ConnectActivated(func() {
			p.NavigationView.Push(NewClusterPrefPage(cluster, &p.Window.Window).NavigationPage)
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

type ClusterPrefPage struct {
	*adw.NavigationPage
	name *adw.EntryRow
	host *adw.EntryRow
	cert *adw.EntryRow
	key  *adw.EntryRow
	ca   *adw.EntryRow
}

func NewClusterPrefPage(prefs *state.ClusterPreferences, window *gtk.Window) *ClusterPrefPage {
	content := gtk.NewBox(gtk.OrientationVertical, 0)
	headerBar := adw.NewHeaderBar()
	headerBar.SetShowEndTitleButtons(false)
	content.Append(headerBar)
	page := adw.NewPreferencesPage()
	content.Append(page)

	p := ClusterPrefPage{NavigationPage: adw.NewNavigationPage(content, "Cluster")}

	if prefs == nil {
		loadBtn := adw.NewActionRow()
		loadBtn.SetActivatable(true)
		loadBtn.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
		loadBtn.SetTitle("Load kubeconfig")
		loadBtn.AddCSSClass("suggested-action")
		group := adw.NewPreferencesGroup()
		group.Add(loadBtn)
		page.Add(group)

		loadBtn.ConnectActivated(func() {
			fileChooser := gtk.NewFileChooserNative("Select kubeconfig", window, gtk.FileChooserActionOpen, "Open", "Cancel")
			filter := gtk.NewFileFilter()
			filter.AddMIMEType("text/plain")
			filter.AddMIMEType("application/yaml")
			filter.AddMIMEType("application/json")
			filter.SetName("Text")
			fileChooser.AddFilter(filter)
			fileChooser.ConnectResponse(func(responseId int) {
				if responseId == int(gtk.ResponseAccept) {
					path := fileChooser.File().Path()
					config, err := clientcmd.BuildConfigFromFlags("", path)
					if err != nil {
						dlg := adw.NewMessageDialog(window, "Error loading kubeconfig", err.Error())
						dlg.AddResponse("Ok", "Ok")
						dlg.Show()
						return
					}
					p.name.SetText(config.ServerName)
					p.host.SetText(config.Host)
					if config.CertFile != "" {
						data, _ := os.ReadFile(config.CertFile)
						p.cert.SetText(string(data))
					} else {
						p.cert.SetText(string(config.CertData))
					}
					if config.KeyFile != "" {
						data, _ := os.ReadFile(config.KeyFile)
						p.key.SetText(string(data))
					} else {
						p.key.SetText(string(config.KeyData))
					}
					if config.CAFile != "" {
						data, _ := os.ReadFile(config.CAFile)
						p.ca.SetText(string(data))
					} else {
						p.ca.SetText(string(config.CAData))
					}
				}
			})
			fileChooser.Show()
		})
	}

	general := adw.NewPreferencesGroup()
	page.Add(general)
	general.SetTitle("General")
	p.name = adw.NewEntryRow()
	p.name.SetTitle("Name")
	general.Add(p.name)
	p.host = adw.NewEntryRow()
	p.host.SetTitle("Host")
	general.Add(p.host)

	auth := adw.NewPreferencesGroup()
	page.Add(auth)
	auth.SetTitle("Authentication")
	p.cert = adw.NewEntryRow()
	p.cert.SetTitle("Client certificate")
	auth.Add(p.cert)
	p.key = adw.NewEntryRow()
	p.key.SetTitle("Client key")
	auth.Add(p.key)
	p.ca = adw.NewEntryRow()
	p.ca.SetTitle("CA certificate")
	auth.Add(p.ca)

	if prefs != nil {
		p.name.SetText(prefs.Name)
		p.host.SetText(prefs.Host)
		p.cert.SetText(string(prefs.TLS.CertData))
		p.key.SetText(string(prefs.TLS.KeyData))
		p.ca.SetText(string(prefs.TLS.CAData))

		resources := adw.NewPreferencesGroup()
		page.Add(resources)
		resources.SetTitle("Favourites")
		for _, fav := range prefs.Navigation.Favourites {
			ar := adw.NewActionRow()
			ar.AddCSSClass("property")
			if fav.Group == "" {
				fav.Group = "core"
			}
			ar.SetTitle(fmt.Sprintf("%s/%s", fav.Group, fav.Version))
			ar.SetSubtitle(fav.Resource)
			resources.Add(ar)
		}
	}

	saveBtn := gtk.NewButton()
	saveBtn.SetLabel("Save")
	saveBtn.AddCSSClass("suggested-action")
	saveBtn.ConnectClicked(func() {
		if err := p.validate(); err != nil {
			dlg := adw.NewMessageDialog(window, "Error", err.Error())
			dlg.AddResponse("Ok", "Ok")
			dlg.Show()
			return
		}

		newPrefs := state.ClusterPreferences{}
		newPrefs.Name = p.name.Text()
		newPrefs.Host = p.host.Text()
		newPrefs.TLS.CertData = []byte(p.cert.Text())
		newPrefs.TLS.KeyData = []byte(p.key.Text())
		newPrefs.TLS.CAData = []byte(p.ca.Text())

		if _, err := state.NewCluster(context.TODO(), &newPrefs); err != nil {
			dlg := adw.NewMessageDialog(window, "Could not connect to cluster", err.Error())
			dlg.AddResponse("Ok", "Ok")
			dlg.Show()
			return
		}

		if prefs == nil {
			application.prefs.Clusters = append(application.prefs.Clusters, &newPrefs)
		} else {
			prefs = &newPrefs
		}
		if err := application.prefs.Save(); err != nil {
			dlg := adw.NewMessageDialog(window, "Could not save config", err.Error())
			dlg.AddResponse("Ok", "Ok")
			dlg.Show()
			return
		}
		p.Parent().(*adw.NavigationView).Pop()
	})
	headerBar.PackEnd(saveBtn)

	return &p
}

func (p *ClusterPrefPage) validate() error {
	if len(p.name.Text()) == 0 {
		return errors.New("Name is required")
	}

	return nil
}
