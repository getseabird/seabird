package behavior

import (
	"context"
	"log"
	"sort"

	"github.com/go-logr/logr"
	"github.com/imkira/go-observer/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type ClusterBehavior struct {
	*Behavior

	client    client.Client
	clientset *kubernetes.Clientset
	dynamic   *dynamic.DynamicClient
	scheme    *runtime.Scheme

	ClusterPreferences observer.Property[ClusterPreferences]

	metrics *Metrics

	Resources  []metav1.APIResource
	Namespaces observer.Property[[]corev1.Namespace]

	SelectedResource observer.Property[*metav1.APIResource]
	SelectedObject   observer.Property[client.Object]

	SearchText   observer.Property[string]
	SearchFilter observer.Property[SearchFilter]
}

func (b *Behavior) WithCluster(ctx context.Context, clusterPrefs observer.Property[ClusterPreferences]) (*ClusterBehavior, error) {
	logf.SetLogger(logr.Discard())

	config := &rest.Config{
		Host:            clusterPrefs.Value().Host,
		TLSClientConfig: clusterPrefs.Value().TLS,
	}

	discovery, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
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

	var namespaces corev1.NamespaceList
	if err := rclient.List(context.TODO(), &namespaces); err != nil {
		return nil, err
	}

	cluster := ClusterBehavior{
		Behavior:           b,
		client:             rclient,
		clientset:          clientset,
		scheme:             scheme,
		ClusterPreferences: clusterPrefs,
		dynamic:            dynamicClient,
		Namespaces:         observer.NewProperty(namespaces.Items),
		SelectedObject:     observer.NewProperty[client.Object](nil),
		SelectedResource:   observer.NewProperty[*metav1.APIResource](nil),
		SearchText:         observer.NewProperty(""),
		SearchFilter:       observer.NewProperty(SearchFilter{}),
	}

	resources, err := discovery.ServerPreferredResources()
	if err != nil {
		return nil, err
	}
	for _, list := range resources {
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
			cluster.Resources = append(cluster.Resources, res)
		}
	}

	cluster.metrics, err = cluster.newMetrics(&cluster)
	if err != nil {
		log.Printf("metrics disabled: %s", err.Error())
	}

	sort.Slice(cluster.Resources, func(i, j int) bool {
		return cluster.Resources[i].Kind[0] < cluster.Resources[j].Kind[0]
	})

	return &cluster, nil

}
