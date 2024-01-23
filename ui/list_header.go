package ui

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/jgillich/kubegio/internal"
	"github.com/jgillich/kubegio/util"
	"github.com/kelindar/event"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ListHeader struct {
	*gtk.Box
}

func NewListHeader(root *ClusterWindow) *ListHeader {
	box := gtk.NewBox(gtk.OrientationHorizontal, 0)
	box.AddCSSClass("linked")
	box.SetMarginStart(12)
	box.SetMarginEnd(12)

	kind := gtk.NewDropDown(gtk.NewStringList(nil), nil)
	// TODO need expression? https://docs.gtk.org/gtk4/property.DropDown.expression.html
	// dropdown.SetEnableSearch(true)
	for _, r := range root.cluster.Resources {
		kind.Model().Cast().(*gtk.StringList).Append(r.Kind)
	}
	kind.Connect("notify::selected-item", func() {
		res := root.cluster.Resources[kind.Selected()]
		root.listView.SetResource(util.ResourceGVR(&res))
	})
	box.Append(kind)

	entry := gtk.NewSearchEntry()
	entry.SetHExpand(true)
	box.Append(entry)

	button := gtk.NewMenuButton()
	button.SetIconName("view-more-symbolic")
	box.Append(button)

	filterNamespace := gio.NewSimpleAction("filterNamespace", glib.NewVariantType("s"))
	filterNamespace.ConnectActivate(func(parameter *glib.Variant) {
		entry.SetText(strings.Trim(fmt.Sprintf("%s ns:%s", entry.Text(), parameter.String()), " "))
	})
	actionGroup := gio.NewSimpleActionGroup()
	actionGroup.AddAction(filterNamespace)
	root.InsertActionGroup("list", actionGroup)

	namespace := gio.NewMenu()
	var list corev1.NamespaceList
	if err := root.cluster.List(context.TODO(), &list); err != nil {
		// TODO
	}
	for _, ns := range list.Items {
		namespace.Append(ns.GetName(), fmt.Sprintf("list.filterNamespace('%s')", ns.GetName()))
	}
	model := gio.NewMenu()
	model.AppendSection("Namespace", namespace)
	popover := gtk.NewPopoverMenuFromModel(model)
	button.SetPopover(popover)

	entry.ConnectSearchChanged(func() {
		root.listView.SetFilter(NewSearchFilter(entry.Text()))
	})

	event.On(func(ev internal.ResourceChanged) {
		var idx uint
		for i, r := range root.cluster.Resources {
			if r.String() == ev.APIResource.String() {
				idx = uint(i)
				break
			}
			glib.IdleAdd(func() {
				kind.SetSelected(idx)
			})
		}
	})

	return &ListHeader{box}
}

type SearchFilter struct {
	Name      []string
	Namespace []string
}

func NewSearchFilter(text string) SearchFilter {
	filter := SearchFilter{}

	for _, term := range strings.Split(text, " ") {
		if strings.HasPrefix(term, "ns:") {
			filter.Namespace = append(filter.Namespace, strings.TrimPrefix(term, "ns:"))
		} else {
			filter.Name = append(filter.Name, term)
		}
	}

	return filter
}

func (f *SearchFilter) Test(object client.Object) bool {
	{
		var ok bool
		for _, n := range f.Namespace {
			if object.GetNamespace() == n {
				ok = true
				break
			}
		}
		if !ok && len(f.Namespace) > 0 {
			return false
		}
	}

	for _, term := range f.Name {
		var ok bool
		trimmed := strings.Trim(term, "\"")
		if strings.Contains(object.GetName(), trimmed) {
			ok = true
			continue
		}
		if term != trimmed {
			continue
		}
		for _, term := range strings.Split(term, "-") {
			for _, name := range strings.Split(object.GetName(), "-") {
				if strutil.Similarity(name, term, metrics.NewHamming()) > 0.5 {
					ok = true
				}
			}
		}
		if !ok {
			return false
		}
	}

	return true
}
