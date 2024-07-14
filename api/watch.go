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
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/klog/v2"
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

	updateProperty, _ := debounce.Debounce(func() {
		if opts.Property != nil {
			opts.Property.Update(objects)
		}
	}, 100*time.Millisecond, debounce.WithMaxWait(time.Second))
	defer updateProperty()

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
						cluster.SetObjectGVK(obj)
						objects = append(objects, obj.(T))
					}
					updateProperty()
					time.Sleep(time.Minute)
				}
			}
		}()
		return
	}

	go func() {
		backoff := flowcontrol.NewBackOff(time.Second, time.Minute)

	watch:
		for {
			backoff.Next(gvr.Resource, time.Now())
			time.Sleep(backoff.Get(gvr.Resource))
			w, err := cluster.DynamicClient.Resource(gvr).Watch(ctx, opts.ListOptions)
			if err != nil {
				klog.Infof("restarting watch: %s", err)
				continue
			}
			for {
				select {
				case res, ok := <-w.ResultChan():
					if !ok {
						klog.Infof("restarting watch: channel closed")
						continue watch
					}
					switch res.Type {
					case watch.Added:
						obj, err := objectFromUnstructured(cluster.Scheme, gvk, res.Object.(*unstructured.Unstructured))
						if err != nil {
							obj = res.Object.(client.Object)
						}
						cluster.SetObjectGVK(obj)
						if opts.AddFunc != nil {
							opts.AddFunc(obj.(T))
						}
						objects = append(objects, obj.(T))
						updateProperty()
					case watch.Modified:
						obj, err := objectFromUnstructured(cluster.Scheme, gvk, res.Object.(*unstructured.Unstructured))
						if err != nil {
							obj = res.Object.(client.Object)
						}
						cluster.SetObjectGVK(obj)
						if opts.UpdateFunc != nil {
							opts.UpdateFunc(obj.(T))
						}
						for i, o := range objects {
							if o.GetUID() == obj.GetUID() {
								objects[i] = obj.(T)
								break
							}
						}
						updateProperty()
					case watch.Deleted:
						obj, err := objectFromUnstructured(cluster.Scheme, gvk, res.Object.(*unstructured.Unstructured))
						if err != nil {
							obj = res.Object.(client.Object)
						}
						cluster.SetObjectGVK(obj)
						if opts.DeleteFunc != nil {
							opts.DeleteFunc(obj.(T))
						}
						for i, o := range objects {
							if o.GetUID() == obj.GetUID() {
								objects = slices.Delete(objects, i, i+1)
								break
							}
						}
						updateProperty()
					}
				case <-ctx.Done():
					w.Stop()
					return
				}
			}
		}
	}()
}
