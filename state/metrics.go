package state

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/types"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type Metrics struct {
	cluster *Cluster
}

func NewMetrics(cluster *Cluster) (*Metrics, error) {
	if !metricsAPIAvailable(cluster) {
		return nil, errors.New("No compatible metrics API detected")
	}
	return &Metrics{cluster: cluster}, nil
}

func (m *Metrics) Pods(ctx context.Context) ([]metricsv1beta1.PodMetrics, error) {
	var list metricsv1beta1.PodMetricsList
	if err := m.cluster.List(ctx, &list); err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (m *Metrics) Pod(ctx context.Context, key types.NamespacedName) (*metricsv1beta1.PodMetrics, error) {
	var obj metricsv1beta1.PodMetrics
	if err := m.cluster.Get(ctx, key, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func metricsAPIAvailable(cluster *Cluster) bool {
	for _, res := range cluster.Resources {
		if res.Group == metricsv1beta1.SchemeGroupVersion.Group && res.Version == metricsv1beta1.SchemeGroupVersion.Version {
			return true
		}
	}
	return false
}
