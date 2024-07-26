package api

import (
	"context"
	"time"

	"github.com/getseabird/seabird/internal/pubsub"
	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Events struct {
	events pubsub.Property[[]*eventsv1.Event]
}

func newEvents(ctx context.Context, clientset *kubernetes.Clientset) *Events {
	e := Events{
		events: pubsub.NewProperty([]*eventsv1.Event{}),
	}
	var events []*eventsv1.Event
	watchlist := cache.NewListWatchFromClient(clientset.EventsV1().RESTClient(), "events", v1.NamespaceAll,
		fields.Everything())
	_, controller := cache.NewInformer(watchlist, &eventsv1.Event{}, time.Minute*10,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(o interface{}) {
				switch obj := o.(type) {
				case *eventsv1.Event:
					events = append(events, obj)
					e.events.Pub(events)
				}
			},
			DeleteFunc: func(o interface{}) {
				switch obj := o.(type) {
				case *eventsv1.Event:
					for i, o := range events {
						if o.GetUID() == obj.GetUID() {
							events = append(events[:i], events[i+1:]...)
							e.events.Pub(events)
							break
						}
					}
				}

			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				switch obj := newObj.(type) {
				case *eventsv1.Event:
					for i, o := range events {
						if o.GetUID() == obj.GetUID() {
							events[i] = obj
							e.events.Pub(events)
							break
						}
					}
				}
			},
		},
	)

	go controller.Run(ctx.Done())

	return &e
}

func (e *Events) For(object client.Object) []*eventsv1.Event {
	var events []*eventsv1.Event
	for _, ev := range e.events.Value() {
		if ev.Regarding.UID == object.GetUID() {
			events = append(events, ev)
		}
	}
	return events
}
