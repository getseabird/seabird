package api

import (
	"context"
	"errors"
	"time"

	"github.com/getseabird/seabird/internal/pubsub"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Metrics struct {
	podMetrics  pubsub.Property[[]metricsv1beta1.PodMetrics]
	nodeMetrics pubsub.Property[[]metricsv1beta1.NodeMetrics]
}

func newMetrics(ctx context.Context, client client.Client, resources []metav1.APIResource) (*Metrics, error) {
	m := Metrics{
		podMetrics:  pubsub.NewProperty([]metricsv1beta1.PodMetrics{}),
		nodeMetrics: pubsub.NewProperty([]metricsv1beta1.NodeMetrics{}),
	}

	if !metricsAPIAvailable(resources) {
		return &m, errors.New("no compatible metrics API detected")
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var podMetricsList metricsv1beta1.PodMetricsList
				if err := client.List(ctx, &podMetricsList); err != nil {
					klog.Infof("unable to fetch pod metrics: %s", err.Error())
				}
				m.podMetrics.Pub(podMetricsList.Items)

				var nodeMetricsList metricsv1beta1.NodeMetricsList
				if err := client.List(ctx, &nodeMetricsList); err != nil {
					klog.Infof("unable to fetch node metrics: %s", err.Error())
				}
				m.nodeMetrics.Pub(nodeMetricsList.Items)

				time.Sleep(1 * time.Minute)
			}
		}
	}()

	return &m, nil
}

func (m *Metrics) Pod(name types.NamespacedName) *metricsv1beta1.PodMetrics {
	for _, v := range m.podMetrics.Value() {
		if v.Name == name.Name && v.Namespace == name.Namespace {
			return &v
		}
	}
	return nil
}

func (m *Metrics) PodSum(name types.NamespacedName) (*resource.Quantity, *resource.Quantity) {
	for _, metrics := range m.podMetrics.Value() {
		if metrics.Name == name.Name && metrics.Namespace == name.Namespace {
			mem := resource.NewQuantity(0, resource.DecimalSI)
			cpu := resource.NewQuantity(0, resource.DecimalSI)
			for _, container := range metrics.Containers {
				if m := container.Usage.Memory(); m != nil {
					mem.Add(*m)
				}
				if c := container.Usage.Cpu(); c != nil {
					cpu.Add(*c)
				}
			}

			return mem, cpu
		}
	}
	return nil, nil
}

func (m *Metrics) Container(pod types.NamespacedName, container string) *metricsv1beta1.ContainerMetrics {
	for _, v := range m.podMetrics.Value() {
		if v.Name == pod.Name && v.Namespace == pod.Namespace {
			for _, c := range v.Containers {
				if c.Name == container {
					return &c
				}
			}
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
