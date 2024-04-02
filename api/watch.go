package api

import (
	"context"
	"log"
	"slices"
	"time"

	"github.com/getseabird/seabird/internal/util"
	"github.com/imkira/go-observer/v2"
	"github.com/zmwangx/debounce"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WatchOptions[T client.Object] struct {
	ListOptions metav1.ListOptions
	Property    observer.Property[[]T]
	AddFunc     func(T)
	UpdateFunc  func(T)
	DeleteFunc  func(T)
}

func Watch[T client.Object](ctx context.Context, cluster *Cluster, resource *metav1.APIResource, opts WatchOptions[T]) {
	objects := []T{}
	gvr := util.GVRForResource(resource)
	gvk := util.GVKForResource(resource)

	update, _ := debounce.Debounce(func() {
		for _, object := range objects {
			cluster.SetObjectGVK(object)
		}
		if opts.Property != nil {
			opts.Property.Update(objects)
		}
	}, 100*time.Millisecond, debounce.WithMaxWait(time.Second))
	defer update()

	if !slices.Contains(resource.Verbs, "watch") {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					list, err := cluster.DynamicClient.Resource(gvr).List(ctx, opts.ListOptions)
					if err != nil {
						log.Printf("list failed: %s", err.Error())
						continue
					}
					for _, i := range list.Items {
						obj, err := objectFromUnstructured(cluster.Scheme, gvk, &i)
						if err != nil {
							log.Printf("error converting obj: %s", err)
							continue
						}
						objects = append(objects, obj.(T))
					}
					update()
					time.Sleep(time.Minute)
				}
			}
		}()
		return
	}

	go func() {
		w, err := cluster.DynamicClient.Resource(gvr).Watch(ctx, opts.ListOptions)
		if err != nil {
			log.Printf("watch failed: %s", err.Error())
			return
		}
		for {
			select {
			case res := <-w.ResultChan():
				switch res.Type {
				case watch.Added:
					obj, err := objectFromUnstructured(cluster.Scheme, gvk, res.Object.(*unstructured.Unstructured))
					if err != nil {
						obj = res.Object.(client.Object)
					}
					if opts.AddFunc != nil {
						opts.AddFunc(obj.(T))
					}
					objects = append(objects, obj.(T))
					update()
				case watch.Modified:
					obj, err := objectFromUnstructured(cluster.Scheme, gvk, res.Object.(*unstructured.Unstructured))
					if err != nil {
						obj = res.Object.(client.Object)
					}
					if opts.UpdateFunc != nil {
						opts.UpdateFunc(obj.(T))
					}
					for i, o := range objects {
						if o.GetUID() == obj.GetUID() {
							objects[i] = obj.(T)
							break
						}
					}
					update()
				case watch.Deleted:
					obj, err := objectFromUnstructured(cluster.Scheme, gvk, res.Object.(*unstructured.Unstructured))
					if err != nil {
						obj = res.Object.(client.Object)
					}
					if opts.DeleteFunc != nil {
						opts.DeleteFunc(obj.(T))
					}
					for i, o := range objects {
						if o.GetUID() == obj.GetUID() {
							objects = slices.Delete(objects, i, i+1)
							break
						}
					}
					update()
				}
			case <-ctx.Done():
				w.Stop()
				return
			}
		}
	}()

	// TODO remove this. keeping it around for reference in case it's needed again
	// 	httpClient, err := rest.HTTPClientFor(cluster.Config)
	// if err != nil {
	// 	log.Printf("watcher httpClient: %s", err)
	// 	return
	// }
	// getter, err := apiutil.RESTClientForGVK(util.GVKForResource(resource), false, cluster.Config, serializer.NewCodecFactory(cluster.Scheme), httpClient)
	// watchlist := cache.NewListWatchFromClient(getter, resource.Name, corev1.NamespaceAll,
	// 	fields.Everything())
	// informer := cache.NewSharedInformer(watchlist, obj, 0)
	// informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
	// 	AddFunc: func(obj interface{}) {
	// 		objects = append(objects, obj.(T))
	// 		update()
	// 	},
	// 	DeleteFunc: func(o interface{}) {
	// 		obj := o.(client.Object)
	// 		for i, o := range objects {
	// 			if o.GetUID() == obj.GetUID() {
	// 				objects = append(objects[:i], objects[i+1:]...)
	// 				break
	// 			}
	// 		}
	// 		update()
	// 	},
	// 	UpdateFunc: func(oldObj, newObj interface{}) {
	// 		obj := newObj.(T)
	// 		for i, o := range objects {
	// 			if o.GetUID() == obj.GetUID() {
	// 				objects[i] = obj
	// 				break
	// 			}
	// 		}
	// 		update()
	// 	},
	// })
	// go informer.Run(ctx.Done())
}
