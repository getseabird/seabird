package ui

import (
	"context"
	"errors"
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/getseabird/seabird/internal/pubsub"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/util"
	"github.com/getseabird/seabird/widget"
	"golang.org/x/exp/maps"
	"k8s.io/client-go/tools/clientcmd"
)

type ClusterPrefPage struct {
	*adw.NavigationPage
	*common.State
	ctx        context.Context
	content    *adw.Bin
	prefs      pubsub.Property[api.ClusterPreferences]
	name       *adw.EntryRow
	host       *adw.EntryRow
	cert       *adw.EntryRow
	key        *adw.EntryRow
	ca         *adw.EntryRow
	bearer     *adw.EntryRow
	exec       *adw.ActionRow
	readonly   *adw.SwitchRow
	execDelete *gtk.Button
	actions    *adw.Bin
}

func NewClusterPrefPage(ctx context.Context, state *common.State, prefs pubsub.Property[api.ClusterPreferences]) *ClusterPrefPage {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	content := adw.NewBin()
	p := ClusterPrefPage{
		ctx:            ctx,
		State:          state,
		NavigationPage: adw.NewNavigationPage(box, "Cluster"),
		content:        content,
		prefs:          prefs,
	}

	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(false)
	header.PackEnd(p.createSaveButton())
	box.Append(header)
	box.Append(content)
	content.SetChild(p.createContent())

	p.prefs.Sub(ctx, func(prefs api.ClusterPreferences) {
		p.updateValues(prefs)
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
	p.execDelete.SetIconName("user-trash-symbolic")
	p.execDelete.AddCSSClass("flat")
	p.execDelete.ConnectClicked(func() {
		p.exec.SetSubtitle("")
		p.execDelete.SetSensitive(false)
	})
	p.exec.AddSuffix(p.execDelete)
	auth.AddRow(p.exec)

	p.updateValues(p.prefs.Value())

	p.actions = adw.NewBin()
	p.actions.SetChild(p.createActions())
	group := adw.NewPreferencesGroup()
	group.Add(p.actions)
	page.Add(group)

	return page
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
			_, err := p.NewClusterState(p.ctx, pubsub.NewProperty(cluster))
			glib.IdleAdd(func() {
				defer spinner.Stop()
				if err != nil {
					widget.ShowErrorDialog(p.ctx, "Cluster connection failed", err)
					return
				}
				p.prefs.Pub(cluster)
				if util.Index(p.Preferences.Value().Clusters, p.prefs) < 0 {
					prefs := p.Preferences.Value()
					prefs.Clusters = append(prefs.Clusters, p.prefs)
					p.Preferences.Pub(prefs)
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
	load.SetTitle("Import kubeconfig")

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

	if util.Index(p.Preferences.Value().Clusters, p.prefs) >= 0 {
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
			dialog.Present()
			dialog.ConnectResponse(func(response string) {
				if response == "delete" {
					prefs := p.Preferences.Value()
					if i := util.Index(prefs.Clusters, p.prefs); i >= 0 {
						prefs.Clusters = append(prefs.Clusters[:i], prefs.Clusters[i+1:]...)
						p.Preferences.Pub(prefs)
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
	defer dialog.Present()
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
			p.prefs.Pub(prefs)
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
