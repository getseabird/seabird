package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ListView struct {
	*gtk.Box
	list      *gtk.StringList
	resource  *metav1.APIResource
	items     []client.Object
	selection *gtk.SingleSelection
}

func NewListView() *ListView {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.AddCSSClass("view")

	header := adw.NewHeaderBar()
	header.AddCSSClass("flat")
	header.SetShowEndTitleButtons(false)
	header.SetShowStartTitleButtons(false)
	search := gtk.NewSearchBar()
	entry := gtk.NewSearchEntry()
	search.ConnectEntry(entry)
	entry.Show()
	search.Show()
	b := gtk.NewBox(gtk.OrientationVertical, 0)
	b.Append(search)
	b.Append(entry)
	header.SetTitleWidget(b)
	box.Append(header)

	list := gtk.NewStringList([]string{})
	selection := gtk.NewSingleSelection(list)
	columnView := gtk.NewColumnView(selection)
	columnView.SetHExpand(true)
	columnView.SetVExpand(true)
	columnView.SetMarginStart(16)
	columnView.SetMarginEnd(16)
	box.Append(columnView)

	columns := []string{"Name", "Namespace"}
	for i, name := range columns {
		ii := i
		factory := gtk.NewSignalListItemFactory()
		factory.ConnectBind(func(listitem *gtk.ListItem) {
			s := listitem.Item().Cast().(*gtk.StringObject).String()
			label := gtk.NewLabel(strings.Split(s, "|")[ii])
			label.SetHAlign(gtk.AlignStart)
			listitem.SetChild(label)
		})
		column := gtk.NewColumnViewColumn(name, &factory.ListItemFactory)
		column.SetResizable(true)
		columnView.AppendColumn(column)
	}

	self := ListView{
		Box:       box,
		list:      list,
		selection: selection,
		resource:  nil,
	}

	selection.ConnectSelectionChanged(func(_, _ uint) {
		application.detailView.SetObject(self.items[selection.Selected()])
	})

	// self.SetResource(schema.GroupVersionResource{Version: "v1", Resource: "pods"})

	return &self
}

func (r *ListView) SetResource(gvr schema.GroupVersionResource) error {
	for {
		length := uint(len(r.items))
		if length > 0 {
			r.list.Remove(length - 1)
			r.items = r.items[:length-1]
		} else {
			break
		}
	}

	for _, res := range application.cluster.Resources {
		if res.Group == gvr.Group && res.Version == gvr.Version && res.Name == gvr.Resource {
			r.resource = &res
			break
		}
	}

	list, err := application.cluster.Dynamic.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	r.items = []client.Object{}
	for _, item := range list.Items {
		r.items = append(r.items, &item)
		r.list.Append(fmt.Sprintf("%s|%s", item.GetName(), item.GetNamespace()))
	}

	if len(r.items) > 0 {
		r.selection.SetSelected(0)
		application.detailView.SetObject(r.items[0])
	}

	return nil

}
