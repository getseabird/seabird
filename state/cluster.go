package state

import (
	"context"
	"sort"

	"github.com/go-logr/logr"
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

type Cluster struct {
	client.Client
	Clientset   *kubernetes.Clientset
	Dynamic     *dynamic.DynamicClient
	Preferences *ClusterPreferences
	Scheme      *runtime.Scheme
	Resources   []metav1.APIResource
}

func NewCluster(ctx context.Context, prefs *ClusterPreferences) (*Cluster, error) {
	logf.SetLogger(logr.Discard())

	config := &rest.Config{
		Host:            prefs.Host,
		TLSClientConfig: prefs.TLS,
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

	cluster := Cluster{
		Client:      rclient,
		Clientset:   clientset,
		Preferences: prefs,
		Scheme:      scheme,
		Dynamic:     dynamicClient,
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
	sort.Slice(cluster.Resources, func(i, j int) bool {
		return cluster.Resources[i].Kind[0] < cluster.Resources[j].Kind[0]
	})

	return &cluster, nil

}
