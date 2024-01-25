package ui

import (
	"encoding/json"
	"log"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/behavior"
	"github.com/getseabird/seabird/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
)

type Navigation struct {
	*adw.ToolbarView
	behavior *behavior.ClusterBehavior
	list     *gtk.ListBox
	rows     []*gtk.ListBoxRow
}

func NewNavigation(b *behavior.ClusterBehavior) *Navigation {
	n := &Navigation{ToolbarView: adw.NewToolbarView(), behavior: b}
	n.SetSizeRequest(200, 200)
	n.SetVExpand(true)

	header := adw.NewHeaderBar()
	title := gtk.NewLabel(b.ClusterPreferences.Value().Name)
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
	n.SetContent(n.createFavourites(b.ClusterPreferences.Value()))

	onChange(b.ClusterPreferences, func(prefs behavior.ClusterPreferences) {
		n.SetContent(n.createFavourites(prefs))
	})

	onChange(b.SelectedResource, func(res *metav1.APIResource) {
		var idx *int
		for i, r := range b.ClusterPreferences.Value().Navigation.Favourites {
			if util.ResourceGVR(res).String() == r.String() {
				idx = ptr.To(i)
				break
			}
		}
		if idx != nil {
			n.list.SelectRow(n.rows[*idx])
		} else {
			n.list.SelectRow(nil)
		}
	})

	return n
}

func (n *Navigation) createFavourites(prefs behavior.ClusterPreferences) *gtk.ListBox {
	n.list = gtk.NewListBox()
	n.list.AddCSSClass("dim-label")
	n.list.AddCSSClass("navigation-sidebar")
	n.list.SetVExpand(true)
	n.list.ConnectRowSelected(func(row *gtk.ListBoxRow) {
		if row == nil {
			return
		}
		var gvr schema.GroupVersionResource
		if err := json.Unmarshal([]byte(row.Name()), &gvr); err != nil {
			log.Printf("failed to unmarshal gvr: %v", err)
			return
		}

		for _, res := range n.behavior.Resources {
			if util.ResourceGVR(&res).String() == gvr.String() {
				n.behavior.SelectedResource.Update(&res)
				break
			}
		}

	})

	n.rows = nil

	for i, gvr := range prefs.Navigation.Favourites {
		var resource *v1.APIResource
		for _, r := range n.behavior.Resources {
			if r.Group == gvr.Group && r.Version == gvr.Version && r.Name == gvr.Resource {
				resource = &r
				break
			}
		}
		if resource == nil {
			log.Printf("ignoring unknown resource %s", gvr.String())
			n.rows = append(n.rows, nil)
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
		n.list.Append(row)
		n.rows = append(n.rows, row)

		if i == 0 {
			n.list.SelectRow(row)
		}
	}

	return n.list
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
