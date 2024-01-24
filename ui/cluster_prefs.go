package ui

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/state"
	"github.com/getseabird/seabird/util"
	"k8s.io/client-go/tools/clientcmd"
)

type ClusterPrefPage struct {
	*adw.NavigationPage
	parent     *gtk.Window
	prefs      *state.Preferences
	active     *state.ClusterPreferences
	page       *adw.PreferencesPage
	name       *adw.EntryRow
	host       *adw.EntryRow
	cert       *adw.EntryRow
	key        *adw.EntryRow
	ca         *adw.EntryRow
	favourites *adw.PreferencesGroup
}

func NewClusterPrefPage(parent *gtk.Window, prefs *state.Preferences, active *state.ClusterPreferences) *ClusterPrefPage {
	content := gtk.NewBox(gtk.OrientationVertical, 0)
	p := ClusterPrefPage{NavigationPage: adw.NewNavigationPage(content, "Cluster")}

	p.parent = parent
	p.prefs = prefs
	p.active = active

	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(false)
	header.PackEnd(p.createSaveButton())
	content.Append(header)
	p.page = adw.NewPreferencesPage()
	content.Append(p.page)

	if active == nil {
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

	if active != nil {
		p.setPrefs(active)

		delete := adw.NewActionRow()
		delete.SetActivatable(true)
		delete.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
		delete.SetTitle("Delete")
		delete.AddCSSClass("error")
		delete.ConnectActivated(func() {
			dialog := adw.NewMessageDialog(parent, "Delete cluster?", fmt.Sprintf("Are you sure you want to delete cluster \"%s\"?", active.Name))
			dialog.AddResponse("cancel", "Cancel")
			dialog.AddResponse("delete", "Delete")
			dialog.SetResponseAppearance("delete", adw.ResponseDestructive)
			dialog.Show()
			dialog.ConnectResponse(func(response string) {
				if response == "delete" {
					var idx int
					for i, c := range prefs.Clusters {
						if c == active {
							idx = i
							break
						}
					}
					prefs.Clusters = append(prefs.Clusters[:idx], prefs.Clusters[idx+1:]...)
					p.Parent().(*adw.NavigationView).Pop()
				}
			})
		})
		group := adw.NewPreferencesGroup()
		group.Add(delete)
		p.page.Add(group)
	}

	return &p
}

func (p *ClusterPrefPage) createFavourites() *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()
	group.SetTitle("Favourites")
	group.SetHeaderSuffix(p.createFavouritesAddButton())

	for i, fav := range p.active.Navigation.Favourites {
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
			p.active.Navigation.Favourites = append(p.active.Navigation.Favourites[:idx], p.active.Navigation.Favourites[idx+1:]...)
			p.setPrefs(p.active)
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

		cluster, _ := state.NewCluster(context.TODO(), p.active)
		for _, r := range cluster.Resources {
			res := r
			exists := false
			for _, fav := range p.active.Navigation.Favourites {
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
				p.active.Navigation.Favourites = append(p.active.Navigation.Favourites, util.ResourceGVR(&res))
				p.Parent().(*adw.NavigationView).Pop()
				p.setPrefs(p.active)
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

		cluster := state.ClusterPreferences{}
		cluster.Name = p.name.Text()
		cluster.Host = p.host.Text()
		cluster.TLS.CertData = []byte(p.cert.Text())
		cluster.TLS.KeyData = []byte(p.key.Text())
		cluster.TLS.CAData = []byte(p.ca.Text())
		cluster.Defaults()

		if _, err := state.NewCluster(context.TODO(), &cluster); err != nil {
			ShowErrorDialog(p.parent, "Cluster connection failed", err)
			return
		}

		if p.active == nil {
			p.prefs.Clusters = append(p.prefs.Clusters, &cluster)
		} else {
			p.active = &cluster
		}
		p.Parent().(*adw.NavigationView).Pop()
	})

	return button
}

func (p *ClusterPrefPage) setPrefs(clusterPrefs *state.ClusterPreferences) {
	p.active = clusterPrefs
	p.name.SetText(clusterPrefs.Name)
	p.host.SetText(clusterPrefs.Host)
	p.cert.SetText(string(clusterPrefs.TLS.CertData))
	p.key.SetText(string(clusterPrefs.TLS.KeyData))
	p.ca.SetText(string(clusterPrefs.TLS.CAData))

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
