package state

import (
	"context"
	"path/filepath"
	"sort"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Cluster struct {
	client.Client
	Resources []metav1.APIResource
}

func NewCluster(ctx context.Context) (*Cluster, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	discovery, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	corev1.AddToScheme(scheme.Scheme)
	apiextensionsv1.AddToScheme(scheme.Scheme)

	rclient, err := client.New(config, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return nil, err
	}

	cluster := Cluster{
		Client: rclient,
	}

	resources, err := discovery.ServerPreferredResources()
	if err != nil {
		return nil, err
	}
	for _, list := range resources {
		cluster.Resources = append(cluster.Resources, list.APIResources...)
	}
	sort.Slice(cluster.Resources, func(i, j int) bool {
		return cluster.Resources[i].Kind[0] < cluster.Resources[j].Kind[0]
	})

	return &cluster, nil

}
