package ui

import (
	"log"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
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

	for _, resource := range application.cluster.Resources {
		row := gtk.NewListBoxRow()
		row.SetName(resource.Kind)
		box := gtk.NewBox(gtk.OrientationHorizontal, 8)
		img := gtk.NewImageFromIconName("applications-system-symbolic")
		img.SetIconSize(gtk.IconSizeNormal)
		box.Append(img)
		label := gtk.NewLabel(resource.Kind)
		box.Append(label)
		row.SetChild(box)
		n.listBox.Append(row)
	}
}
