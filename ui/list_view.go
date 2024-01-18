package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ListView struct {
	*gtk.ColumnView
	list     *gtk.StringList
	resource *metav1.APIResource
	items    []client.Object
}

func NewListView() *ListView {
	list := gtk.NewStringList([]string{})
	selection := gtk.NewSingleSelection(list)
	columnView := gtk.NewColumnView(selection)
	columnView.SetHExpand(true)
	columnView.SetVExpand(true)

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
		columnView.AppendColumn(column)
	}

	// cv.SetModel(store)

	self := ListView{
		ColumnView: columnView,
		list:       list,
		resource:   nil,
	}

	selection.ConnectSelectionChanged(func(position, _ uint) {
		application.DetailView(self.items[position])
	})

	self.SetResource(metav1.APIResource{})

	return &self
}

func (r *ListView) SetResource(resource metav1.APIResource) error {
	r.resource = &resource

	var pods corev1.PodList
	if err := application.cluster.List(context.TODO(), &pods); err != nil {
		return err
	}

	r.items = []client.Object{}
	for _, p := range pods.Items {
		pod := p
		r.items = append(r.items, &pod)
		r.list.Append(fmt.Sprintf("%s|%s", pod.Name, pod.Namespace))
	}

	return nil

}
