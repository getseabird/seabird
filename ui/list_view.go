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
	root       *ClusterWindow
	resource   *metav1.APIResource
	items      []client.Object
	selection  *gtk.SingleSelection
	columnView *gtk.ColumnView
	columns    []*gtk.ColumnViewColumn
}

func NewListView(root *ClusterWindow) *ListView {
	l := ListView{Box: gtk.NewBox(gtk.OrientationVertical, 0), root: root}
	l.AddCSSClass("view")

	header := adw.NewHeaderBar()
	header.AddCSSClass("flat")
	header.SetShowEndTitleButtons(false)
	header.SetShowStartTitleButtons(false)
	header.SetTitleWidget(NewSearchBar(root))
	l.Append(header)

	l.selection = l.createModel()
	l.columnView = gtk.NewColumnView(l.selection)
	l.columnView.SetMarginStart(16)
	l.columnView.SetMarginEnd(16)
	l.columnView.SetMarginBottom(16)

	sw := gtk.NewScrolledWindow()
	sw.SetVExpand(true)
	sw.SetHExpand(true)
	sw.SetSizeRequest(500, 0)
	vp := gtk.NewViewport(nil, nil)
	vp.SetChild(l.columnView)
	sw.SetChild(vp)
	l.Append(sw)

	return &l
}

func (l *ListView) SetResource(gvr schema.GroupVersionResource) error {
	for _, res := range l.root.cluster.Resources {
		if res.Group == gvr.Group && res.Version == gvr.Version && res.Name == gvr.Resource {
			l.resource = &res
			break
		}
	}

	l.selection = l.createModel()
	l.columnView.SetModel(l.selection)

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
		l.selection.Model().Cast().(*gtk.StringList).Append(strconv.Itoa(i))
	}

	if len(l.items) > 0 {
		l.selection.SetSelected(0)
		l.root.detailView.SetObject(l.items[0])
	}

	return nil

}

func (l *ListView) SetFilter(filter SearchFilter) {
	l.selection = l.createModel()
	l.columnView.SetModel(l.selection)
	for i, object := range l.items {
		if filter.Test(object) {
			l.selection.Model().Cast().(*gtk.StringList).Append(strconv.Itoa(i))
		}
	}
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
func (l *ListView) listResource(ctx context.Context, gvr schema.GroupVersionResource) ([]client.Object, error) {
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
		if err := l.root.cluster.List(ctx, list); err != nil {
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
		list, err := l.root.cluster.Dynamic.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
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

func (l *ListView) createModel() *gtk.SingleSelection {
	selection := gtk.NewSingleSelection(gtk.NewStringList([]string{}))
	selection.ConnectSelectionChanged(func(_, _ uint) {
		l.root.detailView.SetObject(l.items[l.selection.Selected()])
	})
	return selection
}
