package list

import (
	"context"
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/ui/editor"
	"github.com/getseabird/seabird/internal/util"
	"github.com/getseabird/seabird/widget"
	corev1 "k8s.io/api/core/v1"
)

type ListHeader struct {
	*adw.HeaderBar
	*common.ClusterState
}

func newListHeader(ctx context.Context, state *common.ClusterState, editor *editor.EditorWindow) *ListHeader {
	header := adw.NewHeaderBar()
	header.AddCSSClass("flat")
	header.SetShowStartTitleButtons(false)

	createButton := gtk.NewButton()
	createButton.SetIconName("text-editor-symbolic")
	createButton.SetTooltipText("New Resource")
	createButton.ConnectClicked(func() {
		gvk := util.GVKForResource(state.SelectedResource.Value())
		err := editor.AddPage(&gvk, nil)
		if err != nil {
			widget.ShowErrorDialog(ctx, "Error loading editor", err)
			return
		}
		editor.Present()
	})
	header.PackStart(createButton)

	box := gtk.NewBox(gtk.OrientationHorizontal, 0)
	box.AddCSSClass("linked")
	box.SetMarginStart(32)
	box.SetMarginEnd(32)
	header.SetTitleWidget(box)

	entry := gtk.NewSearchEntry()
	entry.SetMaxWidthChars(75)
	entry.SetObjectProperty("placeholder-text", "Search")
	box.Append(entry)
	entry.ConnectSearchChanged(func() {
		if entry.Text() != state.SearchText.Value() {
			state.SearchText.Pub(entry.Text())
		}
	})
	state.SearchText.Sub(ctx, func(txt string) {
		if txt != entry.Text() {
			entry.SetText(txt)
		}
	})

	filterButton := gtk.NewMenuButton()
	filterButton.SetIconName("funnel-symbolic")
	filterButton.SetTooltipText("Filter")
	box.Append(filterButton)
	namespace := gio.NewMenu()
	state.Namespaces.Sub(ctx, func(ns []*corev1.Namespace) {
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
		state.SearchFilter.Pub(common.NewSearchFilter(entry.Text()))
	})

	state.ClusterPreferences.Sub(ctx, func(prefs api.ClusterPreferences) {
		createButton.SetVisible(!prefs.ReadOnly)
	})

	return &ListHeader{HeaderBar: header, ClusterState: state}
}
