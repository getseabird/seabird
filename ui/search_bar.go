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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SearchBar struct {
	*gtk.Box
}

func NewSearchBar(root *ClusterWindow) *SearchBar {
	box := gtk.NewBox(gtk.OrientationHorizontal, 0)
	box.SetMarginStart(32)
	box.SetMarginEnd(32)

	entry := gtk.NewSearchEntry()
	entry.SetHExpand(true)
	box.Append(entry)

	button := gtk.NewMenuButton()
	button.AddCSSClass("flat")
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

	return &SearchBar{box}
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

	{
		var ok bool
		for _, term := range f.Name {
			trimmed := strings.Trim(term, "\"")
			if strings.Contains(object.GetName(), trimmed) {
				ok = true
				break
			}
			if term != trimmed {
				continue
			}

			for _, name := range strings.Split(object.GetName(), "-") {
				for _, term := range strings.Split(term, "-") {
					if strutil.Similarity(name, term, metrics.NewHamming()) > 0.5 {
						ok = true
					}
				}
			}
		}
		if !ok && len(f.Name) > 0 {
			return false
		}
	}

	return true
}
