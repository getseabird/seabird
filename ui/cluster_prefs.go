package ui

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/behavior"
	"github.com/getseabird/seabird/util"
	"github.com/getseabird/seabird/widget"
	"github.com/imkira/go-observer/v2"
	"golang.org/x/exp/maps"
	"k8s.io/client-go/tools/clientcmd"
)

type ClusterPrefPage struct {
	*adw.NavigationPage
	parent     *gtk.Window
	content    *adw.Bin
	behavior   *behavior.Behavior
	prefs      observer.Property[behavior.ClusterPreferences]
	name       *adw.EntryRow
	host       *adw.EntryRow
	cert       *adw.EntryRow
	key        *adw.EntryRow
	ca         *adw.EntryRow
	bearer     *adw.EntryRow
	exec       *adw.ActionRow
	execDelete *gtk.Button
	favourites *adw.Bin
	actions    *adw.Bin
}

func NewClusterPrefPage(parent *gtk.Window, b *behavior.Behavior, prefs observer.Property[behavior.ClusterPreferences]) *ClusterPrefPage {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	content := adw.NewBin()
	p := ClusterPrefPage{
		NavigationPage: adw.NewNavigationPage(box, "Cluster"),
		content:        content,
		behavior:       b,
		parent:         parent,
		prefs:          prefs,
	}

	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(false)
	header.PackEnd(p.createSaveButton())
	box.Append(header)
	box.Append(content)
	content.SetChild(p.createContent())

	onChange(p.prefs, func(prefs behavior.ClusterPreferences) {
		p.name.SetText(prefs.Name)
		p.host.SetText(prefs.Host)
		p.cert.SetText(string(prefs.TLS.CertData))
		p.key.SetText(string(prefs.TLS.KeyData))
		p.ca.SetText(string(prefs.TLS.CAData))
		p.bearer.SetText(string(prefs.BearerToken))
		if prefs.Exec != nil {
			p.exec.SetSubtitle(prefs.Exec.Command)
			p.execDelete.SetSensitive(true)
		} else {
			p.exec.SetSubtitle("")
			p.execDelete.SetSensitive(false)
		}

		p.favourites.SetChild(p.createFavourites())
		p.actions.SetChild(p.createActions())
	})

	return &p
}

func (p *ClusterPrefPage) createContent() *adw.PreferencesPage {
	page := adw.NewPreferencesPage()

	general := adw.NewPreferencesGroup()
	page.Add(general)
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
	p.ca = adw.NewEntryRow()
	p.ca.SetTitle("CA certificate")
	auth.AddRow(p.ca)
	p.bearer = adw.NewEntryRow()
	p.bearer.SetTitle("Bearer token")
	auth.AddRow(p.bearer)
	p.exec = adw.NewActionRow()
	p.exec.SetTitle("Exec")
	p.exec.AddCSSClass("property")
	p.execDelete = gtk.NewButton()
	p.execDelete.SetIconName("edit-delete-symbolic")
	p.execDelete.AddCSSClass("flat")
	p.execDelete.ConnectClicked(func() {
		p.exec.SetSubtitle("")
		p.execDelete.SetSensitive(false)
	})
	p.exec.AddSuffix(p.execDelete)
	auth.AddRow(p.exec)

	p.name.SetText(p.prefs.Value().Name)
	p.host.SetText(p.prefs.Value().Host)
	p.cert.SetText(string(p.prefs.Value().TLS.CertData))
	p.key.SetText(string(p.prefs.Value().TLS.KeyData))
	p.ca.SetText(string(p.prefs.Value().TLS.CAData))
	p.bearer.SetText(string(p.prefs.Value().BearerToken))
	if exec := p.prefs.Value().Exec; exec != nil {
		p.exec.SetSubtitle(exec.Command)
	} else {
		p.execDelete.SetSensitive(false)
	}

	p.favourites = adw.NewBin()
	p.favourites.SetChild(p.createFavourites())
	group := adw.NewPreferencesGroup()
	group.Add(p.favourites)
	if util.Index(p.behavior.Preferences.Value().Clusters, p.prefs) >= 0 {
		page.Add(group)
	}

	p.actions = adw.NewBin()
	p.actions.SetChild(p.createActions())
	group = adw.NewPreferencesGroup()
	group.Add(p.actions)
	page.Add(group)

	return page
}

func (p *ClusterPrefPage) createFavourites() *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()
	group.SetTitle("Favourites")
	group.SetHeaderSuffix(p.createFavouritesAddButton())

	for i, fav := range p.prefs.Value().Navigation.Favourites {
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
			v := p.prefs.Value()
			v.Navigation.Favourites = append(p.prefs.Value().Navigation.Favourites[:idx], p.prefs.Value().Navigation.Favourites[idx+1:]...)
			p.prefs.Update(v)
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

		cluster, _ := p.behavior.WithCluster(context.TODO(), p.prefs)
		for _, r := range cluster.Resources {
			res := r
			exists := false
			for _, fav := range p.prefs.Value().Navigation.Favourites {
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
				v := p.prefs.Value()
				v.Navigation.Favourites = append(p.prefs.Value().Navigation.Favourites, util.ResourceGVR(&res))
				p.prefs.Update(v)
				p.Parent().(*adw.NavigationView).Pop()
			})
			group.Add(row)
		}
	})

	return button
}

func (p *ClusterPrefPage) createSaveButton() *gtk.Button {
	button := gtk.NewButton()
	spinner := widget.NewFallbackSpinner(gtk.NewLabel("Save"))
	button.SetChild(spinner)
	button.AddCSSClass("suggested-action")
	button.ConnectClicked(func() {
		spinner.Start()
		cluster := p.prefs.Value()
		cluster.Name = p.name.Text()
		cluster.Host = p.host.Text()
		cluster.TLS.CertData = []byte(p.cert.Text())
		cluster.TLS.KeyData = []byte(p.key.Text())
		cluster.TLS.CAData = []byte(p.ca.Text())
		cluster.BearerToken = p.bearer.Text()
		if p.exec.Subtitle() == "" {
			cluster.Exec = nil
		}
		cluster.Defaults()

		if err := p.validate(cluster); err != nil {
			ShowErrorDialog(p.parent, "Validation failed", err)
			return
		}
		go func() {
			_, err := p.behavior.WithCluster(context.TODO(), observer.NewProperty(cluster))
			glib.IdleAdd(func() {
				defer spinner.Stop()
				if err != nil {
					ShowErrorDialog(p.parent, "Cluster connection failed", err)
					return
				}
				p.prefs.Update(cluster)
				if util.Index(p.behavior.Preferences.Value().Clusters, p.prefs) < 0 {
					prefs := p.behavior.Preferences.Value()
					prefs.Clusters = append(prefs.Clusters, p.prefs)
					p.behavior.Preferences.Update(prefs)
				}

				p.Parent().(*adw.NavigationView).Pop()
			})

		}()
	})

	return button
}

func (p *ClusterPrefPage) createActions() *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()

	load := adw.NewActionRow()
	load.SetActivatable(true)
	load.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
	load.SetTitle("Load kubeconfig")

	load.ConnectActivated(func() {
		fileChooser := gtk.NewFileChooserNative("Select kubeconfig", p.parent, gtk.FileChooserActionOpen, "Open", "Cancel")
		defer fileChooser.Show()
		fileChooser.ConnectResponse(func(responseId int) {
			if responseId == int(gtk.ResponseAccept) {
				p.showContextSelection(fileChooser.File().Path())
			}
		})
	})
	group.Add(load)

	if util.Index(p.behavior.Preferences.Value().Clusters, p.prefs) >= 0 {
		delete := adw.NewActionRow()
		delete.SetActivatable(true)
		delete.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
		delete.SetTitle("Delete")
		delete.AddCSSClass("error")
		delete.ConnectActivated(func() {
			dialog := adw.NewMessageDialog(p.parent, "Delete cluster?", fmt.Sprintf("Are you sure you want to delete cluster \"%s\"?", p.prefs.Value().Name))
			dialog.AddResponse("cancel", "Cancel")
			dialog.AddResponse("delete", "Delete")
			dialog.SetResponseAppearance("delete", adw.ResponseDestructive)
			dialog.Show()
			dialog.ConnectResponse(func(response string) {
				if response == "delete" {
					prefs := p.behavior.Preferences.Value()
					if i := util.Index(prefs.Clusters, p.prefs); i >= 0 {
						prefs.Clusters = append(prefs.Clusters[:i], prefs.Clusters[i+1:]...)
						p.behavior.Preferences.Update(prefs)
						p.Parent().(*adw.NavigationView).Pop()
					}
				}
			})
		})
		group.Add(delete)
	}

	return group
}

func (p *ClusterPrefPage) showContextSelection(path string) {
	rules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: path}
	apiConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, nil).
		ConfigAccess().GetStartingConfig()
	if err != nil {
		ShowErrorDialog(p.parent, "Error loading kubeconfig", err)
		return
	}

	if len(apiConfig.Contexts) == 0 {
		ShowErrorDialog(p.parent, "Error loading kubeconfig", errors.New("No contexts found."))
		return
	}

	dialog := adw.NewMessageDialog(p.parent, "Select Context", "")
	defer dialog.Show()
	dialog.AddResponse("cancel", "Cancel")
	dialog.AddResponse("confirm", "Confirm")
	dialog.SetResponseAppearance("confirm", adw.ResponseSuggested)
	box := dialog.Child().(*gtk.WindowHandle).Child().(*gtk.Box).FirstChild().(*gtk.Box)

	var group *gtk.CheckButton
	for i, context := range maps.Keys(apiConfig.Contexts) {
		radio := gtk.NewCheckButtonWithLabel(context)
		if i == 0 {
			group = radio
			radio.SetActive(true)
		} else {
			radio.SetGroup(group)
		}
		box.Append(radio)
	}

	dialog.ConnectResponse(func(response string) {
		if response == "confirm" {
			var context string
			for {
				b := group
				if b.Active() {
					context = b.Label()
					break
				}
				if s := b.NextSibling(); s != nil {
					b = s.(*gtk.CheckButton)
				} else {
					return
				}
			}

			config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(&clientcmd.ClientConfigLoadingRules{ExplicitPath: path}, &clientcmd.ConfigOverrides{CurrentContext: context}).ClientConfig()
			if err != nil {
				ShowErrorDialog(p.parent, "Error loading kubeconfig", err)
				return
			}

			newPrefs := p.prefs.Value()
			newPrefs.Name = context
			newPrefs.Host = config.Host
			newPrefs.Exec = config.ExecProvider

			if config.CertFile != "" {
				data, err := os.ReadFile(config.CertFile)
				if err != nil {
					ShowErrorDialog(p.parent, "Error loading kubeconfig", err)
					return
				}
				newPrefs.TLS.CertData = data
			} else {
				newPrefs.TLS.CertData = config.CertData
			}
			if config.KeyFile != "" {
				data, err := os.ReadFile(config.KeyFile)
				if err != nil {
					ShowErrorDialog(p.parent, "Error loading kubeconfig", err)
					return
				}
				newPrefs.TLS.KeyData = data
			} else {
				newPrefs.TLS.KeyData = config.KeyData
			}
			if config.CAFile != "" {
				data, err := os.ReadFile(config.CAFile)
				if err != nil {
					ShowErrorDialog(p.parent, "Error loading kubeconfig", err)
					return
				}
				newPrefs.TLS.CAData = data
			} else {
				newPrefs.TLS.CAData = config.CAData
			}
			if config.BearerTokenFile != "" {
				data, err := os.ReadFile(config.BearerTokenFile)
				if err != nil {
					ShowErrorDialog(p.parent, "Error loading kubeconfig", err)
					return
				}
				newPrefs.BearerToken = string(data)
			} else {
				newPrefs.BearerToken = config.BearerToken
			}

			p.prefs.Update(newPrefs)
		}
	})
}

func (p *ClusterPrefPage) validate(pref behavior.ClusterPreferences) error {
	if len(pref.Name) == 0 {
		return errors.New("Name is required")
	}

	return nil
}
