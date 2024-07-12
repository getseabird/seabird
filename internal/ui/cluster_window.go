package ui

import (
	"context"
	"fmt"
	"math"
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
	dialog       *adw.Dialog
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

	w.dialog = adw.NewDialog()
	w.dialog.SetPresentationMode(adw.DialogBottomSheet)
	w.dialog.SetFollowsContentSize(true)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				glib.IdleAdd(func() {
					w.dialog.SetSizeRequest(int(math.Min(float64(w.Width())*0.6, 1000)), -1)
				})
			}
			time.Sleep(time.Second)
		}
	}()

	w.navigation = NewNavigation(ctx, w.ClusterState, viewStack, editor)
	w.navigation.SetSizeRequest(225, -1)
	paned.SetStartChild(w.navigation)

	navView := adw.NewNavigationView()
	w.objectView = NewObjectView(ctx, w.ClusterState, editor, navView, w.navigation)
	navView.Add(w.objectView.NavigationPage)
	navView.SetHExpand(true)
	w.dialog.SetChild(navView)

	w.listView = list.NewList(ctx, w.ClusterState, w.dialog, editor)
	viewStack.AddChild(w.listView).SetName("list")
	viewStack.SetVisibleChild(w.listView)
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
		prefs.Present()
	})
	w.AddAction(action)

	action = gio.NewSimpleAction("about", nil)
	action.ConnectActivate(func(_ *glib.Variant) {
		NewAboutWindow(&w.Window).Present()
	})
	w.AddAction(action)

}
