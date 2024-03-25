package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strconv"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/style"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/util"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Navigation struct {
	*adw.ToolbarView
	*common.ClusterState
	list *gtk.ListBox
	rows []*gtk.ListBoxRow
}

func NewNavigation(ctx context.Context, state *common.ClusterState) *Navigation {
	n := &Navigation{ToolbarView: adw.NewToolbarView(), ClusterState: state}
	n.SetVExpand(true)
	n.AddCSSClass("sidebar-pane")

	header := adw.NewHeaderBar()
	title := gtk.NewLabel(n.ClusterPreferences.Value().Name)
	title.SetEllipsize(pango.EllipsizeEnd)
	title.AddCSSClass("heading")
	header.SetTitleWidget(title)
	header.SetShowEndTitleButtons(false)
	switch style.Get() {
	case style.Darwin:
		header.SetShowTitle(false)
	default:
		header.SetShowStartTitleButtons(false)
	}

	button := gtk.NewMenuButton()
	button.SetIconName("open-menu-symbolic")

	windowSection := gio.NewMenu()
	windowSection.Append("New Window", "win.newWindow")
	windowSection.Append("Disconnect", "win.disconnect")

	prefSection := gio.NewMenu()
	prefSection.Append("Preferences", "win.prefs")
	// prefSection.Append("Keyboard Shortcuts", "win.shortcuts")
	prefSection.Append("About", "win.about")

	m := gio.NewMenu()
	m.AppendSection("", windowSection)
	m.AppendSection("", prefSection)

	popover := gtk.NewPopoverMenuFromModel(m)
	button.SetPopover(popover)

	header.PackEnd(button)
	n.AddTopBar(header)

	content := gtk.NewBox(gtk.OrientationVertical, 0)
	sw := gtk.NewScrolledWindow()
	sw.SetChild(content)
	n.SetContent(sw)

	listBin := adw.NewBin()
	listBin.SetVExpand(false)
	content.Append(listBin)

	common.OnChange(ctx, n.ClusterPreferences, func(prefs api.ClusterPreferences) {
		listBin.SetChild(n.createList(prefs))
	})

	listBin.SetChild(n.createList(n.ClusterPreferences.Value()))
	if len(n.rows) > 0 {
		n.list.SelectRow(n.rows[0])
	}

	// TODO actions should be able to use "u" for uint but I can't get it to work
	actionGroup := gio.NewSimpleActionGroup()
	pin := gio.NewSimpleAction("pin", glib.NewVariantType("s"))
	pin.ConnectActivate(func(idx *glib.Variant) {
		id, _ := strconv.Atoi(idx.String())
		prefs := n.ClusterPreferences.Value()
		prefs.Navigation.Favourites = append(prefs.Navigation.Favourites, util.ResourceGVR(&n.Resources[id]))
		n.ClusterPreferences.Update(prefs)
	})
	actionGroup.AddAction(pin)
	unpin := gio.NewSimpleAction("unpin", glib.NewVariantType("s"))
	unpin.ConnectActivate(func(idx *glib.Variant) {
		id, _ := strconv.Atoi(idx.String())
		prefs := n.ClusterPreferences.Value()
		for i, f := range prefs.Navigation.Favourites {
			if util.GVREquals(f, util.ResourceGVR(&n.Resources[id])) {
				prefs.Navigation.Favourites = slices.Delete(prefs.Navigation.Favourites, i, i+1)
				n.ClusterPreferences.Update(prefs)
				break
			}
		}
	})
	actionGroup.AddAction(unpin)
	n.InsertActionGroup("navigation", actionGroup)

	return n
}

func (n *Navigation) createList(prefs api.ClusterPreferences) *gtk.ListBox {
	n.list = gtk.NewListBox()
	n.list.SetHeaderFunc(func(row, before *gtk.ListBoxRow) {
		switch {
		case row.Index() == 0:
			label := gtk.NewLabel("Favourites")
			label.AddCSSClass("caption-heading")
			label.SetHAlign(gtk.AlignStart)
			r := gtk.NewListBoxRow()
			r.SetChild(label)
			r.SetSensitive(false)
			row.SetHeader(r)
		case row.Index() == len(prefs.Navigation.Favourites):
			label := gtk.NewLabel("Resources")
			label.AddCSSClass("caption-heading")
			label.SetHAlign(gtk.AlignStart)
			r := gtk.NewListBoxRow()
			r.SetChild(label)
			r.SetSensitive(false)
			row.SetHeader(r)
		}
	})
	n.list.AddCSSClass("dim-label")
	n.list.AddCSSClass("navigation-sidebar")
	n.list.ConnectRowSelected(func(row *gtk.ListBoxRow) {
		if row == nil {
			return
		}
		var gvr schema.GroupVersionResource
		if err := json.Unmarshal([]byte(row.Name()), &gvr); err != nil {
			log.Printf("failed to unmarshal gvr: %v", err)
			return
		}
		for _, res := range n.Resources {
			if util.GVREquals(util.ResourceGVR(&res), gvr) && !util.ResourceEquals(n.SelectedResource.Value(), &res) {
				n.SelectedResource.Update(&res)
				break
			}
		}
	})

	n.rows = nil

	var favs int
	for i, resource := range n.Resources {
		var fav bool
		for _, f := range prefs.Navigation.Favourites {
			if util.GVREquals(f, util.ResourceGVR(&resource)) {
				favs++
				fav = true
			}
		}
		row := createResourceRow(&resource, i, fav)
		if fav && len(n.rows) > 0 {
			n.rows = slices.Insert(n.rows, favs-1, row)
		} else {
			n.rows = append(n.rows, row)
		}
	}

	for _, row := range n.rows {
		n.list.Append(row)
	}

	return n.list
}

func createResourceRow(resource *metav1.APIResource, idx int, fav bool) *gtk.ListBoxRow {
	gvr := util.ResourceGVR(resource)

	row := gtk.NewListBoxRow()
	json, err := json.Marshal(gvr)
	if err != nil {
		panic(err)
	}
	row.SetName(string(json))
	box := gtk.NewBox(gtk.OrientationHorizontal, 8)
	box.SetMarginTop(4)
	box.SetMarginBottom(4)
	box.Append(resourceImage(gvr))
	vbox := gtk.NewBox(gtk.OrientationVertical, 2)
	vbox.SetVAlign(gtk.AlignCenter)
	box.Append(vbox)
	label := gtk.NewLabel(resource.Kind)
	label.SetHAlign(gtk.AlignStart)
	label.SetEllipsize(pango.EllipsizeEnd)
	vbox.Append(label)
	label = gtk.NewLabel(resource.Group)
	if resource.Group == "" {
		label.SetText("k8s.io")
	}
	label.SetHAlign(gtk.AlignStart)
	label.AddCSSClass("caption")
	label.AddCSSClass("dim-label")
	label.SetEllipsize(pango.EllipsizeEnd)
	vbox.Append(label)
	row.SetChild(box)

	gesture := gtk.NewGestureClick()
	gesture.SetButton(gdk.BUTTON_SECONDARY)
	gesture.ConnectPressed(func(nPress int, x, y float64) {
		menu := gio.NewMenu()
		if fav {
			menu.Append("Move to Resources", fmt.Sprintf("navigation.unpin('%d')", idx))
		} else {
			menu.Append("Move to Favourites", fmt.Sprintf("navigation.pin('%d')", idx))
		}
		popover := gtk.NewPopoverMenuFromModel(menu)
		popover.SetHasArrow(false)
		row.FirstChild().(*gtk.Box).Append(popover)
		popover.Show()
	})
	row.AddController(gesture)

	return row
}

func resourceImage(gvk schema.GroupVersionResource) *gtk.Image {
	switch gvk.Group {
	case corev1.GroupName:
		{
			switch gvk.Resource {
			case "pods":
				return gtk.NewImageFromIconName("box-symbolic")
			case "configmaps":
				return gtk.NewImageFromIconName("file-sliders-symbolic")
			case "secrets":
				return gtk.NewImageFromIconName("file-key-2-symbolic")
			case "namespaces":
				return gtk.NewImageFromIconName("orbit-symbolic")
			case "services":
				return gtk.NewImageFromIconName("waypoints-symbolic")
			case "nodes":
				return gtk.NewImageFromIconName("server-symbolic")
			case "persistentvolumes":
				return gtk.NewImageFromIconName("hard-drive-download-symbolic")
			case "persistentvolumeclaims":
				return gtk.NewImageFromIconName("hard-drive-upload-symbolic")
			}
		}
	case appsv1.GroupName:
		switch gvk.Resource {
		case "replicasets":
			return gtk.NewImageFromIconName("layers-2-symbolic")
		case "deployments":
			return gtk.NewImageFromIconName("layers-3-symbolic")
		case "statefulsets":
			return gtk.NewImageFromIconName("database-symbolic")
		}
	case batchv1.GroupName:
		switch gvk.Resource {
		case "jobs":
			return gtk.NewImageFromIconName("briefcase-symbolic")
		case "cronjobs":
			return gtk.NewImageFromIconName("timer-reset-symbolic")
		}
	case networkingv1.GroupName:
		switch gvk.Resource {
		case "ingresses":
			return gtk.NewImageFromIconName("radio-tower-symbolic")
		}
	}

	return gtk.NewImageFromIconName("blocks")
}
