package ui

import (
	"context"
	"fmt"
	"runtime"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/internal/behavior"
	"github.com/getseabird/seabird/internal/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ListHeader struct {
	*adw.HeaderBar
}

func NewListHeader(ctx context.Context, b *behavior.ListBehavior, breakpoint *adw.Breakpoint, showSidebar func()) *ListHeader {
	header := adw.NewHeaderBar()
	header.AddCSSClass("flat")
	header.SetShowEndTitleButtons(false)
	header.SetShowStartTitleButtons(false)
	switch runtime.GOOS {
	case "darwin":
		breakpoint.AddSetter(header, "show-start-title-buttons", true)
	}

	btn := gtk.NewButton()
	btn.SetIconName("sidebar-show-symbolic")
	btn.SetVisible(false)
	btn.ConnectClicked(showSidebar)
	header.PackStart(btn)
	breakpoint.AddSetter(btn, "visible", true)

	box := gtk.NewBox(gtk.OrientationHorizontal, 0)
	box.AddCSSClass("linked")
	header.SetTitleWidget(box)

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
	entry.SetMaxWidthChars(50)
	box.Append(entry)
	entry.ConnectChanged(func() {
		if entry.Text() != b.SearchText.Value() {
			b.SearchText.Update(entry.Text())
		}
	})
	onChange(ctx, b.SearchText, func(txt string) {
		if txt != entry.Text() {
			entry.SetText(txt)
		}
	})

	button := gtk.NewMenuButton()
	button.SetIconName("funnel-symbolic")
	box.Append(button)

	namespace := gio.NewMenu()
	onChange(ctx, b.Namespaces, func(ns []*corev1.Namespace) {
		namespace.RemoveAll()
		for _, ns := range ns {
			namespace.Append(ns.GetName(), fmt.Sprintf("list.filterNamespace('%s')", ns.GetName()))
		}
	})
	model := gio.NewMenu()
	model.AppendSection("Namespace", namespace)
	popover := gtk.NewPopoverMenuFromModel(model)
	button.SetPopover(popover)

	entry.ConnectSearchChanged(func() {
		b.SearchFilter.Update(behavior.NewSearchFilter(entry.Text()))
	})

	onChange(ctx, b.SelectedResource, func(res *metav1.APIResource) {
		var idx uint
		for i, r := range b.Resources {
			if util.ResourceEquals(&r, res) {
				idx = uint(i)
				break
			}
		}
		kind.SetSelected(idx)
	})

	return &ListHeader{header}
}
