package extension

import (
	"context"
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/util"
	"github.com/getseabird/seabird/widget"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	Extensions = append(Extensions, func(cluster *api.Cluster) Extension {
		return &Apps{Cluster: cluster}
	})
}

type Apps struct {
	*api.Cluster
}

func (e *Apps) CreateColumns(ctx context.Context, resource *metav1.APIResource, columns []api.Column) []api.Column {
	switch util.ResourceGVR(resource).String() {
	case appsv1.SchemeGroupVersion.WithResource("deployments").String():
		columns = append(columns,
			api.Column{
				Name:     "Status",
				Priority: 70,
				Bind: func(listitem *gtk.ListItem, object client.Object) {
					listitem.SetChild(widget.NewStatusIcon(isReady(object)))
				},
				Compare: func(a, b client.Object) int {
					if isReady(a) == isReady(b) {
						return 0
					}
					if isReady(a) {
						return 1
					}
					return -1
				},
			},
			api.Column{
				Name:     "Available",
				Priority: 60,
				Bind: func(listitem *gtk.ListItem, object client.Object) {
					deployment := object.(*appsv1.Deployment)
					label := gtk.NewLabel(fmt.Sprintf("%d/%d", deployment.Status.AvailableReplicas, deployment.Status.Replicas))
					label.SetHAlign(gtk.AlignStart)
					listitem.SetChild(label)
				},
			},
		)
	case appsv1.SchemeGroupVersion.WithResource("statefulsets").String():
		columns = append(columns,
			api.Column{
				Name:     "Status",
				Priority: 70,
				Bind: func(listitem *gtk.ListItem, object client.Object) {
					listitem.SetChild(widget.NewStatusIcon(isReady(object)))
				},
				Compare: func(a, b client.Object) int {
					if isReady(a) == isReady(b) {
						return 0
					}
					if isReady(a) {
						return 1
					}
					return -1
				},
			},
			api.Column{
				Name:     "Available",
				Priority: 60,
				Bind: func(listitem *gtk.ListItem, object client.Object) {
					statefulSet := object.(*appsv1.StatefulSet)
					label := gtk.NewLabel(fmt.Sprintf("%d/%d", statefulSet.Status.AvailableReplicas, statefulSet.Status.Replicas))
					label.SetHAlign(gtk.AlignStart)
					listitem.SetChild(label)
				},
			},
		)
	}

	return columns
}

func (e *Apps) CreateObjectProperties(ctx context.Context, object client.Object, props []api.Property) []api.Property {
	switch object := object.(type) {
	case *appsv1.Deployment:
		prop := &api.GroupProperty{Name: "Pods"}
		var pods corev1.PodList
		e.List(ctx, &pods, client.InNamespace(object.Namespace), client.MatchingLabels(object.Spec.Selector.MatchLabels))
		// TODO should we also filter pods by owner? takes one more api call to fetch replicasets
		for i, pod := range pods.Items {
			prop.Children = append(prop.Children, &api.TextProperty{
				ID:        fmt.Sprintf("pods.%d", i),
				Reference: api.NewObjectReference(&pod),
				Value:     pod.Name,
				Widget: func(w gtk.Widgetter, nv *adw.NavigationView) {
					podWidget(ctx, pod, w, nv)
				},
			})
		}
		props = append(props, prop)
	case *appsv1.StatefulSet:
		prop := &api.GroupProperty{Name: "Pods"}
		var pods corev1.PodList
		e.List(ctx, &pods, client.InNamespace(object.Namespace), client.MatchingLabels(object.Spec.Selector.MatchLabels))
		for i, pod := range pods.Items {
			var ok bool
			for _, owner := range pod.OwnerReferences {
				if owner.UID == object.UID {
					ok = true
				}
			}
			if !ok {
				continue
			}
			prop.Children = append(prop.Children, &api.TextProperty{
				ID:        fmt.Sprintf("pods.%d", i),
				Reference: api.NewObjectReference(&pod),
				Value:     pod.Name,
				Widget: func(w gtk.Widgetter, nv *adw.NavigationView) {
					podWidget(ctx, pod, w, nv)
				},
			})
		}
		props = append(props, prop)
	}

	return props
}

func podWidget(ctx context.Context, pod corev1.Pod, w gtk.Widgetter, nv *adw.NavigationView) {
	switch row := w.(type) {
	case *adw.ActionRow:
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.ContainersReady {
				row.AddPrefix(widget.NewStatusIcon(cond.Status == corev1.ConditionTrue))
			}
		}
	}
}
