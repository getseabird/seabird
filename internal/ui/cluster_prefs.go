package ui

import (
	"context"
	"errors"
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/behavior"
	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/util"
	"github.com/getseabird/seabird/widget"
	"github.com/imkira/go-observer/v2"
	"golang.org/x/exp/maps"
	"k8s.io/client-go/tools/clientcmd"
)

type ClusterPrefPage struct {
	*adw.NavigationPage
	ctx        context.Context
	content    *adw.Bin
	behavior   *behavior.Behavior
	prefs      observer.Property[api.ClusterPreferences]
	name       *adw.EntryRow
	host       *adw.EntryRow
	cert       *adw.EntryRow
	key        *adw.EntryRow
	ca         *adw.EntryRow
	bearer     *adw.EntryRow
	exec       *adw.ActionRow
	readonly   *adw.SwitchRow
	execDelete *gtk.Button
	favourites *adw.Bin
	actions    *adw.Bin
}

func NewClusterPrefPage(ctx context.Context, b *behavior.Behavior, prefs observer.Property[api.ClusterPreferences]) *ClusterPrefPage {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	content := adw.NewBin()
	p := ClusterPrefPage{
		ctx:            ctx,
		NavigationPage: adw.NewNavigationPage(box, "Cluster"),
		content:        content,
		behavior:       b,
		prefs:          prefs,
	}

	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(false)
	header.PackEnd(p.createSaveButton())
	box.Append(header)
	box.Append(content)
	content.SetChild(p.createContent())

	common.OnChange(ctx, p.prefs, func(prefs api.ClusterPreferences) {
		p.updateValues(prefs)
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
	p.readonly = adw.NewSwitchRow()
	p.readonly.SetTitle("Read-only")
	general.Add(p.readonly)

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

	p.updateValues(p.prefs.Value())

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

		cluster, _ := p.behavior.WithCluster(p.ctx, p.prefs)
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
		cluster.ReadOnly = p.readonly.Active()
		cluster.TLS.CertData = []byte(p.cert.Text())
		cluster.TLS.KeyData = []byte(p.key.Text())
		cluster.TLS.CAData = []byte(p.ca.Text())
		cluster.BearerToken = p.bearer.Text()
		if p.exec.Subtitle() == "" {
			cluster.Exec = nil
		}
		cluster.Defaults()

		if showClusterPrefsErrorDialog(p.ctx, cluster) {
			spinner.Stop()
			return
		}

		go func() {
			_, err := p.behavior.WithCluster(p.ctx, observer.NewProperty(cluster))
			glib.IdleAdd(func() {
				defer spinner.Stop()
				if err != nil {
					widget.ShowErrorDialog(p.ctx, "Cluster connection failed", err)
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
	load.SetSensitive(p.prefs.Value().Kubeconfig == nil)
	load.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
	load.SetTitle("Load kubeconfig")

	load.ConnectActivated(func() {
		fileChooser := gtk.NewFileChooserNative("Select kubeconfig", ctxt.MustFrom[*gtk.Window](p.ctx), gtk.FileChooserActionOpen, "Open", "Cancel")
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
		delete.SetSensitive(p.prefs.Value().Kubeconfig == nil)
		delete.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
		delete.SetTitle("Delete")
		delete.AddCSSClass("error")
		delete.ConnectActivated(func() {
			dialog := adw.NewMessageDialog(ctxt.MustFrom[*gtk.Window](p.ctx), "Delete cluster?", fmt.Sprintf("Are you sure you want to delete cluster \"%s\"?", p.prefs.Value().Name))
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
		widget.ShowErrorDialog(p.ctx, "Error loading kubeconfig", err)
		return
	}

	if len(apiConfig.Contexts) == 0 {
		widget.ShowErrorDialog(p.ctx, "Error loading kubeconfig", errors.New("no contexts found"))
		return
	}

	dialog := adw.NewMessageDialog(ctxt.MustFrom[*gtk.Window](p.ctx), "Select Context", "")
	defer dialog.Show()
	dialog.AddResponse("cancel", "Cancel")
	dialog.AddResponse("confirm", "Confirm")
	dialog.SetResponseAppearance("confirm", adw.ResponseSuggested)
	root := dialog.Child().(*gtk.WindowHandle).Child().(*gtk.Box).FirstChild().(*gtk.Box)
	sw := gtk.NewScrolledWindow()
	sw.SetMinContentHeight(100)
	root.Append(sw)
	box := gtk.NewBox(gtk.OrientationVertical, 1)
	box.SetVAlign(gtk.AlignCenter)
	sw.SetChild(box)

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
		switch response {
		case "confirm":
			var context string
			button := group
			for {
				if button.Active() {
					context = button.Label()
					break
				}
				button = button.NextSibling().(*gtk.CheckButton)
			}
			prefs := p.prefs.Value()
			if err := api.UpdateClusterPreferences(&prefs, path, context); err != nil {
				widget.ShowErrorDialog(p.ctx, "Error loading kubeconfig", err)
				return
			}
			p.prefs.Update(prefs)
		}
	})
}

func (p *ClusterPrefPage) updateValues(prefs api.ClusterPreferences) {
	p.name.SetText(prefs.Name)
	p.host.SetText(prefs.Host)
	p.readonly.SetActive(prefs.ReadOnly)
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

	if prefs.Kubeconfig != nil {
		p.name.SetSensitive(false)
		p.host.SetSensitive(false)
		p.cert.SetSensitive(false)
		p.key.SetSensitive(false)
		p.ca.SetSensitive(false)
		p.bearer.SetSensitive(false)
		p.execDelete.SetSensitive(false)
	}
}
