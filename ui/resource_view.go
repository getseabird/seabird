package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceView struct {
	*gtk.ColumnView
	list     *gtk.StringList
	resource *metav1.APIResource
}

func NewResourceView() *ResourceView {
	// store := gtk.NewListStore([]glib.Type{glib.TypeString, glib.TypeString})
	// listStore := gtk.NewSliceListModel([]glib.Type{glib.TypeString, glib.TypeString}, 0, 1)

	list := gtk.NewStringList([]string{})

	// sl := gtk.NewStringList

	cv := gtk.NewColumnView(gtk.NewSingleSelection(list))
	cv.SetHExpand(true)
	cv.SetVExpand(true)

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
		column.SetExpand(true)
		column.SetResizable(true)
		cv.AppendColumn(column)
	}

	// cv.SetModel(store)

	self := ResourceView{
		ColumnView: cv,
		list:       list,
		resource:   nil,
	}

	self.SetResource(metav1.APIResource{})

	return &self
}

func (r *ResourceView) SetResource(resource metav1.APIResource) {
	r.resource = &resource

	var pods corev1.PodList
	if err := application.cluster.List(context.TODO(), &pods); err != nil {
		panic(err)
	}

	for _, pod := range pods.Items {
		r.list.Append(fmt.Sprintf("%s|%s", pod.Name, pod.Namespace))
	}

}
