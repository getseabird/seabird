package api

import (
	"github.com/imkira/go-observer/v2"
	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Events struct {
	events observer.Property[[]*eventsv1.Event]
	stopCh chan struct{}
}

func newEvents(clientset *kubernetes.Clientset) *Events {
	e := Events{
		events: observer.NewProperty([]*eventsv1.Event{}),
		stopCh: make(chan struct{}),
	}
	var event eventsv1.Event
	var events []*eventsv1.Event
	watchlist := cache.NewListWatchFromClient(clientset.EventsV1().RESTClient(), "events", v1.NamespaceAll,
		fields.Everything())
	_, controller := cache.NewInformer(watchlist, &event, 0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				events = append(events, obj.(*eventsv1.Event))
				e.events.Update(events)
			},
			DeleteFunc: func(o interface{}) {
				obj := o.(*eventsv1.Event)
				for i, o := range events {
					if o.GetUID() == obj.GetUID() {
						events = append(events[:i], events[i+1:]...)
						e.events.Update(events)
						break
					}
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				obj := newObj.(*eventsv1.Event)
				for i, o := range events {
					if o.GetUID() == obj.GetUID() {
						events[i] = obj
						e.events.Update(events)
						break
					}
				}
			},
		},
	)

	go controller.Run(e.stopCh)

	return &e
}

func (e *Events) stop() {
	close(e.stopCh)
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
