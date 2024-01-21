package ui

import (
	"context"
	"strconv"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/jgillich/kubegio/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ListView struct {
	*gtk.Box
	list       *gtk.StringList
	resource   *metav1.APIResource
	items      []client.Object
	selection  *gtk.SingleSelection
	columnView *gtk.ColumnView
	columns    []*gtk.ColumnViewColumn
}

func NewListView() *ListView {
	l := ListView{}

	l.Box = gtk.NewBox(gtk.OrientationVertical, 0)
	l.AddCSSClass("view")

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
	l.Append(header)

	l.list = gtk.NewStringList([]string{})
	l.selection = gtk.NewSingleSelection(l.list)
	l.columnView = gtk.NewColumnView(l.selection)
	l.columnView.SetMarginStart(16)
	l.columnView.SetMarginEnd(16)
	sw := gtk.NewScrolledWindow()
	sw.SetVExpand(true)
	sw.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	sw.SetChild(l.columnView)
	l.Append(sw)

	l.selection.ConnectSelectionChanged(func(_, _ uint) {
		application.detailView.SetObject(l.items[l.selection.Selected()])
	})

	return &l
}

func (l *ListView) SetResource(gvr schema.GroupVersionResource) error {
	for _, res := range application.cluster.Resources {
		if res.Group == gvr.Group && res.Version == gvr.Version && res.Name == gvr.Resource {
			l.resource = &res
			break
		}
	}

	for {
		length := uint(len(l.items))
		if length > 0 {
			l.list.Remove(length - 1)
			l.items = l.items[:length-1]
		} else {
			break
		}
	}

	for _, column := range l.columns {
		l.columnView.RemoveColumn(column)
	}
	l.columns = l.createColumns()
	for _, column := range l.columns {
		l.columnView.AppendColumn(column)
	}

	list, err := l.listResource(context.TODO(), gvr)
	if err != nil {
		return err
	}
	l.items = list
	for i, _ := range list {
		l.list.Append(strconv.Itoa(i))
	}

	if len(l.items) > 0 {
		l.selection.SetSelected(0)
		application.detailView.SetObject(l.items[0])
	}

	return nil

}

func (l *ListView) createColumns() []*gtk.ColumnViewColumn {
	var columns []*gtk.ColumnViewColumn

	columns = append(columns, l.createColumn("Name", func(listitem *gtk.ListItem, object client.Object) {
		label := gtk.NewLabel(object.GetName())
		label.SetHAlign(gtk.AlignStart)
		listitem.SetChild(label)
	}))

	if l.resource.Namespaced {
		columns = append(columns, l.createColumn("Namespace", func(listitem *gtk.ListItem, object client.Object) {
			label := gtk.NewLabel(object.GetNamespace())
			label.SetHAlign(gtk.AlignStart)
			listitem.SetChild(label)
		}))
	}

	columns = append(columns, l.createColumn("Age", func(listitem *gtk.ListItem, object client.Object) {
		duration := time.Since(object.GetCreationTimestamp().Time)
		label := gtk.NewLabel(util.HumanizeApproximateDuration(duration))
		label.SetHAlign(gtk.AlignStart)
		listitem.SetChild(label)
	}))

	switch util.ResourceGVR(l.resource).String() {
	case corev1.SchemeGroupVersion.WithResource("pods").String():
		columns = append(columns,
			l.createColumn("Status", func(listitem *gtk.ListItem, object client.Object) {
				pod := object.(*corev1.Pod)
				for _, cond := range pod.Status.Conditions {
					if cond.Type == corev1.ContainersReady {
						if cond.Status == corev1.ConditionTrue {
							icon := gtk.NewImageFromIconName("emblem-ok-symbolic")
							icon.AddCSSClass("success")
							listitem.SetChild(icon)
						} else {
							icon := gtk.NewImageFromIconName("dialog-warning")
							icon.AddCSSClass("warning")
							listitem.SetChild(icon)
						}
					}
				}
			}),
			l.createColumn("Restarts", func(listitem *gtk.ListItem, object client.Object) {
				pod := object.(*corev1.Pod)
				var restartCount int
				for _, container := range pod.Status.ContainerStatuses {
					restartCount += int(container.RestartCount)
				}
				label := gtk.NewLabel(strconv.Itoa(restartCount))
				label.SetHAlign(gtk.AlignStart)
				listitem.SetChild(label)
			}),
		)
	case appsv1.SchemeGroupVersion.WithResource("deployments").String():
		columns = append(columns,
			l.createColumn("Status", func(listitem *gtk.ListItem, object client.Object) {
				deployment := object.(*appsv1.Deployment)
				for _, cond := range deployment.Status.Conditions {
					if cond.Type == appsv1.DeploymentAvailable {
						if cond.Status == corev1.ConditionTrue {
							icon := gtk.NewImageFromIconName("emblem-ok-symbolic")
							icon.AddCSSClass("success")
							listitem.SetChild(icon)
						} else {
							icon := gtk.NewImageFromIconName("dialog-warning")
							icon.AddCSSClass("warning")
							listitem.SetChild(icon)
						}
					}
				}
			}),
		)
	case appsv1.SchemeGroupVersion.WithResource("statefulsets").String():
		columns = append(columns,
			l.createColumn("Status", func(listitem *gtk.ListItem, object client.Object) {
				statefulset := object.(*appsv1.StatefulSet)
				if statefulset.Status.ReadyReplicas == statefulset.Status.Replicas {
					icon := gtk.NewImageFromIconName("emblem-ok-symbolic")
					icon.AddCSSClass("success")
					listitem.SetChild(icon)
				} else {
					icon := gtk.NewImageFromIconName("dialog-warning")
					icon.AddCSSClass("warning")
					listitem.SetChild(icon)
				}
			}),
		)
	}

	return columns
}

func (l *ListView) createColumn(name string, bind func(listitem *gtk.ListItem, object client.Object)) *gtk.ColumnViewColumn {
	factory := gtk.NewSignalListItemFactory()
	factory.ConnectBind(func(listitem *gtk.ListItem) {
		idx, _ := strconv.Atoi(listitem.Item().Cast().(*gtk.StringObject).String())
		object := l.items[idx]
		bind(listitem, object)
	})
	column := gtk.NewColumnViewColumn(name, &factory.ListItemFactory)
	column.SetResizable(true)
	return column
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
