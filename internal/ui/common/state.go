package common

import (
	"context"

	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/extension"
	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/go-logr/logr"
	"github.com/imkira/go-observer/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type State struct {
	Preferences observer.Property[api.Preferences]
}

type ClusterState struct {
	*State
	*api.Cluster
	Extensions       []extension.Extension
	Namespaces       observer.Property[[]*corev1.Namespace]
	SelectedResource observer.Property[*metav1.APIResource]
	SearchText       observer.Property[string]
	SearchFilter     observer.Property[SearchFilter]
	SelectedObject   observer.Property[client.Object]
	Objects          observer.Property[[]client.Object]
}

func NewState() (*State, error) {
	prefs, err := api.LoadPreferences()
	if err != nil {
		return nil, err
	}
	prefs.Defaults()

	return &State{
		Preferences: observer.NewProperty(*prefs),
	}, nil
}

func (s *State) NewClusterState(ctx context.Context, clusterPrefs observer.Property[api.ClusterPreferences]) (*ClusterState, error) {
	logf.SetLogger(logr.Discard())

	cluster, err := api.NewCluster(ctx, clusterPrefs)
	if err != nil {
		return nil, err
	}
	ctx = ctxt.With[*api.Cluster](ctx, cluster)

	state := ClusterState{
		State:            s,
		Cluster:          cluster,
		Namespaces:       observer.NewProperty([]*corev1.Namespace{}),
		SelectedResource: observer.NewProperty[*metav1.APIResource](nil),
		SearchText:       observer.NewProperty(""),
		SearchFilter:     observer.NewProperty(SearchFilter{}),
		SelectedObject:   observer.NewProperty[client.Object](nil),
		Objects:          observer.NewProperty[[]client.Object](nil),
	}

	if err := api.InformerConnectProperty(ctx, cluster, schema.GroupVersionResource{Version: "v1", Resource: "namespaces"}, state.Namespaces); err != nil {
		klog.Errorf("watching namespaces: %v", err)
	}

	for _, new := range extension.Extensions {
		state.Extensions = append(state.Extensions, new(cluster))
	}

	return &state, nil
}
