package util

import (
	"context"
	"log"
	"time"

	"github.com/getseabird/seabird/api"
	"github.com/imkira/go-observer/v2"
	"github.com/zmwangx/debounce"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ObjectWatcher[T client.Object](cluster *api.Cluster, gvr schema.GroupVersionResource, stopWatch chan struct{}, prop observer.Property[[]T]) {
	objects := []T{}
	update, _ := debounce.Debounce(func() {
		prop.Update(objects)
	}, 100*time.Millisecond, debounce.WithMaxWait(time.Second))
	defer update()

	var obj runtime.Object
	switch gvr.String() {
	case corev1.SchemeGroupVersion.WithResource("namespaces").String():
		obj = &corev1.Namespace{}
	case corev1.SchemeGroupVersion.WithResource("pods").String():
		obj = &corev1.Pod{}
	case corev1.SchemeGroupVersion.WithResource("configmaps").String():
		obj = &corev1.ConfigMap{}
	case corev1.SchemeGroupVersion.WithResource("secrets").String():
		obj = &corev1.Secret{}
	case corev1.SchemeGroupVersion.WithResource("services").String():
		obj = &corev1.Service{}
	case corev1.SchemeGroupVersion.WithResource("persistentvolumes").String():
		obj = &corev1.PersistentVolume{}
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
			w, err := cluster.DynamicClient.Resource(gvr).Watch(context.TODO(), metav1.ListOptions{})
			if err != nil {
				log.Printf("watch failed: %s", err.Error())
				list, err := cluster.DynamicClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					log.Printf("list failed: %s", err.Error())
					return
				}
				for _, i := range list.Items {
					objects = append(objects, client.Object(&i).(T))
				}
				return
			}
			for {
				select {
				case res := <-w.ResultChan():
					switch res.Type {
					case watch.Added:
						objects = append(objects, res.Object.(T))
						update()
					case watch.Modified:
						obj := res.Object.(T)
						for i, o := range objects {
							if o.GetUID() == obj.GetUID() {
								objects[i] = obj
								break
							}
						}
						update()
					case watch.Deleted:
						obj := res.Object.(T)
						for i, o := range objects {
							if o.GetUID() == obj.GetUID() {
								objects = append(objects[:i], objects[i+1:]...)
								break
							}
						}
						update()
					}
				case <-stopWatch:
					w.Stop()
					return
				}
			}

		}()
		return
	}

	var getter cache.Getter
	switch (metav1.GroupVersion{Group: gvr.Group, Version: gvr.Version}).String() {
	case corev1.SchemeGroupVersion.String():
		getter = cluster.CoreV1().RESTClient()
	case appsv1.SchemeGroupVersion.String():
		getter = cluster.AppsV1().RESTClient()
	}

	watchlist := cache.NewListWatchFromClient(getter, gvr.Resource, corev1.NamespaceAll,
		fields.Everything())
	_, controller := cache.NewInformer(watchlist, obj, 0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				objects = append(objects, obj.(T))
				update()
			},
			DeleteFunc: func(o interface{}) {
				obj := o.(client.Object)
				for i, o := range objects {
					if o.GetUID() == obj.GetUID() {
						objects = append(objects[:i], objects[i+1:]...)
						break
					}
				}
				update()
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				obj := newObj.(T)
				for i, o := range objects {
					if o.GetUID() == obj.GetUID() {
						objects[i] = obj
						break
					}
				}
				update()
			},
		},
	)
	go controller.Run(stopWatch)
}
