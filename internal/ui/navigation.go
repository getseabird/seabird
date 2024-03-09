package ui

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
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
	"k8s.io/utils/ptr"
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
	n.AddCSSClass("background")

	header := adw.NewHeaderBar()
	title := gtk.NewLabel(n.ClusterPreferences.Value().Name)
	title.SetEllipsize(pango.EllipsizeEnd)
	title.AddCSSClass("heading")
	header.SetTitleWidget(title)
	header.SetShowEndTitleButtons(false)
	switch style.Get() {
	case style.Darwin:
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

	favouritesBin := adw.NewBin()
	favouritesBin.SetVExpand(false)
	content.Append(favouritesBin)
	addFavouriteBin := adw.NewBin()
	content.Append(addFavouriteBin)

	common.OnChange(ctx, n.ClusterPreferences, func(prefs api.ClusterPreferences) {
		favouritesBin.SetChild(n.createFavourites(prefs))
	})

	common.OnChange(ctx, n.SelectedResource, func(res *metav1.APIResource) {
		if res == nil {
			return
		}
		var idx *int
		for i, r := range n.ClusterPreferences.Value().Navigation.Favourites {
			if util.ResourceGVR(res).String() == r.String() {
				idx = ptr.To(i)
				break
			}
		}
		if idx != nil {
			n.list.SelectRow(n.rows[*idx])
			addFavouriteBin.SetChild(nil)
		} else {
			n.list.SelectRow(nil)
			addFavouriteBin.SetChild(n.createAddFavourite(res))
		}
	})

	favouritesBin.SetChild(n.createFavourites(n.ClusterPreferences.Value()))
	if len(n.rows) > 0 {
		n.list.SelectRow(n.rows[0])
	}

	return n
}

func (n *Navigation) createFavourites(prefs api.ClusterPreferences) *gtk.ListBox {
	n.list = gtk.NewListBox()
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

	for _, gvr := range prefs.Navigation.Favourites {
		var resource *metav1.APIResource
		for _, r := range n.Resources {
			if r.Group == gvr.Group && r.Version == gvr.Version && r.Name == gvr.Resource {
				resource = &r
				break
			}
		}
		if resource == nil {
			log.Printf("ignoring unknown resource %s", gvr.String())
			n.rows = append(n.rows, nil)
			continue
		}

		row := gtk.NewListBoxRow()
		json, err := json.Marshal(gvr)
		if err != nil {
			panic(err)
		}
		row.SetName(string(json))
		box := gtk.NewBox(gtk.OrientationHorizontal, 8)
		box.SetMarginTop(4)
		box.SetMarginBottom(4)
		box.Append(n.resIcon(gvr))
		vbox := gtk.NewBox(gtk.OrientationVertical, 2)
		vbox.SetVAlign(gtk.AlignCenter)
		box.Append(vbox)
		label := gtk.NewLabel(resource.Kind)
		if len(resource.Kind) > 15 && len(resource.ShortNames) > 0 {
			label.SetText(strings.ToUpper(resource.ShortNames[0]))
		}
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

		// TODO add right click menu with an option to remove favourite
		// gesture := gtk.NewGestureClick()
		// gesture.SetButton(gdk.BUTTON_SECONDARY)
		// gesture.ConnectPressed(func(nPress int, x, y float64) {
		// 	log.Printf("pressed")
		// 	// model := gtk.NewStringList([]string{})
		// 	popover := gtk.NewPopoverMenuFromModel(nil)
		// 	popover.SetChild(gtk.NewLabel("popover"))
		// 	row.FirstChild().(*gtk.Box).Append(popover)
		// 	popover.Show()
		// })
		// row.AddController(gesture)

		n.list.Append(row)
		n.rows = append(n.rows, row)

		if res := n.SelectedResource.Value(); res != nil && util.GVREquals(util.ResourceGVR(res), gvr) {
			n.list.SelectRow(row)
		}
	}

	return n.list
}

func (n *Navigation) createAddFavourite(res *metav1.APIResource) *gtk.Box {
	content := gtk.NewBox(gtk.OrientationVertical, 0)
	content.Append(gtk.NewSeparator(gtk.OrientationHorizontal))
	list := gtk.NewListBox()
	list.AddCSSClass("dim-label")
	list.AddCSSClass("navigation-sidebar")
	row := gtk.NewListBoxRow()
	row.AddCSSClass("accent")
	box := gtk.NewBox(gtk.OrientationHorizontal, 8)
	box.Append(gtk.NewImageFromIconName("list-add"))
	box.Append(gtk.NewLabel(res.Kind))
	row.SetChild(box)
	list.Append(row)
	list.ConnectRowSelected(func(row *gtk.ListBoxRow) {
		v := n.ClusterPreferences.Value()
		v.Navigation.Favourites = append(v.Navigation.Favourites, util.ResourceGVR(res))
		n.ClusterPreferences.Update(v)
		content.Hide()
	})
	content.Append(list)
	return content
}

func (n *Navigation) resIcon(gvk schema.GroupVersionResource) *gtk.Image {
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
