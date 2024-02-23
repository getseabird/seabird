package extension

import (
	"context"
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/widget"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
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

func (e *Apps) CreateObjectProperties(object client.Object, props []api.Property) []api.Property {
	switch object := object.(type) {
	case *appsv1.Deployment:
		prop := &api.GroupProperty{Name: "Pods"}
		var pods v1.PodList
		e.List(context.TODO(), &pods, client.InNamespace(object.Namespace), client.MatchingLabels(object.Spec.Selector.MatchLabels))
		// TODO should we also filter pods by owner? takes one more api call to fetch replicasets
		for i, pod := range pods.Items {
			prop.Children = append(prop.Children, &api.TextProperty{
				ID:     fmt.Sprintf("pods.%d", i),
				Source: &pod,
				Value:  pod.Name,
				Widget: func(w gtk.Widgetter, nv *adw.NavigationView) {
					podWidget(pod, w, nv)
				},
			})
		}
		props = append(props, prop)
	case *appsv1.StatefulSet:
		prop := &api.GroupProperty{Name: "Pods"}
		var pods v1.PodList
		e.List(context.TODO(), &pods, client.InNamespace(object.Namespace), client.MatchingLabels(object.Spec.Selector.MatchLabels))
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
				ID:     fmt.Sprintf("pods.%d", i),
				Source: &pod,
				Value:  pod.Name,
				Widget: func(w gtk.Widgetter, nv *adw.NavigationView) {
					podWidget(pod, w, nv)
				},
			})
		}
		props = append(props, prop)
	}

	return props
}

func podWidget(pod v1.Pod, w gtk.Widgetter, nv *adw.NavigationView) {
	switch row := w.(type) {
	case *adw.ActionRow:
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.ContainersReady {
				row.AddPrefix(widget.NewStatusIcon(cond.Status == corev1.ConditionTrue))
			}
		}
	}
}
