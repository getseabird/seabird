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
	"k8s.io/client-go/tools/reference"
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
	switch util.GVRForResource(resource).String() {
	case appsv1.SchemeGroupVersion.WithResource("deployments").String():
		columns = append(columns,
			api.Column{
				Name:     "Status",
				Priority: 70,
				Bind: func(listitem *gtk.ColumnViewCell, object client.Object) {
					listitem.SetChild(widget.ObjectStatus(object).Icon())
				},
				Compare: widget.CompareObjectStatus,
			},
			api.Column{
				Name:     "Available",
				Priority: 60,
				Bind: func(listitem *gtk.ColumnViewCell, object client.Object) {
					deployment := object.(*appsv1.Deployment)
					label := gtk.NewLabel(fmt.Sprintf("%d/%d", deployment.Status.AvailableReplicas, deployment.Status.Replicas))
					label.SetHAlign(gtk.AlignStart)
					listitem.SetChild(label)
				},
			},
		)
	case appsv1.SchemeGroupVersion.WithResource("replicasets").String():
		columns = append(columns,
			api.Column{
				Name:     "Status",
				Priority: 70,
				Bind: func(listitem *gtk.ColumnViewCell, object client.Object) {
					listitem.SetChild(widget.ObjectStatus(object).Icon())
				},
				Compare: widget.CompareObjectStatus,
			},
			api.Column{
				Name:     "Available",
				Priority: 60,
				Bind: func(listitem *gtk.ColumnViewCell, object client.Object) {
					replicaSet := object.(*appsv1.ReplicaSet)
					label := gtk.NewLabel(fmt.Sprintf("%d/%d", replicaSet.Status.AvailableReplicas, replicaSet.Status.Replicas))
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
				Bind: func(listitem *gtk.ColumnViewCell, object client.Object) {
					listitem.SetChild(widget.ObjectStatus(object).Icon())
				},
				Compare: widget.CompareObjectStatus,
			},
			api.Column{
				Name:     "Available",
				Priority: 60,
				Bind: func(listitem *gtk.ColumnViewCell, object client.Object) {
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

func (e *Apps) CreateObjectProperties(ctx context.Context, _ *metav1.APIResource, object client.Object, props []api.Property) []api.Property {
	switch object := object.(type) {
	case *appsv1.Deployment:
		prop := &api.GroupProperty{Name: "Pods"}
		var pods corev1.PodList
		e.List(ctx, &pods, client.InNamespace(object.Namespace), client.MatchingLabels(object.Spec.Selector.MatchLabels))
		for i, pod := range pods.Items {
			ref, _ := reference.GetReference(e.Scheme, &pod)
			prop.Children = append(prop.Children, &api.TextProperty{
				ID:        fmt.Sprintf("pods.%d", i),
				Reference: ref,
				Value:     pod.Name,
				Widget: func(w gtk.Widgetter, nv *adw.NavigationView) {
					switch row := w.(type) {
					case *adw.ActionRow:
						row.AddPrefix(widget.ObjectStatus(&pod).Icon())
					}
				},
			})
		}
		props = append(props, prop)
	case *appsv1.ReplicaSet:
		prop := &api.GroupProperty{Name: "Pods"}
		var pods corev1.PodList
		e.List(ctx, &pods, client.InNamespace(object.Namespace), client.MatchingLabels(object.Spec.Selector.MatchLabels))
		// TODO should we also filter pods by owner? takes one more api call to fetch replicasets
		for i, pod := range pods.Items {
			ref, _ := reference.GetReference(e.Scheme, &pod)
			prop.Children = append(prop.Children, &api.TextProperty{
				ID:        fmt.Sprintf("pods.%d", i),
				Reference: ref,
				Value:     pod.Name,
				Widget: func(w gtk.Widgetter, nv *adw.NavigationView) {
					switch row := w.(type) {
					case *adw.ActionRow:
						row.AddPrefix(widget.ObjectStatus(&pod).Icon())
					}
				},
			})
		}
		props = append(props, prop)
	case *appsv1.StatefulSet:
		podsProp := &api.GroupProperty{Name: "Pods"}
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
			ref, _ := reference.GetReference(e.Scheme, &pod)
			podsProp.Children = append(podsProp.Children, &api.TextProperty{
				ID:        fmt.Sprintf("pods.%d", i),
				Reference: ref,
				Value:     pod.Name,
				Widget: func(w gtk.Widgetter, nv *adw.NavigationView) {
					switch row := w.(type) {
					case *adw.ActionRow:
						row.AddPrefix(widget.ObjectStatus(&pod).Icon())
					}
				},
			})
		}
		props = append(props, podsProp)

		if len(object.Spec.VolumeClaimTemplates) > 0 {
			claimProp := &api.GroupProperty{Name: "Volume Claims"}
			for _, claim := range object.Spec.VolumeClaimTemplates {
				for replica := 0; replica < int(*object.Spec.Replicas); replica++ {
					e.SetObjectGVK(&claim)
					ref := corev1.ObjectReference{
						Kind:       claim.Kind,
						APIVersion: claim.APIVersion,
						Name:       fmt.Sprintf("%s-%s-%d", claim.Name, object.Name, replica),
						Namespace:  object.Namespace,
					}
					pv, _ := e.GetReference(ctx, ref)
					prop := &api.TextProperty{
						Reference: &ref,
						Value:     claim.Name,
						Widget: func(w gtk.Widgetter, nv *adw.NavigationView) {
							switch row := w.(type) {
							case *adw.ActionRow:
								if pv != nil {
									row.AddPrefix(widget.ObjectStatus(pv).Icon())
								}
							}
						},
					}
					claimProp.Children = append(claimProp.Children, prop)
				}
			}
			props = append(props, claimProp)
		}
	}

	return props
}
