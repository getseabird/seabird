package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/behavior"
	"github.com/getseabird/seabird/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const GtkInvalidListPosition uint = 4294967295

type ListView struct {
	*gtk.Box
	behavior   *behavior.ListBehavior
	parent     *gtk.Window
	selection  *gtk.SingleSelection
	columnView *gtk.ColumnView
	columns    []*gtk.ColumnViewColumn
	columnType *v1.APIResource
	objects    []client.Object
	selected   types.UID
}

func NewListView(parent *gtk.Window, behavior *behavior.ListBehavior) *ListView {
	l := ListView{
		Box:      gtk.NewBox(gtk.OrientationVertical, 0),
		parent:   parent,
		behavior: behavior,
	}
	l.AddCSSClass("view")

	header := adw.NewHeaderBar()
	header.AddCSSClass("flat")
	header.SetShowEndTitleButtons(false)
	header.SetShowStartTitleButtons(false)
	header.SetTitleWidget(NewListHeader(behavior))
	l.Append(header)

	l.selection = l.createModel()
	l.columnView = gtk.NewColumnView(l.selection)
	l.columnView.SetMarginStart(16)
	l.columnView.SetMarginEnd(16)
	l.columnView.SetMarginBottom(16)

	sw := gtk.NewScrolledWindow()
	sw.SetHExpand(true)
	sw.SetVExpand(true)
	sw.SetSizeRequest(500, 0)
	vp := gtk.NewViewport(nil, nil)
	vp.SetChild(l.columnView)
	sw.SetChild(vp)
	l.Append(sw)

	onChange(l.behavior.Objects, l.onObjectsChange)
	onChange(l.behavior.SearchFilter, l.onSearchFilterChange)

	return &l
}

func (l *ListView) onObjectsChange(objects []client.Object) {
	l.objects = objects

	if l.columnType == nil || !util.ResourceEquals(l.columnType, l.behavior.SelectedResource.Value()) {
		l.columnType = l.behavior.SelectedResource.Value()

		l.selection = l.createModel()
		l.columnView.SetModel(l.selection)

		for _, column := range l.columns {
			l.columnView.RemoveColumn(column)
		}
		l.columns = l.createColumns()
		for _, column := range l.columns {
			l.columnView.AppendColumn(column)
		}
	} else {
		list := l.selection.Model().Cast().(*gtk.StringList)
		list.Splice(0, list.NItems(), nil)
	}

	filter := l.behavior.SearchFilter.Value()
	for i, o := range objects {
		if !filter.Test(o) {
			continue
		}
		l.selection.Model().Cast().(*gtk.StringList).Append(strconv.Itoa(i))
		if o.GetUID() == l.selected {
			l.selection.SetSelected(uint(i))
		}
	}

	if len(l.objects) > 0 {
		if selected := l.selection.Selected(); selected == GtkInvalidListPosition {
			l.selection.SetSelected(0)
			l.behavior.RootDetailBehavior.SelectedObject.Update(l.objects[0])
		} else {
			l.behavior.RootDetailBehavior.SelectedObject.Update(l.objects[selected])
		}
	} else {
		l.behavior.RootDetailBehavior.SelectedObject.Update(nil)
	}
}

func (l *ListView) onSearchFilterChange(filter behavior.SearchFilter) {
	l.selection = l.createModel()
	l.columnView.SetModel(l.selection)
	for i, object := range l.behavior.Objects.Value() {
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

	if l.behavior.SelectedResource.Value().Namespaced {
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

	switch util.ResourceGVR(l.behavior.SelectedResource.Value()).String() {
	case corev1.SchemeGroupVersion.WithResource("pods").String():
		columns = append(columns,
			l.createColumn("Status", func(listitem *gtk.ListItem, object client.Object) {
				pod := object.(*corev1.Pod)
				for _, cond := range pod.Status.Conditions {
					if cond.Type == corev1.ContainersReady {
						listitem.SetChild(createStatusIcon(cond.Status == corev1.ConditionTrue || cond.Reason == "PodCompleted"))
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
	case corev1.SchemeGroupVersion.WithResource("persistentvolumeclaims").String():
		columns = append(columns,
			l.createColumn("Size", func(listitem *gtk.ListItem, object client.Object) {
				pvc := object.(*corev1.PersistentVolumeClaim)
				listitem.SetChild(gtk.NewLabel(pvc.Spec.Resources.Requests.Storage().String()))
			}),
		)
	case corev1.SchemeGroupVersion.WithResource("persistentvolumes").String():
		columns = append(columns,
			l.createColumn("Size", func(listitem *gtk.ListItem, object client.Object) {
				pvc := object.(*corev1.PersistentVolume)
				listitem.SetChild(gtk.NewLabel(pvc.Spec.Capacity.Storage().String()))
			}),
		)
	case appsv1.SchemeGroupVersion.WithResource("deployments").String():
		columns = append(columns,
			l.createColumn("Status", func(listitem *gtk.ListItem, object client.Object) {
				deployment := object.(*appsv1.Deployment)
				for _, cond := range deployment.Status.Conditions {
					if cond.Type == appsv1.DeploymentAvailable {
						listitem.SetChild(createStatusIcon(cond.Status == corev1.ConditionTrue))
					}
				}
			}),
			l.createColumn("Available", func(listitem *gtk.ListItem, object client.Object) {
				deployment := object.(*appsv1.Deployment)
				label := gtk.NewLabel(fmt.Sprintf("%d/%d", deployment.Status.AvailableReplicas, deployment.Status.Replicas))
				label.SetHAlign(gtk.AlignStart)
				listitem.SetChild(label)
			}),
		)
	case appsv1.SchemeGroupVersion.WithResource("statefulsets").String():
		columns = append(columns,
			l.createColumn("Status", func(listitem *gtk.ListItem, object client.Object) {
				statefulset := object.(*appsv1.StatefulSet)
				listitem.SetChild(createStatusIcon(statefulset.Status.ReadyReplicas == statefulset.Status.Replicas))
			}),
			l.createColumn("Available", func(listitem *gtk.ListItem, object client.Object) {
				statefulSet := object.(*appsv1.StatefulSet)
				label := gtk.NewLabel(fmt.Sprintf("%d/%d", statefulSet.Status.AvailableReplicas, statefulSet.Status.Replicas))
				label.SetHAlign(gtk.AlignStart)
				listitem.SetChild(label)
			}),
		)
	}

	return columns
}

func (l *ListView) createColumn(name string, bind func(listitem *gtk.ListItem, object client.Object)) *gtk.ColumnViewColumn {
	factory := gtk.NewSignalListItemFactory()
	factory.ConnectBind(func(listitem *gtk.ListItem) {
		idx, _ := strconv.Atoi(listitem.Item().Cast().(*gtk.StringObject).String())
		object := l.objects[idx]
		bind(listitem, object)
	})
	column := gtk.NewColumnViewColumn(name, &factory.ListItemFactory)
	column.SetResizable(true)
	column.SetExpand(true)
	return column
}

func (l *ListView) createModel() *gtk.SingleSelection {
	selection := gtk.NewSingleSelection(gtk.NewStringList([]string{}))
	selection.ConnectSelectionChanged(func(_, _ uint) {
		obj := l.objects[l.selection.Selected()]
		l.selected = obj.GetUID()
		l.behavior.RootDetailBehavior.SelectedObject.Update(obj)
	})
	return selection
}

func createStatusIcon(ok bool) *gtk.Image {
	if ok {
		icon := gtk.NewImageFromIconName("emblem-ok-symbolic")
		icon.AddCSSClass("success")
		icon.SetHAlign(gtk.AlignStart)
		return icon
	}
	icon := gtk.NewImageFromIconName("dialog-warning")
	icon.AddCSSClass("warning")
	icon.SetHAlign(gtk.AlignStart)
	return icon
}
