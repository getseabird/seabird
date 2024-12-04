package common

import (
	"context"

	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/extension"
	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/getseabird/seabird/internal/pubsub"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type State struct {
	Preferences pubsub.Property[api.Preferences]
}

type ClusterState struct {
	*State
	*api.Cluster
	Extensions       []extension.Extension
	Namespaces       pubsub.Property[[]*corev1.Namespace]
	SelectedResource pubsub.Property[*metav1.APIResource]
	SearchText       pubsub.Property[string]
	SearchFilter     pubsub.Property[SearchFilter]
	SelectedObject   pubsub.Property[client.Object]
	Objects          pubsub.Property[[]client.Object]
}

func NewState() (*State, error) {
	prefs, err := api.LoadPreferences()
	if err != nil {
		return nil, err
	}
	prefs.Defaults()

	return &State{
		Preferences: pubsub.NewProperty(*prefs),
	}, nil
}

func (s *State) NewClusterState(ctx context.Context, clusterPrefs pubsub.Property[api.ClusterPreferences]) (*ClusterState, error) {
	logf.SetLogger(logr.Discard())

	cluster, err := api.NewCluster(ctx, clusterPrefs)
	if err != nil {
		return nil, err
	}
	ctx = ctxt.With[*api.Cluster](ctx, cluster)

	selected := &cluster.Resources[0]
	for _, r := range cluster.Resources {
		if r.Name == "pods" && r.Group == "" {
			selected = &r
			break
		}
	}

	state := ClusterState{
		State:            s,
		Cluster:          cluster,
		Namespaces:       pubsub.NewProperty([]*corev1.Namespace{}),
		SelectedResource: pubsub.NewProperty[*metav1.APIResource](selected),
		SearchText:       pubsub.NewProperty(""),
		SearchFilter:     pubsub.NewProperty(SearchFilter{}),
		SelectedObject:   pubsub.NewProperty[client.Object](nil),
		Objects:          pubsub.NewProperty[[]client.Object](nil),
	}

	if err := api.InformerConnectProperty(ctx, cluster, schema.GroupVersionResource{Version: "v1", Resource: "namespaces"}, state.Namespaces); err != nil {
		klog.Errorf("watching namespaces: %v", err)
	}

	for _, new := range extension.Extensions {
		ext, err := new(ctx, cluster)
		if err != nil {
			klog.Error("failed to load extension: %s", err)
		}
		state.Extensions = append(state.Extensions, ext)
	}

	return &state, nil
}
