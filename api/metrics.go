package api

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/imkira/go-observer/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Metrics struct {
	podMetrics  observer.Property[[]metricsv1beta1.PodMetrics]
	nodeMetrics observer.Property[[]metricsv1beta1.NodeMetrics]
	stopCh      chan struct{}
}

func newMetrics(client client.Client, resources []metav1.APIResource) (*Metrics, error) {
	if !metricsAPIAvailable(resources) {
		return nil, errors.New("no compatible metrics API detected")
	}

	m := Metrics{
		podMetrics:  observer.NewProperty([]metricsv1beta1.PodMetrics{}),
		nodeMetrics: observer.NewProperty([]metricsv1beta1.NodeMetrics{}),
		stopCh:      make(chan struct{}),
	}

	go func() {
		for {
			select {
			case <-m.stopCh:
				return
			default:
				var podMetricsList metricsv1beta1.PodMetricsList
				if err := client.List(context.TODO(), &podMetricsList); err != nil {
					log.Printf("unable to fetch pod metrics: %s", err.Error())
				}
				m.podMetrics.Update(podMetricsList.Items)

				var nodeMetricsList metricsv1beta1.NodeMetricsList
				if err := client.List(context.TODO(), &nodeMetricsList); err != nil {
					log.Printf("unable to fetch node metrics: %s", err.Error())
				}
				m.nodeMetrics.Update(nodeMetricsList.Items)

				time.Sleep(1 * time.Minute)
			}
		}
	}()

	return &m, nil
}

func (m *Metrics) stop() {
	close(m.stopCh)
}

func (m *Metrics) Pod(name types.NamespacedName) *metricsv1beta1.PodMetrics {
	for _, v := range m.podMetrics.Value() {
		if v.Name == name.Name && v.Namespace == name.Namespace {
			return &v
		}
	}
	return nil
}

func (m *Metrics) Node(name string) *metricsv1beta1.NodeMetrics {
	for _, v := range m.nodeMetrics.Value() {
		if v.Name == name {
			return &v
		}
	}
	return nil
}

func metricsAPIAvailable(resources []metav1.APIResource) bool {
	for _, res := range resources {
		if res.Group == metricsv1beta1.SchemeGroupVersion.Group && res.Version == metricsv1beta1.SchemeGroupVersion.Version {
			return true
		}
	}
	return false
}
