package ui

import (
	"context"
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/style"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/ui/editor"
	"github.com/getseabird/seabird/internal/util"
	"github.com/getseabird/seabird/widget"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ListHeader struct {
	*adw.HeaderBar
	*common.ClusterState
}

func NewListHeader(ctx context.Context, state *common.ClusterState, breakpoint *adw.Breakpoint, showSidebar func(), editor *editor.EditorWindow) *ListHeader {
	header := adw.NewHeaderBar()
	header.AddCSSClass("flat")
	header.SetShowEndTitleButtons(false)
	header.SetShowStartTitleButtons(false)
	switch style.Get() {
	case style.Darwin:
		breakpoint.AddSetter(header, "show-start-title-buttons", true)
	}

	sidebarButton := gtk.NewButton()
	sidebarButton.SetIconName("sidebar-show-symbolic")
	sidebarButton.SetVisible(false)
	sidebarButton.ConnectClicked(showSidebar)
	sidebarButton.SetTooltipText("Show Sidebar")
	header.PackStart(sidebarButton)
	breakpoint.AddSetter(sidebarButton, "visible", true)

	createButton := gtk.NewButton()
	createButton.SetIconName("document-new-symbolic")
	createButton.SetTooltipText("New Resource")
	createButton.ConnectClicked(func() {
		gvk := util.ResourceGVK(state.SelectedResource.Value())
		err := editor.AddPage(&gvk, nil)
		if err != nil {
			widget.ShowErrorDialog(ctx, "Error loading editor", err)
			return
		}
		editor.Present()
	})
	header.PackEnd(createButton)

	box := gtk.NewBox(gtk.OrientationHorizontal, 0)
	box.AddCSSClass("linked")
	header.SetTitleWidget(box)

	// TODO expression triggers G_IS_OBJECT (object) assertion fails
	kind := gtk.NewDropDown(gtk.NewStringList([]string{}), gtk.NewPropertyExpression(gtk.GTypeStringObject, nil, "string"))
	kind.SetEnableSearch(true)
	kind.AddCSSClass("kind-dropdown")
	factory := gtk.NewSignalListItemFactory()
	factory.ConnectSetup(func(listitem *gtk.ListItem) {
		box := gtk.NewBox(gtk.OrientationVertical, 0)
		label := gtk.NewLabel("")
		label.AddCSSClass("caption-heading")
		label.SetHAlign(gtk.AlignStart)
		box.Append(label)
		label = gtk.NewLabel("")
		label.AddCSSClass("caption")
		label.AddCSSClass("dim-label")
		label.SetHAlign(gtk.AlignStart)
		label.SetEllipsize(pango.EllipsizeEnd)
		box.Append(label)
		listitem.SetChild(box)
	})
	factory.ConnectBind(func(listitem *gtk.ListItem) {
		str := listitem.Item().Cast().(*gtk.StringObject).String()
		gk := schema.ParseGroupKind(str)
		label := listitem.Child().(*gtk.Box).FirstChild().(*gtk.Label)
		label.SetText(gk.Kind)
		if gk.Group == "" {
			gk.Group = "k8s.io"
		}
		label.NextSibling().(*gtk.Label).SetText(gk.Group)
	})
	factory.ConnectTeardown(func(listitem *gtk.ListItem) {
		listitem.SetChild(nil)
	})
	kind.SetFactory(&factory.ListItemFactory)

	for _, r := range state.Resources {
		kind.Model().Cast().(*gtk.StringList).Append(schema.GroupKind{Group: r.Group, Kind: r.Kind}.String())
	}
	kind.Connect("notify::selected-item", func() {
		res := state.Resources[kind.Selected()]
		if !util.ResourceEquals(state.SelectedResource.Value(), &res) {
			state.SelectedResource.Update(&res)
		}
	})
	box.Append(kind)

	entry := gtk.NewSearchEntry()
	entry.SetObjectProperty("placeholder-text", state.ClusterPreferences.Value().Name)
	placeholder := entry.FirstChild().(*gtk.Image).NextSibling().(*gtk.Text).FirstChild().(*gtk.Label)
	placeholder.AddCSSClass("heading")
	entry.SetObjectProperty("placeholder-text", "")
	breakpoint.AddSetter(entry, "placeholder-text", state.ClusterPreferences.Value().Name)

	entry.SetMaxWidthChars(50)
	box.Append(entry)
	entry.ConnectChanged(func() {
		if entry.Text() != state.SearchText.Value() {
			state.SearchText.Update(entry.Text())
		}
	})
	common.OnChange(ctx, state.SearchText, func(txt string) {
		if txt != entry.Text() {
			entry.SetText(txt)
		}
	})

	filterButton := gtk.NewMenuButton()
	filterButton.SetIconName("funnel-symbolic")
	filterButton.SetTooltipText("Filter")
	box.Append(filterButton)
	namespace := gio.NewMenu()
	common.OnChange(ctx, state.Namespaces, func(ns []*corev1.Namespace) {
		namespace.RemoveAll()
		for _, ns := range ns {
			namespace.Append(ns.GetName(), fmt.Sprintf("list.filterNamespace('%s')", ns.GetName()))
		}
	})
	model := gio.NewMenu()
	model.AppendSection("Namespace", namespace)
	popover := gtk.NewPopoverMenuFromModel(model)
	filterButton.SetPopover(popover)

	entry.ConnectSearchChanged(func() {
		state.SearchFilter.Update(common.NewSearchFilter(entry.Text()))
	})

	common.OnChange(ctx, state.SelectedResource, func(res *metav1.APIResource) {
		var idx uint
		for i, r := range state.Resources {
			if util.ResourceEquals(&r, res) {
				idx = uint(i)
				break
			}
		}
		kind.SetSelected(idx)
	})

	common.OnChange(ctx, state.ClusterPreferences, func(prefs api.ClusterPreferences) {
		createButton.SetVisible(!prefs.ReadOnly)
	})

	return &ListHeader{HeaderBar: header, ClusterState: state}
}
