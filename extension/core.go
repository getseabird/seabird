package extension

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/style"
	"github.com/getseabird/seabird/internal/util"
	"github.com/getseabird/seabird/widget"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/reference"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	Extensions = append(Extensions, NewCore)
}

func NewCore(_ context.Context, cluster *api.Cluster) (Extension, error) {
	return &Core{
		Cluster:     cluster,
		portforward: PortForwarder{cluster, map[types.NamespacedName]*portforward.PortForwarder{}},
	}, nil
}

type Core struct {
	Noop
	*api.Cluster
	portforward PortForwarder
}

func (e *Core) CreateColumns(ctx context.Context, res *metav1.APIResource, columns []api.Column) []api.Column {
	switch util.GVRForResource(res).String() {
	case corev1.SchemeGroupVersion.WithResource("pods").String():
		columns = append(columns,
			api.Column{
				Name:     "Status",
				Priority: 70,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					box := gtk.NewBox(gtk.OrientationHorizontal, 4)
					for _, icon := range api.NewStatusWithObject(object).Icons() {
						box.Append(icon)
					}
					cell.SetChild(box)
				},
				Compare: api.CompareObjectStatus,
			},
			api.Column{
				Name:     "Memory",
				Priority: 50,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
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
					cell.SetChild(bar)
				},
			},
			api.Column{
				Name:     "CPU",
				Priority: 40,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
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
					cell.SetChild(bar)
				},
			},
		)
	case corev1.SchemeGroupVersion.WithResource("persistentvolumeclaims").String():
		columns = append(columns,
			api.Column{
				Name:     "Size",
				Priority: 70,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					pvc := object.(*corev1.PersistentVolumeClaim)
					label := gtk.NewLabel(pvc.Spec.Resources.Requests.Storage().String())
					label.SetHAlign(gtk.AlignStart)
					cell.SetChild(label)
				},
			},
			api.Column{
				Name:     "Class",
				Priority: 60,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					pvc := object.(*corev1.PersistentVolumeClaim)
					label := gtk.NewLabel(ptr.Deref(pvc.Spec.StorageClassName, ""))
					label.SetHAlign(gtk.AlignStart)
					cell.SetChild(label)
				},
			},
			api.Column{
				Name:     "Volume",
				Priority: 60,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					pvc := object.(*corev1.PersistentVolumeClaim)
					label := gtk.NewLabel(pvc.Spec.VolumeName)
					label.SetHAlign(gtk.AlignStart)
					label.SetEllipsize(pango.EllipsizeEnd)
					cell.SetChild(label)
				},
			},
		)
	case corev1.SchemeGroupVersion.WithResource("persistentvolumes").String():
		columns = append(columns,
			api.Column{
				Name:     "Size",
				Priority: 70,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					pv := object.(*corev1.PersistentVolume)
					label := gtk.NewLabel(pv.Spec.Capacity.Storage().String())
					label.SetHAlign(gtk.AlignStart)
					cell.SetChild(label)
				},
			},
			api.Column{
				Name:     "Phase",
				Priority: 70,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					pv := object.(*corev1.PersistentVolume)
					label := gtk.NewLabel(string(pv.Status.Phase))
					label.SetHAlign(gtk.AlignStart)
					cell.SetChild(label)
				},
			},
			api.Column{
				Name:     "Claim",
				Priority: 70,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					pv := object.(*corev1.PersistentVolume)
					if pv.Spec.ClaimRef != nil {
						label := gtk.NewLabel(string(pv.Spec.ClaimRef.Name))
						label.SetHAlign(gtk.AlignStart)
						label.SetEllipsize(pango.EllipsizeEnd)
						cell.SetChild(label)
					}
				},
			},
		)

	case corev1.SchemeGroupVersion.WithResource("nodes").String():
		columns = append(columns,
			api.Column{
				Name:     "Status",
				Priority: 70,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					cell.SetChild(api.NewStatusWithObject(object).Icon())
				},
				Compare: api.CompareObjectStatus,
			},
			api.Column{
				Name:     "Pods",
				Priority: 60,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					pod := object.(*corev1.Node)
					var pods corev1.PodList
					e.List(ctx, &pods, client.MatchingFieldsSelector{Selector: fields.OneTermEqualSelector("spec.nodeName", pod.Name)})
					label := gtk.NewLabel(fmt.Sprintf("%d", len(pods.Items)))
					label.SetHAlign(gtk.AlignStart)
					cell.SetChild(label)
				},
			},
			api.Column{
				Name:     "Memory",
				Priority: 50,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					node := object.(*corev1.Node)
					metrics := e.Metrics.Node(node.Name)
					if metrics == nil {
						return
					}
					bar := widget.NewResourceBar(metrics.Usage.Memory(), node.Status.Allocatable.Memory(), "")
					bar.SetHAlign(gtk.AlignStart)
					cell.SetChild(bar)
				},
			},
			api.Column{
				Name:     "CPU",
				Priority: 40,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					node := object.(*corev1.Node)
					metrics := e.Metrics.Node(node.Name)
					if metrics == nil {
						return
					}
					bar := widget.NewResourceBar(metrics.Usage.Cpu(), node.Status.Allocatable.Cpu(), "")
					bar.SetHAlign(gtk.AlignStart)
					cell.SetChild(bar)
				},
			},
		)
	case corev1.SchemeGroupVersion.WithResource("namespaces").String():
		columns = append(columns,
			api.Column{
				Name:     "Phase",
				Priority: 70,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					ns := object.(*corev1.Namespace)
					label := gtk.NewLabel(string(ns.Status.Phase))
					label.SetHAlign(gtk.AlignStart)
					cell.SetChild(label)
				},
			},
		)
	case corev1.SchemeGroupVersion.WithResource("configmaps").String():
		columns = append(columns,
			api.Column{
				Name:     "Keys",
				Priority: 70,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					configmap := object.(*corev1.ConfigMap)
					var keys []string
					for key := range configmap.Data {
						keys = append(keys, key)
					}
					label := gtk.NewLabel(strings.Join(keys, ", "))
					label.SetHAlign(gtk.AlignStart)
					label.SetEllipsize(pango.EllipsizeEnd)
					cell.SetChild(label)
				},
			},
		)
	case corev1.SchemeGroupVersion.WithResource("secrets").String():
		columns = append(columns,
			api.Column{
				Name:     "Keys",
				Priority: 70,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					secret := object.(*corev1.Secret)
					var keys []string
					for key := range secret.Data {
						keys = append(keys, key)
					}
					label := gtk.NewLabel(strings.Join(keys, ", "))
					label.SetHAlign(gtk.AlignStart)
					label.SetEllipsize(pango.EllipsizeEnd)
					cell.SetChild(label)
				},
			},
		)
	case corev1.SchemeGroupVersion.WithResource("services").String():
		columns = append(columns,
			api.Column{
				Name:     "Type",
				Priority: 70,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					service := object.(*corev1.Service)
					label := gtk.NewLabel(string(service.Spec.Type))
					label.SetHAlign(gtk.AlignStart)
					cell.SetChild(label)
				},
			},
			api.Column{
				Name:     "Ports",
				Priority: 60,
				Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
					svc := object.(*corev1.Service)
					var ports []string
					for _, port := range svc.Spec.Ports {
						ports = append(ports, strconv.Itoa(int(port.Port)))
					}
					label := gtk.NewLabel(strings.Join(ports, ", "))
					label.SetHAlign(gtk.AlignStart)
					label.SetEllipsize(pango.EllipsizeEnd)
					cell.SetChild(label)
				},
			},
		)
	}
	return columns
}

func (e *Core) CreateObjectProperties(ctx context.Context, _ *metav1.APIResource, object client.Object, props []api.Property) []api.Property {
	switch object := object.(type) {
	case *corev1.Pod:
		var containers []api.Property

		for i, container := range object.Spec.Containers {
			var props []api.Property
			var status corev1.ContainerStatus
			for _, s := range object.Status.ContainerStatuses {
				if s.Name == container.Name {
					status = s
					break
				}
			}

			podMetrics := e.Metrics.Pod(types.NamespacedName{Name: object.Name, Namespace: object.Namespace})
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

			var restartCount int
			for _, container := range object.Status.ContainerStatuses {
				restartCount += int(container.RestartCount)
			}
			props = append(props, &api.TextProperty{Name: "Restarts", Value: fmt.Sprintf("%d", restartCount)})

			props = append(props, &api.TextProperty{Name: "Image", Value: container.Image})

			if len(container.Command) > 0 {
				props = append(props, &api.TextProperty{Name: "Command", Value: strings.Join(container.Command, " ")})
			}

			envs := &api.GroupProperty{Name: "Env"}
			for i, env := range container.Env {
				id := fmt.Sprintf("env.%d", i)
				if from := env.ValueFrom; from != nil {
					if ref := from.ConfigMapKeyRef; ref != nil {
						var cm corev1.ConfigMap
						if err := e.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: object.Namespace}, &cm); err != nil {
							envs.Children = append(envs.Children, &api.TextProperty{ID: id, Name: env.Name, Value: fmt.Sprintf("error: %v", err)})
						} else {
							envs.Children = append(envs.Children, &api.TextProperty{ID: id, Name: env.Name, Value: cm.Data[ref.Key]})
						}
					} else if ref := from.SecretKeyRef; ref != nil {
						var secret corev1.Secret
						if err := e.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: object.Namespace}, &secret); err != nil {
							envs.Children = append(envs.Children, &api.TextProperty{ID: id, Name: env.Name, Value: fmt.Sprintf("error: %v", err)})
						} else {
							envs.Children = append(envs.Children, &api.TextProperty{ID: id, Name: env.Name, Value: string(secret.Data[ref.Key])})
						}
					}
					// TODO field refs
				} else {
					envs.Children = append(envs.Children, &api.TextProperty{ID: id, Name: env.Name, Value: env.Value})
				}
			}
			props = append(props, envs)

			ports := &api.GroupProperty{Name: "Ports"}
			for _, port := range container.Ports {
				ports.Children = append(ports.Children, &api.TextProperty{
					Name:  port.Name,
					Value: fmt.Sprintf("%d", port.ContainerPort),
					Widget: func(w gtk.Widgetter, nv *adw.NavigationView) {
						name := types.NamespacedName{Name: object.Name, Namespace: object.Namespace}
						switch box := w.(type) {
						case *gtk.Box:
							btn := gtk.NewButton()
							e.portforward.UpdateButton(ctx, btn, name, []string{fmt.Sprintf(":%d", port.ContainerPort)})
							box.Append(btn)
						}
					},
				})
			}
			props = append(props, ports)

			var cpu *resource.Quantity
			var mem *resource.Quantity
			if metrics != nil {
				if cpu = metrics.Usage.Cpu(); cpu != nil {
					cpu.RoundUp(resource.Milli)
					cpu.Format = resource.DecimalSI
					props = append(props, &api.TextProperty{Name: "CPU", Value: fmt.Sprintf("%v", cpu)})
				}

				if mem = metrics.Usage.Memory(); mem != nil {
					mem.RoundUp(resource.Mega)
					mem.Format = resource.DecimalSI
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
						row.AddPrefix(api.NewStatusWithObject(object).Icon())
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

						if !style.Eq(style.Windows) {
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
			&api.TextProperty{
				Name:  "Volume",
				Value: object.Spec.VolumeName,
				Reference: &corev1.ObjectReference{
					Kind:       "PersistentVolume",
					APIVersion: corev1.SchemeGroupVersion.String(),
					Name:       object.Spec.VolumeName,
					Namespace:  object.Namespace,
				},
			},
		}})
	case *corev1.PersistentVolume:
		var accessModes []string
		for _, m := range object.Spec.AccessModes {
			accessModes = append(accessModes, string(m))
		}

		group := &api.GroupProperty{Name: "Persistent Volume", Children: []api.Property{
			&api.TextProperty{Name: "Class", Value: object.Spec.StorageClassName},
			&api.TextProperty{Name: "Capacity", Value: object.Spec.Capacity.Storage().String()},
			&api.TextProperty{Name: "Access modes", Value: strings.Join(accessModes, ", ")},
			&api.TextProperty{Name: "Reclaim policy", Value: string(object.Spec.PersistentVolumeReclaimPolicy)},
			&api.TextProperty{Name: "Phase", Value: string(object.Status.Phase)},
		}}
		if object.Spec.ClaimRef != nil {
			group.Children = append(group.Children, &api.TextProperty{Name: "Claim", Value: object.Spec.ClaimRef.Name, Reference: object.Spec.ClaimRef})
		}
		props = append(props, group)

	case *corev1.Node:
		infoProp := &api.GroupProperty{Name: "Info"}
		mem := object.Status.Allocatable.Memory()
		mem.RoundUp(resource.Mega)
		mem.Format = resource.DecimalSI
		cpu := object.Status.Allocatable.Cpu()
		cpu.RoundUp(resource.Milli)
		cpu.Format = resource.DecimalSI
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
			&api.TextProperty{
				Name:  "Memory",
				Value: fmt.Sprintf("%v", mem),
			},
			&api.TextProperty{
				Name:  "CPU",
				Value: fmt.Sprintf("%v", cpu),
			},
		)
		props = append(props, infoProp)

		podsProp := &api.GroupProperty{Name: "Pods"}
		var pods corev1.PodList
		e.List(ctx, &pods, client.MatchingFieldsSelector{Selector: fields.OneTermEqualSelector("spec.nodeName", object.Name)})
		for i, pod := range pods.Items {
			ref, _ := reference.GetReference(e.Scheme, &pod)
			podsProp.Children = append(podsProp.Children, &api.TextProperty{
				ID:        fmt.Sprintf("pods.%d", i),
				Reference: ref,
				Value:     pod.Name,
				Widget: func(w gtk.Widgetter, nv *adw.NavigationView) {
					switch row := w.(type) {
					case *adw.ActionRow:
						row.AddPrefix(api.NewStatusWithObject(&pod).Icon())
					}
				},
			})
		}
		props = append(props, podsProp)
	}

	return props
}
