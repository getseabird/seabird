package extension

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	Extensions = append(Extensions, NewMeta)
}

func NewMeta(_ context.Context, cluster *api.Cluster) (Extension, error) {
	return &Meta{Cluster: cluster}, nil
}

type Meta struct {
	Noop
	*api.Cluster
}

func (e *Meta) CreateColumns(ctx context.Context, resource *metav1.APIResource, columns []api.Column) []api.Column {
	columns = append(columns, api.Column{
		Name:     "Name",
		Priority: 100,
		Bind: func(cell api.Cell, object client.Object) {
			cell.SetLabel(object.GetName())
		},
		Compare: func(a, b client.Object) int {
			return strings.Compare(a.GetName(), b.GetName())
		},
	})

	if resource.Namespaced {
		columns = append(columns, api.Column{
			Name:     "Namespace",
			Priority: 90,
			Bind: func(cell api.Cell, object client.Object) {
				cell.SetLabel(object.GetNamespace())
			},
			Compare: func(a, b client.Object) int {
				return strings.Compare(a.GetNamespace(), b.GetNamespace())
			},
		})
	}

	columns = append(columns, api.Column{
		Name:     "Age",
		Priority: 80,
		Bind: func(cell api.Cell, object client.Object) {
			duration := time.Since(object.GetCreationTimestamp().Time)
			cell.SetLabel(util.HumanizeApproximateDuration(duration))
		},
		Compare: func(a, b client.Object) int {
			return a.GetCreationTimestamp().Compare(b.GetCreationTimestamp().Time)
		},
	})

	return columns
}

func (e *Meta) CreateObjectProperties(ctx context.Context, resource *metav1.APIResource, object client.Object, props []api.Property) []api.Property {
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

	group := object.GetObjectKind().GroupVersionKind().Group
	if len(group) == 0 {
		group = "k8s.io"
	}

	metadata := api.GroupProperty{
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
			&api.TextProperty{
				Name:  "Created",
				Value: object.GetCreationTimestamp().Format(time.RFC822),
			},
			// &api.TextProperty{
			// 	Name:  "Kind",
			// 	Value: object.GetObjectKind().GroupVersionKind().Kind,
			// },
			// &api.TextProperty{
			// 	Name:  "Group",
			// 	Value: group,
			// },
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
	}
	if !resource.Namespaced {
		metadata.Children = slices.Delete(metadata.Children, 1, 2)
	}
	props = append(props, &metadata)

	events := &api.GroupProperty{Name: "Events", Priority: -100}
	for _, ev := range e.Events.For(object) {
		eventTime := ev.EventTime.Time
		if eventTime.IsZero() {
			eventTime = ev.CreationTimestamp.Time
		}
		events.Children = append(events.Children, &api.TextProperty{
			Name:  eventTime.Format(time.RFC822),
			Value: ev.Note,
		})
	}
	if len(events.Children) > 0 {
		props = append(props, events)
	}

	return props
}
