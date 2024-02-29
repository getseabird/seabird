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
	"github.com/getseabird/seabird/widget"
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

	sidebarButton := gtk.NewButton()
	sidebarButton.SetIconName("sidebar-show-symbolic")
	sidebarButton.SetVisible(false)
	sidebarButton.ConnectClicked(showSidebar)
	header.PackStart(sidebarButton)
	breakpoint.AddSetter(sidebarButton, "visible", true)

	createButton := gtk.NewButton()
	createButton.SetIconName("document-new-symbolic")
	createButton.ConnectClicked(func() {
		w, err := NewEditorWindow(ctx, b.SelectedResource.Value(), nil)
		if err != nil {
			widget.ShowErrorDialog(ctx, "Error loading editor", err)
			return
		}
		w.Show()
	})
	header.PackEnd(createButton)

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
	entry.SetObjectProperty("placeholder-text", b.Cluster.ClusterPreferences.Value().Name)
	placeholder := entry.FirstChild().(*gtk.Image).NextSibling().(*gtk.Text).FirstChild().(*gtk.Label)
	placeholder.AddCSSClass("heading")
	entry.SetObjectProperty("placeholder-text", "")
	breakpoint.AddSetter(entry, "placeholder-text", b.Cluster.ClusterPreferences.Value().Name)

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
