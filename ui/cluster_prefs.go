package ui

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/jgillich/kubegio/state"
	"github.com/jgillich/kubegio/util"
	"k8s.io/client-go/tools/clientcmd"
)

type ClusterPrefPage struct {
	*adw.NavigationPage
	parent     *gtk.Window
	prefs      *state.ClusterPreferences
	page       *adw.PreferencesPage
	name       *adw.EntryRow
	host       *adw.EntryRow
	cert       *adw.EntryRow
	key        *adw.EntryRow
	ca         *adw.EntryRow
	favourites *adw.PreferencesGroup
}

func NewClusterPrefPage(parent *gtk.Window, prefs *state.ClusterPreferences) *ClusterPrefPage {
	content := gtk.NewBox(gtk.OrientationVertical, 0)
	p := ClusterPrefPage{NavigationPage: adw.NewNavigationPage(content, "Cluster")}

	p.parent = parent
	p.prefs = prefs

	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(false)
	header.PackEnd(p.createSaveButton())
	content.Append(header)
	p.page = adw.NewPreferencesPage()
	content.Append(p.page)

	if prefs == nil {
		group := adw.NewPreferencesGroup()
		p.page.Add(group)
		group.Add(p.createLoadActionRow())
	}

	general := adw.NewPreferencesGroup()
	p.page.Add(general)
	p.name = adw.NewEntryRow()
	p.name.SetTitle("Name")
	general.Add(p.name)
	p.host = adw.NewEntryRow()
	p.host.SetTitle("Host")
	general.Add(p.host)

	auth := adw.NewExpanderRow()
	general.Add(auth)
	auth.SetTitle("Authentication")
	p.cert = adw.NewEntryRow()
	p.cert.SetTitle("Client certificate")
	auth.AddRow(p.cert)
	p.key = adw.NewEntryRow()
	p.key.SetTitle("Client key")
	auth.AddRow(p.key)
	p.ca = adw.NewEntryRow()
	p.ca.SetTitle("CA certificate")
	auth.AddRow(p.ca)

	if prefs != nil {
		p.setPrefs(prefs)
	}

	return &p
}

func (p *ClusterPrefPage) createFavourites() *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()
	group.SetTitle("Favourites")
	group.SetHeaderSuffix(p.createFavouritesAddButton())

	for i, fav := range p.prefs.Navigation.Favourites {
		idx := i
		row := adw.NewActionRow()
		row.AddCSSClass("property")
		if fav.Group == "" {
			row.SetTitle(fav.Version)
		} else {
			row.SetTitle(fmt.Sprintf("%s/%s", fav.Group, fav.Version))
		}
		row.SetSubtitle(fav.Resource)

		remove := gtk.NewButton()
		remove.AddCSSClass("flat")
		remove.SetIconName("user-trash-symbolic")
		remove.ConnectClicked(func() {
			p.prefs.Navigation.Favourites = append(p.prefs.Navigation.Favourites[:idx], p.prefs.Navigation.Favourites[idx+1:]...)
			p.setPrefs(p.prefs)
		})
		row.AddSuffix(remove)
		group.Add(row)
	}

	return group
}

func (p *ClusterPrefPage) createFavouritesAddButton() *gtk.Button {
	button := gtk.NewButton()
	button.AddCSSClass("flat")
	button.SetIconName("list-add")
	button.ConnectClicked(func() {
		content := gtk.NewBox(gtk.OrientationVertical, 0)
		page := adw.NewNavigationPage(content, "Add Resource")
		p.Parent().(*adw.NavigationView).Push(page)

		header := adw.NewHeaderBar()
		header.SetShowEndTitleButtons(false)
		content.Append(header)

		group := adw.NewPreferencesGroup()
		group.SetTitle("Select Resource")
		pp := adw.NewPreferencesPage()
		pp.SetVExpand(true)
		pp.Add(group)
		sw := gtk.NewScrolledWindow()
		sw.SetChild(pp)
		content.Append(sw)

		cluster, _ := state.NewCluster(context.TODO(), p.prefs)
		for _, r := range cluster.Resources {
			res := r
			exists := false
			for _, fav := range p.prefs.Navigation.Favourites {
				if util.ResourceGVR(&res).String() == fav.String() {
					exists = true
				}
			}
			if exists {
				continue
			}

			row := adw.NewActionRow()
			row.AddCSSClass("property")
			if res.Group == "" {
				row.SetTitle(res.Version)
			} else {
				row.SetTitle(fmt.Sprintf("%s/%s", res.Group, res.Version))
			}
			row.SetSubtitle(res.Name)
			row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
			row.SetActivatable(true)
			row.ConnectActivated(func() {
				p.prefs.Navigation.Favourites = append(p.prefs.Navigation.Favourites, util.ResourceGVR(&res))
				p.Parent().(*adw.NavigationView).Pop()
				p.setPrefs(p.prefs)
			})
			group.Add(row)
		}
	})

	return button
}

func (p *ClusterPrefPage) createLoadActionRow() *adw.ActionRow {
	row := adw.NewActionRow()
	row.SetActivatable(true)
	row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
	row.SetTitle("Load kubeconfig")

	row.ConnectActivated(func() {
		fileChooser := gtk.NewFileChooserNative("Select kubeconfig", p.parent, gtk.FileChooserActionOpen, "Open", "Cancel")
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
					ShowErrorDialog(p.parent, "Error loading kubeconfig", err)
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

	return row
}

func (p *ClusterPrefPage) createSaveButton() *gtk.Button {
	button := gtk.NewButton()
	button.SetLabel("Save")
	button.AddCSSClass("suggested-action")
	button.ConnectClicked(func() {
		if err := p.validate(); err != nil {
			ShowErrorDialog(p.parent, "Validation failed", err)
			return
		}

		newPrefs := state.ClusterPreferences{}
		newPrefs.Name = p.name.Text()
		newPrefs.Host = p.host.Text()
		newPrefs.TLS.CertData = []byte(p.cert.Text())
		newPrefs.TLS.KeyData = []byte(p.key.Text())
		newPrefs.TLS.CAData = []byte(p.ca.Text())
		newPrefs.Defaults()

		if _, err := state.NewCluster(context.TODO(), &newPrefs); err != nil {
			ShowErrorDialog(p.parent, "Cluster connection failed", err)
			return
		}

		if p.prefs == nil {
			application.prefs.Clusters = append(application.prefs.Clusters, &newPrefs)
		} else {
			p.prefs = &newPrefs
		}
		p.Parent().(*adw.NavigationView).Pop()
	})

	return button
}

func (p *ClusterPrefPage) setPrefs(prefs *state.ClusterPreferences) {
	p.prefs = prefs
	p.name.SetText(prefs.Name)
	p.host.SetText(prefs.Host)
	p.cert.SetText(string(prefs.TLS.CertData))
	p.key.SetText(string(prefs.TLS.KeyData))
	p.ca.SetText(string(prefs.TLS.CAData))

	if p.favourites != nil {
		p.page.Remove(p.favourites)
	}
	p.favourites = p.createFavourites()
	p.page.Add(p.favourites)
}

func (p *ClusterPrefPage) validate() error {
	if len(p.name.Text()) == 0 {
		return errors.New("Name is required")
	}

	return nil
}
