package behavior

import (
	"context"
	"errors"

	"github.com/imkira/go-observer/v2"
	"k8s.io/apimachinery/pkg/types"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type Metrics struct {
	PodMetrics observer.Property[[]metricsv1beta1.PodMetrics]
}

func (b *ClusterBehavior) newMetrics(cluster *ClusterBehavior) (*Metrics, error) {
	if !metricsAPIAvailable(cluster) {
		return nil, errors.New("No compatible metrics API detected")
	}

	var list metricsv1beta1.PodMetricsList
	if err := cluster.client.List(context.TODO(), &list); err != nil {
		return nil, err
	}

	m := Metrics{
		PodMetrics: observer.NewProperty(list.Items),
	}

	return &m, nil
}

func (m *Metrics) PodValue(name types.NamespacedName) *metricsv1beta1.PodMetrics {
	for _, v := range m.PodMetrics.Value() {
		if v.Name == name.Name && v.Namespace == name.Namespace {
			return &v
		}
	}
	return nil
}

func metricsAPIAvailable(cluster *ClusterBehavior) bool {
	for _, res := range cluster.Resources {
		if res.Group == metricsv1beta1.SchemeGroupVersion.Group && res.Version == metricsv1beta1.SchemeGroupVersion.Version {
			return true
		}
	}
	return false
}
