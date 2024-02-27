package extension

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/util"
	"github.com/getseabird/seabird/widget"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (e *Core) CreateColumns(ctx context.Context, res *metav1.APIResource, columns []api.Column) []api.Column {
	switch util.ResourceGVR(res).String() {
	case corev1.SchemeGroupVersion.WithResource("pods").String():
		columns = append(columns,
			api.Column{
				Name:     "Status",
				Priority: 70,
				Bind: func(listitem *gtk.ListItem, object client.Object) {
					pod := object.(*corev1.Pod)
					for _, cond := range pod.Status.Conditions {
						if cond.Type == corev1.ContainersReady {
							listitem.SetChild(widget.NewStatusIcon(cond.Status == corev1.ConditionTrue || cond.Reason == "PodCompleted"))
						}
					}
				},
			},
			// api.Column{
			// 	Name:     "Restarts",
			// 	Priority: 60,
			// 	Bind: func(listitem *gtk.ListItem, object client.Object) {
			// 		pod := object.(*corev1.Pod)
			// 		var restartCount int
			// 		for _, container := range pod.Status.ContainerStatuses {
			// 			restartCount += int(container.RestartCount)
			// 		}
			// 		label := gtk.NewLabel(strconv.Itoa(restartCount))
			// 		label.SetHAlign(gtk.AlignStart)
			// 		listitem.SetChild(label)
			// 	},
			// },
		)

		if e.Metrics != nil {
			columns = append(columns,
				api.Column{
					Name:     "Memory",
					Priority: 50,
					Bind: func(listitem *gtk.ListItem, object client.Object) {
						pod := object.(*corev1.Pod)
						req := resource.NewQuantity(0, resource.DecimalSI)
						for _, container := range pod.Spec.Containers {
							if mem := container.Resources.Requests.Memory(); mem != nil {
								req.Add(*mem)
							}
						}
						use, _ := e.Metrics.PodSum(types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace})
						req.RoundUp(resource.Mega)
						if use != nil {
							use.RoundUp(resource.Mega)
						}
						bar := widget.NewResourceBar(use, req, "")
						bar.SetHAlign(gtk.AlignStart)
						listitem.SetChild(bar)
					},
				},
				api.Column{
					Name:     "CPU",
					Priority: 40,
					Bind: func(listitem *gtk.ListItem, object client.Object) {
						pod := object.(*corev1.Pod)
						req := resource.NewQuantity(0, resource.DecimalSI)
						for _, container := range pod.Spec.Containers {
							if cpu := container.Resources.Requests.Cpu(); cpu != nil {
								req.Add(*cpu)
							}
						}
						_, use := e.Metrics.PodSum(types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace})
						req.RoundUp(resource.Milli)
						if use != nil {
							use.RoundUp(resource.Milli)
						}
						bar := widget.NewResourceBar(use, req, "")
						bar.SetHAlign(gtk.AlignStart)
						listitem.SetChild(bar)
					},
				})
		}
	case corev1.SchemeGroupVersion.WithResource("persistentvolumeclaims").String():
		columns = append(columns, api.Column{
			Name:     "Size",
			Priority: 70,
			Bind: func(listitem *gtk.ListItem, object client.Object) {
				pvc := object.(*corev1.PersistentVolumeClaim)
				label := gtk.NewLabel(pvc.Spec.Resources.Requests.Storage().String())
				label.SetHAlign(gtk.AlignStart)
				listitem.SetChild(label)
			},
		})
	case corev1.SchemeGroupVersion.WithResource("persistentvolumes").String():
		columns = append(columns, api.Column{
			Name:     "Size",
			Priority: 70,
			Bind: func(listitem *gtk.ListItem, object client.Object) {
				pvc := object.(*corev1.PersistentVolume)
				label := gtk.NewLabel(pvc.Spec.Capacity.Storage().String())
				label.SetHAlign(gtk.AlignStart)
				listitem.SetChild(label)
			},
		})
	case corev1.SchemeGroupVersion.WithResource("nodes").String():
		columns = append(columns, api.Column{
			Name:     "Status",
			Priority: 70,
			Bind: func(listitem *gtk.ListItem, object client.Object) {
				node := object.(*corev1.Node)
				for _, cond := range node.Status.Conditions {
					if cond.Type == corev1.NodeReady {
						listitem.SetChild(widget.NewStatusIcon(cond.Status == corev1.ConditionTrue))
					}
				}
			},
		})
	}
	return columns
}

func (e *Core) CreateObjectProperties(ctx context.Context, object client.Object, props []api.Property) []api.Property {
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
						if err := e.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: object.Namespace}, &cm); err != nil {
							prop.Children = append(prop.Children, &api.TextProperty{ID: id, Name: env.Name, Value: fmt.Sprintf("error: %v", err)})
						} else {
							prop.Children = append(prop.Children, &api.TextProperty{ID: id, Name: env.Name, Value: cm.Data[ref.Key]})
						}
					} else if ref := from.SecretKeyRef; ref != nil {
						var secret corev1.Secret
						if err := e.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: object.Namespace}, &secret); err != nil {
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

			var cpu *resource.Quantity
			var mem *resource.Quantity
			if metrics != nil {
				if cpu = metrics.Usage.Cpu(); cpu != nil {
					cpu.RoundUp(resource.Milli)
				}
				props = append(props, &api.TextProperty{Name: "CPU", Value: fmt.Sprintf("%v", cpu)})

				if mem = metrics.Usage.Memory(); mem != nil {
					mem.RoundUp(resource.Mega)
					props = append(props, &api.TextProperty{
						Name:  "Memory",
						Value: fmt.Sprintf("%v", mem),
					})
				}
			}

			containers = append(containers, &api.GroupProperty{
				ID:   fmt.Sprintf("containers.%d", i),
				Name: container.Name, Children: props,
				Widget: func(w gtk.Widgetter, nav *adw.NavigationView) {
					switch row := w.(type) {
					case *adw.ExpanderRow:
						row.AddPrefix(widget.NewStatusIcon(status.Ready))
						if cpu != nil {
							req := container.Resources.Requests.Cpu()
							if req == nil || req.IsZero() {
								req = container.Resources.Limits.Cpu()
							}
							row.AddSuffix(widget.NewResourceBar(cpu, req, "cpu-symbolic"))
						}
						if mem != nil {
							req := container.Resources.Requests.Memory()
							if req == nil || req.IsZero() {
								req = container.Resources.Limits.Memory()
							}
							row.AddSuffix(widget.NewResourceBar(mem, req, "memory-stick-symbolic"))
						}

						logs := adw.NewActionRow()
						logs.SetActivatable(true)
						logs.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
						logs.SetTitle("Logs")
						logs.ConnectActivated(func() {
							nav.Push(widget.NewLogPage(ctx, e.Cluster, object, container.Name).NavigationPage)
						})
						row.AddRow(logs)

						if runtime.GOOS != "windows" {
							exec := adw.NewActionRow()
							exec.SetActivatable(true)
							exec.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
							exec.SetTitle("Exec")
							exec.ConnectActivated(func() {
								nav.Push(widget.NewTerminalPage(ctx, e.Cluster, object, container.Name).NavigationPage)
							})
							row.AddRow(exec)
						}
					}
				},
			})
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
		infoProp := &api.GroupProperty{Name: "Info"}
		infoProp.Children = append(infoProp.Children,
			&api.TextProperty{
				Name:  "Architecture",
				Value: object.Status.NodeInfo.Architecture,
			},
			&api.TextProperty{
				Name:  "Container runtime",
				Value: object.Status.NodeInfo.ContainerRuntimeVersion,
			},
			&api.TextProperty{
				Name:  "Kernel",
				Value: object.Status.NodeInfo.KernelVersion,
			},
			&api.TextProperty{
				Name:  "Kubelet",
				Value: object.Status.NodeInfo.KubeletVersion,
			},
			&api.TextProperty{
				Name:  "Operating system image",
				Value: object.Status.NodeInfo.OSImage,
			},
		)
		props = append(props, infoProp)

		podsProp := &api.GroupProperty{Name: "Pods"}
		var pods corev1.PodList
		e.List(ctx, &pods, client.MatchingFieldsSelector{Selector: fields.OneTermEqualSelector("spec.nodeName", object.Name)})
		for i, pod := range pods.Items {
			podsProp.Children = append(podsProp.Children, &api.TextProperty{
				ID:     fmt.Sprintf("pods.%d", i),
				Source: &pod,
				Value:  pod.Name,
				Widget: func(w gtk.Widgetter, nv *adw.NavigationView) {
					podWidget(pod, w, nv)
				},
			})
		}
		props = append(props, podsProp)
	}

	return props
}
