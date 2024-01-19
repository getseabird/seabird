package ui

import (
	"log"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Navigation struct {
	*adw.ToolbarView
}

func NewNavigation() *Navigation {
	n := &Navigation{
		ToolbarView: adw.NewToolbarView(),
	}

	header := adw.NewHeaderBar()
	header.SetShowTitle(false)
	header.SetShowEndTitleButtons(false)
	header.SetShowStartTitleButtons(false)
	prefBtn := gtk.NewButton()
	prefBtn.SetIconName("open-menu-symbolic")
	prefBtn.ConnectClicked(func() {
		w := NewPreferencesWindow()
		w.SetTransientFor(&application.window.Window)
		w.Show()
	})
	header.PackEnd(prefBtn)
	n.AddTopBar(header)

	cb := gtk.NewComboBoxText()
	for _, cluster := range application.prefs.Clusters {
		cb.AppendText(cluster.Name)
	}
	cb.SetActive(0)
	header.PackStart(cb)

	n.SetSizeRequest(250, 100)
	n.SetVExpand(true)
	n.SetContent(n.favourites())

	return n
}

func (n *Navigation) favourites() *gtk.ListBox {
	listBox := gtk.NewListBox()
	listBox.ConnectRowSelected(func(row *gtk.ListBoxRow) {
		log.Printf("%+v", row.Name())
	})
	listBox.AddCSSClass("navigation-sidebar")
	listBox.SetVExpand(true)

	for _, resource := range application.cluster.Preferences.Navigation.Favourites {
		row := gtk.NewListBoxRow()
		row.SetName(resource.Kind)
		box := gtk.NewBox(gtk.OrientationHorizontal, 8)
		box.Append(n.kindIcon(resource))
		label := gtk.NewLabel(resource.Kind)
		box.Append(label)
		row.SetChild(box)
		listBox.Append(row)
	}

	return listBox
}

func (n *Navigation) kindIcon(gvk schema.GroupVersionKind) *gtk.Image {
	switch gvk.Group {
	case "":
		{
			switch gvk.Kind {
			case "Pod":
				return gtk.NewImageFromIconName("application-x-executable-symbolic")
			case "ConfigMap":
				return gtk.NewImageFromIconName("preferences-system-symbolic")
			case "Secret":
				return gtk.NewImageFromIconName("channel-secure-symbolic")
			case "Namespace":
				return gtk.NewImageFromIconName("application-rss+xml-symbolic")
			}
		}
	case "apps":
		switch gvk.Kind {
		case "Deployment":
			return gtk.NewImageFromIconName("preferences-system-network-symbolic")
		case "StatefulSet":
			return gtk.NewImageFromIconName("drive-harddisk-symbolic")
		}
	}

	return gtk.NewImageFromIconName("image-missing-symbolic")

}
