package api

import (
	"context"
	"errors"
	"log"
	"reflect"
	"slices"
	"sort"
	"time"

	"github.com/getseabird/seabird/internal/util"
	"github.com/imkira/go-observer/v2"
	"github.com/zmwangx/debounce"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type Cluster struct {
	client.Client
	*kubernetes.Clientset
	Config                 *rest.Config
	ClusterPreferences     observer.Property[ClusterPreferences]
	Metrics                *Metrics
	Events                 *Events
	RESTMapper             meta.RESTMapper
	DynamicClient          *dynamic.DynamicClient
	Scheme                 *runtime.Scheme
	Encoder                *Encoder
	Resources              []metav1.APIResource
	ctx                    context.Context
	informerFactory        informers.SharedInformerFactory
	dynamicInformerFactory dynamicinformer.DynamicSharedInformerFactory
	sharedInformers        map[schema.GroupVersionResource]informers.GenericInformer
}

func NewCluster(ctx context.Context, clusterPrefs observer.Property[ClusterPreferences]) (*Cluster, error) {
	config := &rest.Config{
		Host:            clusterPrefs.Value().Host,
		BearerToken:     clusterPrefs.Value().BearerToken,
		TLSClientConfig: clusterPrefs.Value().TLS,
		ExecProvider:    clusterPrefs.Value().Exec,
	}

	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	networkingv1.AddToScheme(scheme)
	apiextensionsv1.AddToScheme(scheme)
	appsv1.AddToScheme(scheme)
	rbacv1.AddToScheme(scheme)
	storagev1.AddToScheme(scheme)
	eventsv1.AddToScheme(scheme)
	batchv1.AddToScheme(scheme)
	metricsv1beta1.AddToScheme(scheme)

	rclient, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	res, err := restmapper.GetAPIGroupResources(clientset.Discovery())
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDiscoveryRESTMapper(res)

	informerFactory := informers.NewSharedInformerFactory(clientset, time.Hour)
	informerFactory.Start(ctx.Done())
	dynamicInformerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, time.Hour)
	dynamicInformerFactory.Start(ctx.Done())

	var resources []metav1.APIResource
	preferredResources, err := clientset.Discovery().ServerPreferredResources()
	if err != nil {
		var groupDiscoveryFailed *discovery.ErrGroupDiscoveryFailed
		if errors.As(err, &groupDiscoveryFailed) {
			for api, err := range groupDiscoveryFailed.Groups {
				// TODO display as toast
				log.Printf("group discovery failed for '%s': %s", api.String(), err.Error())
			}
		} else {
			return nil, err
		}
	}
	for _, list := range preferredResources {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			return nil, err
		}
		for _, res := range list.APIResources {
			if res.Group == "" {
				res.Group = gv.Group
			}
			if res.Version == "" {
				res.Version = gv.Version
			}

			if !slices.Contains(res.Verbs, "get") || !slices.Contains(res.Verbs, "list") {
				continue
			}

			resources = append(resources, res)
		}
	}

	metrics, err := newMetrics(ctx, rclient, resources)
	if err != nil {
		log.Printf("metrics disabled: %s", err.Error())
	}

	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Kind[0] < resources[j].Kind[0]
	})

	cluster := Cluster{
		Client:                 rclient,
		Config:                 config,
		Clientset:              clientset,
		RESTMapper:             mapper,
		Scheme:                 scheme,
		Encoder:                &Encoder{Scheme: scheme},
		ClusterPreferences:     clusterPrefs,
		DynamicClient:          dynamicClient,
		Metrics:                metrics,
		Events:                 newEvents(ctx, clientset),
		ctx:                    ctx,
		Resources:              resources,
		informerFactory:        informerFactory,
		dynamicInformerFactory: dynamicInformerFactory,
		sharedInformers:        map[schema.GroupVersionResource]informers.GenericInformer{},
	}

	return &cluster, nil
}

func (cluster *Cluster) GetReference(ctx context.Context, ref corev1.ObjectReference) (client.Object, error) {
	var object client.Object
	gvk := schema.FromAPIVersionAndKind(ref.APIVersion, ref.Kind).String()
	for key, t := range cluster.Scheme.AllKnownTypes() {
		if key.String() == gvk {
			object = reflect.New(t).Interface().(client.Object)
			break
		}
	}

	if err := cluster.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, object); err != nil {
		return nil, err
	}

	if err := cluster.SetObjectGVK(object); err != nil {
		klog.Infof("Cluster/SetObjectGVK: %s", err)
	}

	return object, nil
}

func (cluster *Cluster) GetAPIResource(gvk schema.GroupVersionKind) *metav1.APIResource {
	for _, res := range cluster.Resources {
		if util.GVKEquals(gvk, util.GVKForResource(&res)) {
			return &res
		}
	}
	return nil
}

func (cluster *Cluster) SetObjectGVK(object client.Object) error {
	gvk, err := apiutil.GVKForObject(object, cluster.Scheme)
	if err != nil {
		return err
	}
	object.GetObjectKind().SetGroupVersionKind(gvk)
	return nil
}

func (c *Cluster) GetInformer(gvr schema.GroupVersionResource) informers.GenericInformer {
	if informer, ok := c.sharedInformers[gvr]; ok {
		return informer
	}
	var informer informers.GenericInformer
	informer, err := c.informerFactory.ForResource(gvr)
	if err != nil {
		informer = c.dynamicInformerFactory.ForResource(gvr)
	}
	informer.Informer().SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		if apierrors.IsMethodNotSupported(err) {
			return
		}
		klog.Errorf("%s informer: %s", gvr.Resource, err)
	})
	informer.Informer().SetTransform(func(obj interface{}) (interface{}, error) {
		switch obj := obj.(type) {
		case *unstructured.Unstructured:
			return obj, nil
		case client.Object:
			err := c.SetObjectGVK(obj)
			return obj, err
		default:
			return obj, nil
		}
	})
	go informer.Informer().Run(c.ctx.Done())
	c.sharedInformers[gvr] = informer
	return informer
}

func (c *Cluster) AddInformerEventHandler(ctx context.Context, gvr schema.GroupVersionResource, handler cache.ResourceEventHandler) error {
	informer := c.GetInformer(gvr)
	registration, err := informer.Informer().AddEventHandler(handler)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		informer.Informer().RemoveEventHandler(registration)
	}()
	return nil
}

func InformerConnectProperty[T client.Object](ctx context.Context, cluster *Cluster, gvr schema.GroupVersionResource, prop observer.Property[[]T]) error {
	updateProperty, _ := debounce.Debounce(func() {
		var objects []T
		err := cache.ListAll(cluster.GetInformer(gvr).Informer().GetIndexer(), labels.Everything(), func(m interface{}) {
			objects = append(objects, m.(T))
		})
		if err != nil {
			klog.Warning("list all: %v", err)
			return
		}
		prop.Update(objects)
	}, 100*time.Millisecond)
	defer updateProperty()

	return cluster.AddInformerEventHandler(ctx, gvr, cache.ResourceEventHandlerFuncs{
		AddFunc: func(_ interface{}) {
			updateProperty()
		},
		UpdateFunc: func(_, _ interface{}) {
			updateProperty()
		},
		DeleteFunc: func(_ interface{}) {
			updateProperty()
		},
	})
}
