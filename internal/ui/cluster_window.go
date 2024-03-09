package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/ui/editor"
	"github.com/getseabird/seabird/widget"
)

type ClusterWindow struct {
	*widget.UniversalApplicationWindow
	*common.ClusterState
	ctx          context.Context
	cancel       context.CancelFunc
	navigation   *Navigation
	listView     *ListView
	detailView   *DetailView
	toastOverlay *adw.ToastOverlay
}

func NewClusterWindow(ctx context.Context, app *gtk.Application, state *common.ClusterState) *ClusterWindow {
	window := widget.NewUniversalApplicationWindow(app)
	ctx = ctxt.With[*gtk.Window](ctx, &window.Window)
	ctx = ctxt.With[*api.Cluster](ctx, state.Cluster)
	ctx, cancel := context.WithCancel(ctx)
	w := ClusterWindow{
		ClusterState:               state,
		ctx:                        ctx,
		UniversalApplicationWindow: window,
		cancel:                     cancel,
	}
	w.SetIconName("seabird")
	w.SetTitle(fmt.Sprintf("%s - %s", w.ClusterPreferences.Value().Name, ApplicationName))
	w.SetDefaultSize(1280, 720)

	var h glib.SignalHandle
	h = w.ConnectCloseRequest(func() bool {
		prefs := w.Preferences.Value()
		if err := prefs.Save(); err != nil {
			d := widget.ShowErrorDialog(ctx, "Could not save preferences", err)
			d.ConnectUnrealize(func() {
				w.Close()
			})
			w.HandlerDisconnect(h)
			return true
		}
		return false
	})

	editor := editor.NewEditorWindow(ctx)

	breakpointBin := adw.NewBreakpointBin()
	breakpointBin.SetSizeRequest(800, 500)
	w.toastOverlay = adw.NewToastOverlay()
	breakpointBin.SetChild(w.toastOverlay)
	w.SetContent(breakpointBin)

	splitView := adw.NewOverlaySplitView()
	splitView.SetEnableHideGesture(true)
	splitView.SetEnableShowGesture(true)
	splitView.SetMinSidebarWidth(150)
	splitView.SetMaxSidebarWidth(200)
	splitView.SetSidebarWidthFraction(0.1)
	w.toastOverlay.SetChild(splitView)

	breakpoint := adw.NewBreakpoint(adw.BreakpointConditionParse("max-width: 1500sp"))
	breakpoint.AddSetter(splitView, "collapsed", true)
	breakpointBin.AddBreakpoint(breakpoint)

	rpane := gtk.NewPaned(gtk.OrientationHorizontal)
	rpane.SetShrinkStartChild(false)
	rpane.SetShrinkEndChild(false)
	rpane.SetHExpand(true)
	splitView.SetContent(rpane)

	w.detailView = NewDetailView(ctx, w.ClusterState, editor)
	nav := adw.NewNavigationView()
	nav.Add(w.detailView.NavigationPage)
	nav.SetSizeRequest(350, 350)
	rpane.SetEndChild(nav)

	listHeader := NewListHeader(ctx, w.ClusterState, breakpoint, func() { splitView.SetShowSidebar(true) }, editor)
	w.listView = NewListView(ctx, w.ClusterState, listHeader)
	rpane.SetStartChild(w.listView)
	sw, _ := w.listView.SizeRequest()
	rpane.SetPosition(sw)

	w.navigation = NewNavigation(ctx, w.ClusterState)
	splitView.SetSidebar(w.navigation)

	w.createActions()

	return &w
}

func (w *ClusterWindow) createActions() {
	newWindow := gio.NewSimpleAction("newWindow", nil)
	newWindow.ConnectActivate(func(_ *glib.Variant) {
		prefs, err := api.LoadPreferences()
		if err != nil {
			widget.ShowErrorDialog(w.ctx, "Could not load preferences", err)
			return
		}
		prefs.Defaults()
		NewWelcomeWindow(context.WithoutCancel(w.ctx), w.Application(), w.State).Show()
	})
	w.AddAction(newWindow)
	w.Application().SetAccelsForAction("win.newWindow", []string{"<Ctrl>N"})

	disconnect := gio.NewSimpleAction("disconnect", nil)
	disconnect.ConnectActivate(func(_ *glib.Variant) {
		w.ActivateAction("newWindow", nil)
		w.cancel()
		w.Close()
	})
	w.AddAction(disconnect)
	w.Application().SetAccelsForAction("win.disconnect", []string{"<Ctrl>Q"})

	action := gio.NewSimpleAction("prefs", nil)
	action.ConnectActivate(func(_ *glib.Variant) {
		prefs := NewPreferencesWindow(w.ctx, w.State)
		prefs.SetTransientFor(&w.Window)
		prefs.Show()
	})
	w.AddAction(action)

	action = gio.NewSimpleAction("about", nil)
	action.ConnectActivate(func(_ *glib.Variant) {
		NewAboutWindow(&w.Window).Show()
	})
	w.AddAction(action)

	filterNamespace := gio.NewSimpleAction("filterNamespace", glib.NewVariantType("s"))
	filterNamespace.ConnectActivate(func(parameter *glib.Variant) {
		text := strings.Trim(fmt.Sprintf("%s ns:%s", w.SearchText.Value(), parameter.String()), " ")
		w.SearchText.Update(text)
	})
	actionGroup := gio.NewSimpleActionGroup()
	actionGroup.AddAction(filterNamespace)
	w.InsertActionGroup("list", actionGroup)
}
