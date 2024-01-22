package ui

import (
	"encoding/json"
	"log"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/jgillich/kubegio/state"
	"github.com/kelindar/event"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Navigation struct {
	*adw.ToolbarView
	root *ClusterWindow
}

func NewNavigation(root *ClusterWindow) *Navigation {
	n := &Navigation{ToolbarView: adw.NewToolbarView(), root: root}
	n.SetSizeRequest(250, 250)
	n.SetVExpand(true)

	header := adw.NewHeaderBar()
	title := gtk.NewLabel(root.cluster.Preferences.Name)
	title.AddCSSClass("heading")
	header.SetTitleWidget(title)
	header.SetShowEndTitleButtons(false)
	header.SetShowStartTitleButtons(false)

	button := gtk.NewMenuButton()
	button.SetIconName("open-menu-symbolic")

	windowSection := gio.NewMenu()
	windowSection.Append("New Window", "win.newWindow")
	windowSection.Append("Disconnect", "win.disconnect")

	prefSection := gio.NewMenu()
	prefSection.Append("Preferences", "win.prefs")
	// prefSection.Append("Keyboard Shortcuts", "win.shortcuts")
	prefSection.Append("About", "win.about")

	m := gio.NewMenu()
	m.AppendSection("", windowSection)
	m.AppendSection("", prefSection)

	popover := gtk.NewPopoverMenuFromModel(m)
	button.SetPopover(popover)

	header.PackEnd(button)
	n.AddTopBar(header)
	n.SetContent(n.createFavourites())

	event.On(func(ev state.PreferencesUpdated) {
		glib.IdleAdd(func() {
			n.SetContent(n.createFavourites())
		})
	})

	return n
}

func (n *Navigation) createFavourites() *gtk.ListBox {
	listBox := gtk.NewListBox()
	listBox.ConnectRowSelected(func(row *gtk.ListBoxRow) {
		var gvr schema.GroupVersionResource
		if err := json.Unmarshal([]byte(row.Name()), &gvr); err != nil {
			panic(err)
		}
		n.root.listView.SetResource(gvr)
	})

	listBox.AddCSSClass("dim-label")
	listBox.AddCSSClass("navigation-sidebar")
	listBox.SetVExpand(true)

	for i, gvr := range n.root.cluster.Preferences.Navigation.Favourites {
		var resource *v1.APIResource
		for _, r := range n.root.cluster.Resources {
			if r.Group == gvr.Group && r.Version == gvr.Version && r.Name == gvr.Resource {
				resource = &r
				break
			}
		}
		if resource == nil {
			log.Printf("ignoring unknown resource %s", gvr.String())
			continue
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
