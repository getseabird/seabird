package behavior

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/getseabird/seabird/util"
	"github.com/imkira/go-observer/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DetailBehavior struct {
	*ClusterBehavior

	Yaml       observer.Property[string]
	Properties observer.Property[[]ObjectProperty]
}

func (b *ClusterBehavior) NewDetailBehavior() *DetailBehavior {
	d := DetailBehavior{
		ClusterBehavior: b,
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

		for _, container := range object.Spec.Containers {
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
					message = status.State.Waiting.Reason
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
					props = append(props, ObjectProperty{Name: "CPU", Value: fmt.Sprintf("%v%%", math.Round(cpu.AsApproximateFloat64()*10000)/10000)})
				}
				if mem := metrics.Usage.Memory(); mem != nil {
					bytes, _ := mem.AsInt64()
					props = append(props, ObjectProperty{Name: "Memory", Value: humanize.Bytes(uint64(bytes))})
				}
			}

			containers = append(containers, ObjectProperty{Name: container.Name, Children: props})
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
	}

	b.Properties.Update(properties)

}

type ObjectProperty struct {
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
