package extension

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	Extensions = append(Extensions, func(cluster *api.Cluster) Extension {
		return &Meta{Cluster: cluster}
	})
}

type Meta struct {
	*api.Cluster
}

func (e *Meta) CreateColumns(ctx context.Context, resource *metav1.APIResource, columns []api.Column) []api.Column {
	columns = append(columns, api.Column{
		Name:     "Name",
		Priority: 100,
		Bind: func(listitem *gtk.ListItem, object client.Object) {
			label := gtk.NewLabel(object.GetName())
			label.SetHAlign(gtk.AlignStart)
			listitem.SetChild(label)
		},
		Compare: func(a, b client.Object) int {
			return strings.Compare(a.GetName(), b.GetName())
		},
	})

	if resource.Namespaced {
		columns = append(columns, api.Column{
			Name:     "Namespace",
			Priority: 90,
			Bind: func(listitem *gtk.ListItem, object client.Object) {
				label := gtk.NewLabel(object.GetNamespace())
				label.SetHAlign(gtk.AlignStart)
				listitem.SetChild(label)
			},
			Compare: func(a, b client.Object) int {
				return strings.Compare(a.GetNamespace(), b.GetNamespace())
			},
		})
	}

	columns = append(columns, api.Column{
		Name:     "Age",
		Priority: 80,
		Bind: func(listitem *gtk.ListItem, object client.Object) {
			duration := time.Since(object.GetCreationTimestamp().Time)
			label := gtk.NewLabel(util.HumanizeApproximateDuration(duration))
			label.SetHAlign(gtk.AlignStart)
			listitem.SetChild(label)
		},
		Compare: func(a, b client.Object) int {
			return a.GetCreationTimestamp().Compare(b.GetCreationTimestamp().Time)
		},
	})

	return columns
}

func (e *Meta) CreateObjectProperties(ctx context.Context, object client.Object, props []api.Property) []api.Property {
	var labels []api.Property
	for key, value := range object.GetLabels() {
		labels = append(labels, &api.TextProperty{Name: key, Value: value})
	}
	var annotations []api.Property
	for key, value := range object.GetAnnotations() {
		annotations = append(annotations, &api.TextProperty{Name: key, Value: value})
	}
	var owners []api.Property
	for _, ref := range object.GetOwnerReferences() {
		owners = append(owners, &api.TextProperty{
			Name:  fmt.Sprintf("%s %s", ref.APIVersion, ref.Kind),
			Value: ref.Name,
			Reference: &corev1.ObjectReference{
				APIVersion: ref.APIVersion,
				Kind:       ref.Kind,
				Name:       ref.Name,
				UID:        ref.UID,
				Namespace:  object.GetNamespace(),
			},
		})
	}

	props = append(props,
		&api.GroupProperty{
			Priority: 100,
			Name:     "Metadata",
			Children: []api.Property{
				&api.TextProperty{
					Name:  "Name",
					Value: object.GetName(),
				},
				&api.TextProperty{
					Name:  "Namespace",
					Value: object.GetNamespace(),
				},
				&api.GroupProperty{
					Name:     "Labels",
					Children: labels,
				},
				&api.GroupProperty{
					Name:     "Annotations",
					Children: annotations,
				},
				&api.GroupProperty{
					Name:     "Owners",
					Children: owners,
				},
			},
		})

	events := &api.GroupProperty{Name: "Events", Priority: -100}
	for _, ev := range e.Events.For(object) {
		eventTime := ev.EventTime.Time
		if eventTime.IsZero() {
			eventTime = ev.CreationTimestamp.Time
		}
		events.Children = append(events.Children, &api.TextProperty{
			Name:  eventTime.Format(time.RFC3339),
			Value: ev.Note,
		})
	}
	if len(events.Children) > 0 {
		props = append(props, events)
	}

	return props
}
