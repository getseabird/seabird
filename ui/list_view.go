package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

	list, err := r.listResource(context.TODO(), gvr)
	if err != nil {
		return err
	}
	r.items = list
	for _, item := range list {
		r.list.Append(fmt.Sprintf("%s|%s", item.GetName(), item.GetNamespace()))
	}

	if len(r.items) > 0 {
		r.selection.SetSelected(0)
		application.detailView.SetObject(r.items[0])
	}

	return nil

}

// We want typed objects for known resources so we can type switch them
func (r *ListView) listResource(ctx context.Context, gvr schema.GroupVersionResource) ([]client.Object, error) {
	var res []client.Object
	var list client.ObjectList
	switch gvr.String() {
	case corev1.SchemeGroupVersion.WithResource("pods").String():
		list = &corev1.PodList{}
	case corev1.SchemeGroupVersion.WithResource("configmaps").String():
		list = &corev1.ConfigMapList{}
	case corev1.SchemeGroupVersion.WithResource("secrets").String():
		list = &corev1.SecretList{}
	case appsv1.SchemeGroupVersion.WithResource("deployments").String():
		list = &appsv1.DeploymentList{}
	case appsv1.SchemeGroupVersion.WithResource("statefulsets").String():
		list = &appsv1.StatefulSetList{}
	}
	if list != nil {
		if err := application.cluster.List(ctx, list); err != nil {
			return nil, err
		}
		switch list := list.(type) {
		case *corev1.PodList:
			for _, i := range list.Items {
				ii := i
				res = append(res, &ii)
			}
		case *corev1.ConfigMapList:
			for _, i := range list.Items {
				ii := i
				res = append(res, &ii)
			}
		case *corev1.SecretList:
			for _, i := range list.Items {
				ii := i
				res = append(res, &ii)
			}
		case *appsv1.DeploymentList:
			for _, i := range list.Items {
				ii := i
				res = append(res, &ii)
			}
		case *appsv1.StatefulSetList:
			for _, i := range list.Items {
				ii := i
				res = append(res, &ii)
			}
		}

		return res, nil
	} else {
		list, err := application.cluster.Dynamic.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, i := range list.Items {
			ii := i
			res = append(res, &ii)
		}
		return res, nil
	}
}
