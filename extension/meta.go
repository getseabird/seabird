package extension

import (
	"fmt"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
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

func (e *Meta) CreateObjectProperties(object client.Object, props []api.Property) []api.Property {
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
		owners = append(owners, &api.TextProperty{Name: fmt.Sprintf("%s %s", ref.APIVersion, ref.Kind), Value: ref.Name})
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
			Widget: func(w gtk.Widgetter, _ *adw.NavigationView) {
				button := gtk.NewMenuButton()
				button.SetIconName("view-more-symbolic")
				button.AddCSSClass("flat")
				model := gio.NewMenu()
				model.Append("Delete", "detail.delete")
				button.SetPopover(gtk.NewPopoverMenuFromModel(model))
				w.(*adw.PreferencesGroup).SetHeaderSuffix(button)
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
