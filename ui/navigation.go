package ui

import (
	"encoding/json"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Navigation struct {
	*adw.ToolbarView
}

func NewNavigation() *Navigation {
	n := &Navigation{
		ToolbarView: adw.NewToolbarView(),
	}

	action := gio.NewSimpleAction("preferences", nil)
	action.ConnectActivate(func(parameter *glib.Variant) {
		w := NewPreferencesWindow()
		w.SetTransientFor(&application.window.Window)
		w.Show()
	})
	application.AddAction(action)

	action = gio.NewSimpleAction("about", nil)
	action.ConnectActivate(func(parameter *glib.Variant) {
		w := adw.NewAboutWindow()
		w.SetApplicationName("kubegtk")
		w.SetTransientFor(&application.window.Window)
		w.Show()
	})
	application.AddAction(action)

	header := adw.NewHeaderBar()
	header.SetShowTitle(false)
	header.SetShowEndTitleButtons(false)
	header.SetShowStartTitleButtons(false)
	prefBtn := gtk.NewMenuButton()
	prefBtn.SetIconName("open-menu-symbolic")
	menu := gio.NewMenu()
	menu.Append("Preferences", "app.preferences")
	menu.Append("Keyboard Shortcuts", "app.shortcuts")
	menu.Append("About", "app.about")
	popover := gtk.NewPopoverMenuFromModel(menu)
	prefBtn.SetPopover(popover)

	header.PackEnd(prefBtn)
	n.AddTopBar(header)

	cb := gtk.NewComboBoxText()
	for _, cluster := range application.prefs.Clusters {
		cb.AppendText(cluster.Name)
	}
	cb.SetActive(0)
	header.PackStart(cb)

	n.SetSizeRequest(250, 250)
	n.SetVExpand(true)
	n.SetContent(n.createFavourites())

	return n
}

func (n *Navigation) createFavourites() *gtk.ListBox {
	listBox := gtk.NewListBox()
	listBox.ConnectRowSelected(func(row *gtk.ListBoxRow) {
		var gvr schema.GroupVersionResource
		if err := json.Unmarshal([]byte(row.Name()), &gvr); err != nil {
			panic(err)
		}
		application.listView.SetResource(gvr)
	})

	listBox.AddCSSClass("navigation-sidebar")
	listBox.SetVExpand(true)

	for i, gvr := range application.cluster.Preferences.Navigation.Favourites {
		var resource *v1.APIResource
		for _, r := range application.cluster.Resources {
			if r.Group == gvr.Group && r.Version == gvr.Version && r.Name == gvr.Resource {
				resource = &r
				break
			}
		}

		row := gtk.NewListBoxRow()
		json, err := json.Marshal(gvr)
		if err != nil {
			panic(err)
		}
		row.SetName(string(json))
		box := gtk.NewBox(gtk.OrientationHorizontal, 8)
		box.Append(n.resIcon(gvr))
		label := gtk.NewLabel(resource.Kind)
		box.Append(label)
		row.SetChild(box)
		listBox.Append(row)

		if i == 0 {
			listBox.SelectRow(row)
		}
	}

	return listBox
}

func (n *Navigation) resIcon(gvk schema.GroupVersionResource) *gtk.Image {
	switch gvk.Group {
	case corev1.GroupName:
		{
			switch gvk.Resource {
			case "pods":
				return gtk.NewImageFromIconName("application-x-executable-symbolic")
			case "configmaps":
				return gtk.NewImageFromIconName("preferences-system-symbolic")
			case "secrets":
				return gtk.NewImageFromIconName("channel-secure-symbolic")
			case "namespaces":
				return gtk.NewImageFromIconName("application-rss+xml-symbolic")
			}
		}
	case appsv1.GroupName:
		switch gvk.Resource {
		case "deployments":
			return gtk.NewImageFromIconName("preferences-system-network-symbolic")
		case "statefulsets":
			return gtk.NewImageFromIconName("drive-harddisk-symbolic")
		}
	}

	return gtk.NewImageFromIconName("application-x-addon-symbolic")
}
