package api

import (
	"context"
	"log"
	"slices"
	"time"

	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/getseabird/seabird/internal/util"
	"github.com/imkira/go-observer/v2"
	"github.com/zmwangx/debounce"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ObjectWatcher[T client.Object](ctx context.Context, resource *metav1.APIResource, prop observer.Property[[]T]) {
	cluster, _ := ctxt.From[*Cluster](ctx)
	objects := []T{}

	update, _ := debounce.Debounce(func() {
		for _, object := range objects {
			cluster.SetObjectGVK(object)
		}
		prop.Update(objects)
	}, 100*time.Millisecond, debounce.WithMaxWait(time.Second))
	defer update()

	if !slices.Contains(resource.Verbs, "watch") {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					list, err := cluster.DynamicClient.Resource(util.ResourceGVR(resource)).List(ctx, metav1.ListOptions{})
					if err != nil {
						log.Printf("list failed: %s", err.Error())
						continue
					}
					for _, i := range list.Items {
						objects = append(objects, client.Object(&i).(T))
					}
					update()
					time.Sleep(time.Minute)
				}
			}
		}()
		return
	}

	obj, _ := cluster.Scheme.New(util.ResourceGVK(resource))

	// TODO there's probably a better way? create rest client from scheme?
	var getter cache.Getter
	switch util.ResourceGVR(resource).GroupVersion().String() {
	case corev1.SchemeGroupVersion.String():
		getter = cluster.CoreV1().RESTClient()
	case appsv1.SchemeGroupVersion.String():
		getter = cluster.AppsV1().RESTClient()
	case apiextensions.SchemeGroupVersion.String():
		getter = cluster.ExtensionsV1beta1().RESTClient()
	}

	if obj == nil || getter == nil {
		go func() {
			w, err := cluster.DynamicClient.Resource(util.ResourceGVR(resource)).Watch(ctx, metav1.ListOptions{})
			if err != nil {
				log.Printf("watch failed: %s", err.Error())
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
				case <-ctx.Done():
					w.Stop()
					return
				}
			}

		}()
		return
	}

	watchlist := cache.NewListWatchFromClient(getter, resource.Name, corev1.NamespaceAll,
		fields.Everything())
	informer := cache.NewSharedInformer(watchlist, obj, 0)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
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
	})
	go informer.Run(ctx.Done())
}
