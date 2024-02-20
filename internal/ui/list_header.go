package ui

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/internal/behavior"
	"github.com/getseabird/seabird/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ListHeader struct {
	*gtk.Box
}

func NewListHeader(b *behavior.ListBehavior) *ListHeader {
	box := gtk.NewBox(gtk.OrientationHorizontal, 0)
	box.AddCSSClass("linked")
	box.SetMarginStart(12)
	box.SetMarginEnd(12)

	// TODO expression triggers G_IS_OBJECT (object) assertion fails
	kind := gtk.NewDropDown(gtk.NewStringList([]string{}), gtk.NewPropertyExpression(gtk.GTypeStringObject, nil, "string"))
	kind.SetEnableSearch(true)

	for _, r := range b.Resources {
		kind.Model().Cast().(*gtk.StringList).Append(r.Kind)
	}
	kind.Connect("notify::selected-item", func() {
		res := b.Resources[kind.Selected()]
		if !util.ResourceEquals(b.SelectedResource.Value(), &res) {
			b.SelectedResource.Update(&res)
		}
	})
	box.Append(kind)

	entry := gtk.NewSearchEntry()
	entry.SetHExpand(true)
	box.Append(entry)
	entry.ConnectChanged(func() {
		if entry.Text() != b.SearchText.Value() {
			b.SearchText.Update(entry.Text())
		}
	})
	onChange(b.SearchText, func(txt string) {
		if txt != entry.Text() {
			entry.SetText(txt)
		}
	})

	button := gtk.NewMenuButton()
	button.SetIconName("funnel-symbolic")
	box.Append(button)

	namespace := gio.NewMenu()
	for _, ns := range b.Namespaces.Value() {
		namespace.Append(ns.GetName(), fmt.Sprintf("list.filterNamespace('%s')", ns.GetName()))
	}
	model := gio.NewMenu()
	model.AppendSection("Namespace", namespace)
	popover := gtk.NewPopoverMenuFromModel(model)
	button.SetPopover(popover)

	entry.ConnectSearchChanged(func() {
		b.SearchFilter.Update(behavior.NewSearchFilter(entry.Text()))
	})

	onChange(b.SelectedResource, func(res *metav1.APIResource) {
		var idx uint
		for i, r := range b.Resources {
			if util.ResourceEquals(&r, res) {
				idx = uint(i)
				break
			}
		}
		kind.SetSelected(idx)
	})

	return &ListHeader{box}
}
