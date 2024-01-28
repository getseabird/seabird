package behavior

import (
	"context"
	"time"

	"github.com/getseabird/seabird/util"
	"github.com/imkira/go-observer/v2"
	"github.com/zmwangx/debounce"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ListBehavior struct {
	*ClusterBehavior
	Objects   observer.Property[[]client.Object]
	objects   []client.Object
	resource  *metav1.APIResource
	stopWatch chan struct{}
}

func (b *ClusterBehavior) NewListBehavior() *ListBehavior {
	listView := ListBehavior{
		ClusterBehavior: b,
		Objects:         observer.NewProperty[[]client.Object](nil),
	}

	onChange(listView.SelectedResource, listView.onSelectedResourceChange)

	return &listView
}

// We want typed objects for known resources so we can type switch them
func (b *ListBehavior) onSelectedResourceChange(resource *metav1.APIResource) {
	if b.stopWatch != nil {
		close(b.stopWatch)
	}
	b.stopWatch = make(chan struct{})
	b.objects = []client.Object{}
	update, _ := debounce.Debounce(func() {
		b.Objects.Update(b.objects)
	}, 100*time.Millisecond, debounce.WithMaxWait(time.Second))
	defer update()

	gvr := util.ResourceGVR(resource)

	var obj runtime.Object
	switch gvr.String() {
	case corev1.SchemeGroupVersion.WithResource("pods").String():
		obj = &corev1.Pod{}
	case corev1.SchemeGroupVersion.WithResource("configmaps").String():
		obj = &corev1.ConfigMap{}
	case corev1.SchemeGroupVersion.WithResource("secrets").String():
		obj = &corev1.Secret{}
	case corev1.SchemeGroupVersion.WithResource("services").String():
		obj = &corev1.Service{}
	case corev1.SchemeGroupVersion.WithResource("persistentvolumeclaims").String():
		obj = &corev1.PersistentVolumeClaim{}
	case corev1.SchemeGroupVersion.WithResource("nodes").String():
		obj = &corev1.Node{}
	case appsv1.SchemeGroupVersion.WithResource("deployments").String():
		obj = &appsv1.Deployment{}
	case appsv1.SchemeGroupVersion.WithResource("statefulsets").String():
		obj = &appsv1.StatefulSet{}
	default:
		go func() {
			w, _ := b.dynamic.Resource(gvr).Watch(context.TODO(), metav1.ListOptions{})
			for {
				select {
				case res := <-w.ResultChan():
					switch res.Type {
					case watch.Added:
						b.objects = append(b.objects, res.Object.(client.Object))
						update()
					case watch.Modified:
						obj := res.Object.(client.Object)
						for i, o := range b.objects {
							if o.GetUID() == obj.GetUID() {
								b.objects[i] = obj
								break
							}
						}
						update()
					case watch.Deleted:
						obj := res.Object.(client.Object)
						for i, o := range b.objects {
							if o.GetUID() == obj.GetUID() {
								b.objects = append(b.objects[:i], b.objects[i+1:]...)
								break
							}
						}
						update()
					}
				case <-b.stopWatch:
					w.Stop()
					return
				default:
				}
			}

		}()
		return
	}

	var getter cache.Getter
	switch (metav1.GroupVersion{Group: gvr.Group, Version: gvr.Version}).String() {
	case corev1.SchemeGroupVersion.String():
		getter = b.clientset.CoreV1().RESTClient()
	case appsv1.SchemeGroupVersion.String():
		getter = b.clientset.AppsV1().RESTClient()
	}

	watchlist := cache.NewListWatchFromClient(getter, gvr.Resource, v1.NamespaceAll,
		fields.Everything())
	_, controller := cache.NewInformer(watchlist, obj, 0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				b.objects = append(b.objects, obj.(client.Object))
				update()
			},
			DeleteFunc: func(o interface{}) {
				obj := o.(client.Object)
				for i, o := range b.objects {
					if o.GetUID() == obj.GetUID() {
						b.objects = append(b.objects[:i], b.objects[i+1:]...)
						break
					}
				}
				update()
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				obj := newObj.(client.Object)
				for i, o := range b.objects {
					if o.GetUID() == obj.GetUID() {
						b.objects[i] = obj
						break
					}
				}
				update()
			},
		},
	)
	go controller.Run(b.stopWatch)
}
