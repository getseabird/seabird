package api

import (
	"context"
	"errors"
	"log"
	"slices"
	"sort"

	"github.com/imkira/go-observer/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Cluster struct {
	client.Client
	*kubernetes.Clientset
	Config             *rest.Config
	ClusterPreferences observer.Property[ClusterPreferences]
	Metrics            *Metrics
	Events             *Events
	RESTMapper         meta.RESTMapper
	DynamicClient      *dynamic.DynamicClient
	Scheme             *runtime.Scheme
	Encoder            *Encoder
	Resources          []metav1.APIResource
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
	apiextensionsv1.AddToScheme(scheme)
	appsv1.AddToScheme(scheme)
	rbacv1.AddToScheme(scheme)
	storagev1.AddToScheme(scheme)
	eventsv1.AddToScheme(scheme)
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
		Client:             rclient,
		Config:             config,
		Clientset:          clientset,
		RESTMapper:         mapper,
		Scheme:             scheme,
		Encoder:            &Encoder{Scheme: scheme},
		ClusterPreferences: clusterPrefs,
		DynamicClient:      dynamicClient,
		Metrics:            metrics,
		Events:             newEvents(ctx, clientset),
		Resources:          resources,
	}

	return &cluster, nil
}
