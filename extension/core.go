package extension

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/getseabird/seabird/api"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	Extensions = append(Extensions, func(cluster *api.Cluster) Extension {
		return &Core{Cluster: cluster}
	})
}

type Core struct {
	*api.Cluster
}

func (e *Core) CreateObjectProperties(object client.Object, props []api.Property) []api.Property {
	switch object := object.(type) {
	case *corev1.Pod:
		var containers []api.Property

		var podMetrics *metricsv1beta1.PodMetrics
		if e.Metrics != nil {
			podMetrics = e.Metrics.Pod(types.NamespacedName{Name: object.Name, Namespace: object.Namespace})
		}

		for i, container := range object.Spec.Containers {
			var props []api.Property
			var status corev1.ContainerStatus
			for _, s := range object.Status.ContainerStatuses {
				if s.Name == container.Name {
					status = s
					break
				}
			}

			var metrics *metricsv1beta1.ContainerMetrics
			if podMetrics != nil {
				for _, m := range podMetrics.Containers {
					if m.Name == container.Name {
						metrics = &m
						break
					}
				}
			}

			var state string
			if status.State.Running != nil {
				state = "Running"
			} else if status.State.Terminated != nil {
				message := status.State.Terminated.Message
				if len(message) == 0 {
					message = status.State.Terminated.Reason
				}
				state = fmt.Sprintf("Terminated: %s", message)
			} else if status.State.Waiting != nil {
				message := status.State.Waiting.Message
				if len(message) == 0 {
					message = status.State.Waiting.Reason
				}
				state = fmt.Sprintf("Waiting: %s", message)
			}
			props = append(props, &api.TextProperty{Name: "State", Value: state})

			props = append(props, &api.TextProperty{Name: "Image", Value: container.Image})

			if len(container.Command) > 0 {
				props = append(props, &api.TextProperty{Name: "Command", Value: strings.Join(container.Command, " ")})
			}

			prop := &api.GroupProperty{Name: "Env"}
			for i, env := range container.Env {
				id := fmt.Sprintf("env.%d", i)
				if from := env.ValueFrom; from != nil {
					if ref := from.ConfigMapKeyRef; ref != nil {
						var cm corev1.ConfigMap
						if err := e.Get(context.TODO(), types.NamespacedName{Name: ref.Name, Namespace: object.Namespace}, &cm); err != nil {
							prop.Children = append(prop.Children, &api.TextProperty{ID: id, Name: env.Name, Value: fmt.Sprintf("error: %v", err)})
						} else {
							prop.Children = append(prop.Children, &api.TextProperty{ID: id, Name: env.Name, Value: cm.Data[ref.Key]})
						}
					} else if ref := from.SecretKeyRef; ref != nil {
						var secret corev1.Secret
						if err := e.Get(context.TODO(), types.NamespacedName{Name: ref.Name, Namespace: object.Namespace}, &secret); err != nil {
							prop.Children = append(prop.Children, &api.TextProperty{ID: id, Name: env.Name, Value: fmt.Sprintf("error: %v", err)})
						} else {
							prop.Children = append(prop.Children, &api.TextProperty{ID: id, Name: env.Name, Value: string(secret.Data[ref.Key])})
						}
					}
					// TODO field refs
				} else {
					prop.Children = append(prop.Children, &api.TextProperty{ID: id, Name: env.Name, Value: env.Value})
				}
			}
			props = append(props, prop)

			if metrics != nil {
				if cpu := metrics.Usage.Cpu(); cpu != nil {
					c, _ := cpu.AsInt64()
					cpu = resource.NewQuantity(c, resource.DecimalSI)
					cpu.RoundUp(resource.Milli)
					props = append(props, &api.TextProperty{Name: "CPU", Value: fmt.Sprintf("%v", cpu)})
				}
				if mem := metrics.Usage.Memory(); mem != nil {
					m, _ := mem.AsInt64()
					mem = resource.NewQuantity(m, resource.DecimalSI)
					mem.RoundUp(resource.Mega)
					props = append(props, &api.TextProperty{Name: "Memory", Value: fmt.Sprintf("%v", mem)})
				}
			}

			containers = append(containers, &api.GroupProperty{ID: fmt.Sprintf("containers.%d", i), Name: container.Name, Children: props})
		}

		props = append(props, &api.GroupProperty{Name: "Containers", Children: containers})
	case *corev1.ConfigMap:
		var data []api.Property
		for key, value := range object.Data {
			data = append(data, &api.TextProperty{Name: key, Value: value})
		}
		props = append(props, &api.GroupProperty{Name: "Data", Children: data})
	case *corev1.Secret:
		var data []api.Property
		for key, value := range object.Data {
			data = append(data, &api.TextProperty{Name: key, Value: string(value)})
		}
		props = append(props, &api.GroupProperty{Name: "Data", Children: data})
	case *corev1.Service:
		var ports []api.Property
		for _, p := range object.Spec.Ports {
			ports = append(ports, &api.TextProperty{Name: p.Name, Value: strconv.Itoa(int(p.Port))})
		}
		props = append(props, &api.GroupProperty{Name: "Service", Children: []api.Property{
			&api.TextProperty{Name: "Cluster IP", Value: object.Spec.ClusterIP},
			&api.GroupProperty{Name: "Ports", Children: ports},
		}})
	case *corev1.PersistentVolumeClaim:
		var accessModes []string
		for _, m := range object.Spec.AccessModes {
			accessModes = append(accessModes, string(m))
		}
		var storageClass string
		if object.Spec.StorageClassName != nil {
			storageClass = *object.Spec.StorageClassName
		}
		props = append(props, &api.GroupProperty{Name: "Persistent Volume Claim", Children: []api.Property{
			&api.TextProperty{Name: "Class", Value: storageClass},
			&api.TextProperty{Name: "Request", Value: object.Spec.Resources.Requests.Storage().String()},
			&api.TextProperty{Name: "Access modes", Value: strings.Join(accessModes, ", ")},
		}})
	case *corev1.PersistentVolume:
		var accessModes []string
		for _, m := range object.Spec.AccessModes {
			accessModes = append(accessModes, string(m))
		}
		props = append(props, &api.GroupProperty{Name: "Persistent Volume", Children: []api.Property{
			&api.TextProperty{Name: "Class", Value: object.Spec.StorageClassName},
			&api.TextProperty{Name: "Capacity", Value: object.Spec.Capacity.Storage().String()},
			&api.TextProperty{Name: "Access modes", Value: strings.Join(accessModes, ", ")},
		}})
	case *corev1.Node:
		prop := &api.GroupProperty{Name: "Pods"}
		var pods v1.PodList
		e.List(context.TODO(), &pods, client.MatchingFieldsSelector{Selector: fields.OneTermEqualSelector("spec.nodeName", object.Name)})
		for i, pod := range pods.Items {
			prop.Children = append(prop.Children, &api.TextProperty{ID: fmt.Sprintf("pods.%d", i), Source: &pod, Value: pod.Name})
		}
		props = append(props, prop)
	}

	return props
}
