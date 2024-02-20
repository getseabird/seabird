package behavior

import (
	"context"

	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/extension"
	"github.com/go-logr/logr"
	"github.com/imkira/go-observer/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type ClusterBehavior struct {
	*Behavior
	*api.Cluster
	extensions         []extension.Extension
	Namespaces         observer.Property[[]corev1.Namespace]
	SelectedResource   observer.Property[*metav1.APIResource]
	SearchText         observer.Property[string]
	SearchFilter       observer.Property[SearchFilter]
	RootDetailBehavior *DetailBehavior
}

func (b *Behavior) WithCluster(ctx context.Context, clusterPrefs observer.Property[api.ClusterPreferences]) (*ClusterBehavior, error) {
	logf.SetLogger(logr.Discard())

	clusterApi, err := api.NewCluster(ctx, clusterPrefs)
	if err != nil {
		return nil, err
	}

	var namespaces corev1.NamespaceList
	if err := clusterApi.List(context.TODO(), &namespaces); err != nil {
		return nil, err
	}

	cluster := ClusterBehavior{
		Behavior:         b,
		Cluster:          clusterApi,
		Namespaces:       observer.NewProperty(namespaces.Items),
		SelectedResource: observer.NewProperty[*metav1.APIResource](nil),
		SearchText:       observer.NewProperty(""),
		SearchFilter:     observer.NewProperty(SearchFilter{}),
	}

	for _, new := range extension.Extensions {
		cluster.extensions = append(cluster.extensions, new(clusterApi))
	}

	return &cluster, nil

}
