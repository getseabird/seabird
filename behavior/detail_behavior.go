package behavior

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/getseabird/seabird/util"
	"github.com/imkira/go-observer/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DetailBehavior struct {
	*ClusterBehavior

	SelectedObject observer.Property[client.Object]
	Yaml           observer.Property[string]
	Properties     observer.Property[[]ObjectProperty]
}

func (b *ClusterBehavior) NewRootDetailBehavior() *DetailBehavior {
	db := b.NewDetailBehavior()
	b.RootDetailBehavior = db
	return db
}

func (b *ClusterBehavior) NewDetailBehavior() *DetailBehavior {
	d := DetailBehavior{
		ClusterBehavior: b,
		SelectedObject:  observer.NewProperty[client.Object](nil),
		Yaml:            observer.NewProperty[string](""),
		Properties:      observer.NewProperty[[]ObjectProperty](nil),
	}

	onChange(d.SelectedObject, d.onObjectChange)

	return &d
}

func (b *DetailBehavior) onObjectChange(object client.Object) {
	if object == nil {
		b.Properties.Update([]ObjectProperty{})
		b.Yaml.Update("")
		return
	}

	codec := unstructured.NewJSONFallbackEncoder(serializer.NewCodecFactory(b.scheme).LegacyCodec(b.scheme.PreferredVersionAllGroups()...))
	encoded, err := runtime.Encode(codec, object)
	if err != nil {
		b.Yaml.Update(fmt.Sprintf("error: %v", err))
	} else {
		yaml, err := util.JsonToYaml(encoded)
		if err != nil {
			b.Yaml.Update(fmt.Sprintf("error: %v", err))
		} else {
			b.Yaml.Update(string(yaml))
		}
	}

	var labels []ObjectProperty
	for key, value := range object.GetLabels() {
		labels = append(labels, ObjectProperty{Name: key, Value: value})
	}
	var annotations []ObjectProperty
	for key, value := range object.GetAnnotations() {
		annotations = append(annotations, ObjectProperty{Name: key, Value: value})
	}
	var owners []ObjectProperty
	for _, ref := range object.GetOwnerReferences() {
		owners = append(owners, ObjectProperty{Name: fmt.Sprintf("%s %s", ref.APIVersion, ref.Kind), Value: ref.Name})
	}

	var properties []ObjectProperty
	properties = append(properties,
		ObjectProperty{
			Name: "Metadata",
			Children: []ObjectProperty{
				ObjectProperty{
					Name:  "Name",
					Value: object.GetName(),
				},
				ObjectProperty{
					Name:  "Namespace",
					Value: object.GetNamespace(),
				},
				ObjectProperty{
					Name:     "Labels",
					Children: labels,
				},
				ObjectProperty{
					Name:     "Annotations",
					Children: annotations,
				},
				ObjectProperty{
					Name:     "Owners",
					Children: owners,
				},
			},
		})

	switch object := object.(type) {
	case *corev1.Pod:
		var containers []ObjectProperty

		var podMetrics *metricsv1beta1.PodMetrics
		if b.metrics != nil {
			podMetrics = b.metrics.PodValue(types.NamespacedName{Name: object.Name, Namespace: object.Namespace})
		}

		for _, c := range object.Spec.Containers {
			container := c
			var props []ObjectProperty
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
			props = append(props, ObjectProperty{Name: "State", Value: state})

			props = append(props, ObjectProperty{Name: "Image", Value: container.Image})

			if len(container.Command) > 0 {
				props = append(props, ObjectProperty{Name: "Command", Value: strings.Join(container.Command, " ")})
			}

			prop := ObjectProperty{Name: "Env"}
			for _, env := range container.Env {
				if from := env.ValueFrom; from != nil {
					if ref := from.ConfigMapKeyRef; ref != nil {
						var cm corev1.ConfigMap
						if err := b.client.Get(context.TODO(), types.NamespacedName{Name: ref.Name, Namespace: object.Namespace}, &cm); err != nil {
							prop.Children = append(prop.Children, ObjectProperty{Name: env.Name, Value: fmt.Sprintf("error: %v", err)})
						} else {
							prop.Children = append(prop.Children, ObjectProperty{Name: env.Name, Value: cm.Data[ref.Key]})
						}
					} else if ref := from.SecretKeyRef; ref != nil {
						var secret corev1.Secret
						if err := b.client.Get(context.TODO(), types.NamespacedName{Name: ref.Name, Namespace: object.Namespace}, &secret); err != nil {
							prop.Children = append(prop.Children, ObjectProperty{Name: env.Name, Value: fmt.Sprintf("error: %v", err)})
						} else {
							prop.Children = append(prop.Children, ObjectProperty{Name: env.Name, Value: string(secret.Data[ref.Key])})
						}
					}
					// TODO field refs
				} else {
					prop.Children = append(prop.Children, ObjectProperty{Name: env.Name, Value: env.Value})
				}
			}
			props = append(props, prop)

			if metrics != nil {
				if cpu := metrics.Usage.Cpu(); cpu != nil {
					c, _ := cpu.AsInt64()
					cpu = resource.NewQuantity(c, resource.DecimalSI)
					cpu.RoundUp(resource.Milli)
					props = append(props, ObjectProperty{Name: "CPU", Value: fmt.Sprintf("%v", cpu)})
				}
				if mem := metrics.Usage.Memory(); mem != nil {
					m, _ := mem.AsInt64()
					mem = resource.NewQuantity(m, resource.DecimalSI)
					mem.RoundUp(resource.Mega)
					props = append(props, ObjectProperty{Name: "Memory", Value: fmt.Sprintf("%v", mem)})
				}
			}

			containers = append(containers, ObjectProperty{Name: container.Name, Children: props, Object: &container})
		}

		properties = append(properties, ObjectProperty{Name: "Containers", Children: containers})
	case *corev1.ConfigMap:
		var data []ObjectProperty
		for key, value := range object.Data {
			data = append(data, ObjectProperty{Name: key, Value: value})
		}
		properties = append(properties, ObjectProperty{Name: "Data", Children: data})
	case *corev1.Secret:
		var data []ObjectProperty
		for key, value := range object.Data {
			data = append(data, ObjectProperty{Name: key, Value: string(value)})
		}
		properties = append(properties, ObjectProperty{Name: "Data", Children: data})
	case *corev1.Service:
		var ports []ObjectProperty
		for _, p := range object.Spec.Ports {
			ports = append(ports, ObjectProperty{Name: p.Name, Value: strconv.Itoa(int(p.Port))})
		}
		properties = append(properties, ObjectProperty{Name: "Service", Children: []ObjectProperty{
			{Name: "Cluster IP", Value: object.Spec.ClusterIP},
			{Name: "Ports", Children: ports},
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
		properties = append(properties, ObjectProperty{Name: "Persistent Volume Claim", Children: []ObjectProperty{
			{Name: "Class", Value: storageClass},
			{Name: "Request", Value: object.Spec.Resources.Requests.Storage().String()},
			{Name: "Access modes", Value: strings.Join(accessModes, ", ")},
		}})
	case *corev1.PersistentVolume:
		var accessModes []string
		for _, m := range object.Spec.AccessModes {
			accessModes = append(accessModes, string(m))
		}
		properties = append(properties, ObjectProperty{Name: "Persistent Volume", Children: []ObjectProperty{
			{Name: "Class", Value: object.Spec.StorageClassName},
			{Name: "Capacity", Value: object.Spec.Capacity.Storage().String()},
			{Name: "Access modes", Value: strings.Join(accessModes, ", ")},
		}})
	case *corev1.Node:
		prop := ObjectProperty{Name: "Pods"}
		var pods v1.PodList
		b.client.List(context.TODO(), &pods, client.MatchingFieldsSelector{Selector: fields.OneTermEqualSelector("spec.nodeName", object.Name)})
		for _, p := range pods.Items {
			pod := p
			prop.Children = append(prop.Children, ObjectProperty{Value: pod.Name, Object: &pod})
		}
		properties = append(properties, prop)
	case *appsv1.Deployment:
		prop := ObjectProperty{Name: "Pods"}
		var pods v1.PodList
		b.client.List(context.TODO(), &pods, client.InNamespace(object.Namespace), client.MatchingLabels(object.Spec.Selector.MatchLabels))
		// TODO should we also filter pods by owner? takes one more api call to fetch replicasets
		for _, p := range pods.Items {
			pod := p
			prop.Children = append(prop.Children, ObjectProperty{Value: pod.Name, Object: &pod})
		}
		properties = append(properties, prop)
	case *appsv1.StatefulSet:
		prop := ObjectProperty{Name: "Pods"}
		var pods v1.PodList
		b.client.List(context.TODO(), &pods, client.InNamespace(object.Namespace), client.MatchingLabels(object.Spec.Selector.MatchLabels))
		for _, p := range pods.Items {
			pod := p
			var ok bool
			for _, owner := range pod.OwnerReferences {
				if owner.UID == object.UID {
					ok = true
				}
			}
			if !ok {
				continue
			}
			prop.Children = append(prop.Children, ObjectProperty{Value: pod.Name, Object: &pod})
		}
		properties = append(properties, prop)
	}

	b.Properties.Update(properties)

}

type ObjectProperty struct {
	Object   any
	Name     string
	Value    string
	Children []ObjectProperty
}

func (b *DetailBehavior) PodLogs() ([]byte, error) {
	pod := b.SelectedObject.Value().(*corev1.Pod)
	req := b.clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
	r, err := req.Stream(context.TODO())
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}
