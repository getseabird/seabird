package behavior

import (
	"context"

	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/extension"
	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/getseabird/seabird/internal/util"
	"github.com/go-logr/logr"
	"github.com/imkira/go-observer/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type ClusterBehavior struct {
	*Behavior
	*api.Cluster
	Extensions         []extension.Extension
	Namespaces         observer.Property[[]*corev1.Namespace]
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
	ctx = ctxt.With[*api.Cluster](ctx, clusterApi)

	cluster := ClusterBehavior{
		Behavior:         b,
		Cluster:          clusterApi,
		Namespaces:       observer.NewProperty([]*corev1.Namespace{}),
		SelectedResource: observer.NewProperty[*metav1.APIResource](nil),
		SearchText:       observer.NewProperty(""),
		SearchFilter:     observer.NewProperty(SearchFilter{}),
	}

	var ns *metav1.APIResource
	for _, r := range clusterApi.Resources {
		if r.Group == corev1.SchemeGroupVersion.Group && r.Version == corev1.SchemeGroupVersion.Version && r.Name == "namespaces" {
			ns = &r
			break
		}
	}
	util.ObjectWatcher(ctx, ns, cluster.Namespaces)

	for _, new := range extension.Extensions {
		cluster.Extensions = append(cluster.Extensions, new(clusterApi))
	}

	return &cluster, nil
}
