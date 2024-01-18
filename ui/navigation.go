package ui

import (
	"log"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Navigation struct {
	*gtk.ScrolledWindow
	listBox *gtk.ListBox
}

func NewNavigation() *Navigation {
	lb := gtk.NewListBox()
	lb.ConnectRowSelected(func(row *gtk.ListBoxRow) {
		log.Printf("%+v", row.Name())
	})
	lb.AddCSSClass("navigation-sidebar")
	lb.SetVExpand(true)

	sw := gtk.NewScrolledWindow()
	sw.SetSizeRequest(250, 100)

	sw.SetVExpand(true)
	vp := gtk.NewViewport(nil, nil)
	vp.SetVExpand(true)
	vp.SetChild(lb)
	sw.SetChild(vp)

	navigation := &Navigation{
		ScrolledWindow: sw,
		listBox:        lb,
	}
	navigation.Refresh()

	return navigation

}

func (n *Navigation) Refresh() {
	// for {
	// 	child := n.listBox.FirstChild()
	// 	if child == nil {
	// 		break
	// 	}
	// 	n.listBox.Remove(child)
	// }

	for _, resource := range application.config.Navigation.Favourites {
		row := gtk.NewListBoxRow()
		row.SetName(resource.Kind)
		box := gtk.NewBox(gtk.OrientationHorizontal, 8)
		box.Append(n.kindIcon(resource))
		label := gtk.NewLabel(resource.Kind)
		box.Append(label)
		row.SetChild(box)
		n.listBox.Append(row)
	}
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
