package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/ui/editor"
	"github.com/getseabird/seabird/internal/ui/list"
	"github.com/getseabird/seabird/widget"
)

type ClusterWindow struct {
	*widget.UniversalApplicationWindow
	*common.ClusterState
	ctx          context.Context
	cancel       context.CancelFunc
	navigation   *Navigation
	listView     *list.List
	objectView   *ObjectView
	toastOverlay *adw.ToastOverlay
	overlay      *adw.OverlaySplitView
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
	w.SetDefaultSize(1000, 600)

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

	viewStack := gtk.NewStack()
	viewStack.SetTransitionType(gtk.StackTransitionTypeCrossfade)

	w.toastOverlay = adw.NewToastOverlay()
	w.SetContent(w.toastOverlay)

	paned := gtk.NewPaned(gtk.OrientationHorizontal)
	paned.SetPosition(225)
	paned.SetShrinkStartChild(false)
	paned.SetShrinkEndChild(false)
	w.toastOverlay.SetChild(paned)

	// replace split view with sheet dialog? in adw 1.5
	// https://gnome.pages.gitlab.gnome.org/libadwaita/doc/1-latest/class.Dialog.html
	w.overlay = adw.NewOverlaySplitView()
	w.overlay.SetEnableHideGesture(true)
	w.overlay.SetEnableShowGesture(true)
	w.overlay.SetCollapsed(true)
	w.overlay.SetSidebarPosition(gtk.PackEnd)
	w.overlay.NotifyProperty("show-sidebar", w.resizeOverlay)

	w.listView = list.NewList(ctx, w.ClusterState, w.overlay, editor)
	w.overlay.SetContent(w.listView)

	w.navigation = NewNavigation(ctx, w.ClusterState, viewStack, editor)
	w.navigation.SetSizeRequest(225, -1)
	paned.SetStartChild(w.navigation)

	navView := adw.NewNavigationView()
	w.objectView = NewObjectView(ctx, w.ClusterState, editor, navView, w.navigation)
	navView.Add(w.objectView.NavigationPage)
	navView.SetHExpand(true)
	navView.SetSizeRequest(400, -1)
	w.overlay.SetSidebar(navView)

	viewStack.AddChild(w.overlay).SetName("list")
	viewStack.SetVisibleChild(w.overlay)
	paned.SetEndChild(viewStack)

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

}

func (w *ClusterWindow) resizeOverlay() {
	w.overlay.SetMaxSidebarWidth(float64(w.Width()-w.navigation.Width()) - 150)
	// Workaround for lack of window resize signals
	// could probably use size_allocate when subclassing is available
	go func() {
		time.Sleep(250 * time.Millisecond)
		glib.IdleAdd(func() {
			if w.overlay.ShowSidebar() {
				w.resizeOverlay()
			}
		})
	}()
}
