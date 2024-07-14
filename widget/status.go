package widget

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatusType string

const StatusTypeInfo StatusType = "accent"
const StatusTypeSuccess StatusType = "success"
const StatusTypeWarning StatusType = "warning"
const StatusTypeError StatusType = "error"

type Status struct {
	Condition string
	Reason    string
	Type      StatusType
	Children  []*Status
}

func NewStatus(cond string, reason string, typ StatusType) *Status {
	return &Status{Condition: cond, Reason: reason, Type: typ}
}

func ObjectStatus(object client.Object) *Status {
	switch object := object.(type) {
	case *corev1.Pod:
		var children []*Status
		for _, cs := range object.Status.ContainerStatuses {
			if cs.State.Running != nil {
				children = append(children, &Status{
					Condition: "Running",
					Type:      StatusTypeSuccess,
				})
			} else if cs.State.Terminated != nil && cs.State.Terminated.Reason == "Completed" {
				children = append(children, &Status{
					Condition: "Terminated",
					Reason:    cs.State.Terminated.Reason,
					Type:      StatusTypeInfo,
				})
			} else {
				children = append(children, &Status{
					Type: StatusTypeWarning,
				})
			}
		}
		for _, cond := range object.Status.Conditions {
			if cond.Type == corev1.ContainersReady {
				if cond.Status == corev1.ConditionTrue {
					return &Status{
						Condition: string(corev1.ContainersReady),
						Reason:    cond.Reason,
						Type:      StatusTypeSuccess,
						Children:  children,
					}
				} else {
					if cond.Reason == "PodCompleted" {
						return &Status{
							Condition: string(corev1.ContainersReady),
							Reason:    cond.Reason,
							Type:      StatusTypeInfo,
							Children:  children,
						}
					}
					return &Status{
						Condition: string(corev1.ContainersReady),
						Reason:    cond.Reason,
						Type:      StatusTypeWarning,
						Children:  children,
					}
				}
			}
		}
	case *corev1.Node:
		for _, cond := range object.Status.Conditions {
			if cond.Type == corev1.NodeReady {
				if cond.Status == corev1.ConditionTrue {
					return &Status{
						Condition: string(corev1.NodeReady),
						Reason:    cond.Reason,
						Type:      StatusTypeSuccess,
					}
				} else {
					return &Status{
						Condition: string(corev1.NodeReady),
						Reason:    cond.Reason,
						Type:      StatusTypeWarning,
					}
				}
			}
		}
	case *appsv1.Deployment:
		for _, cond := range object.Status.Conditions {
			if cond.Type == appsv1.DeploymentAvailable {
				if cond.Status == corev1.ConditionTrue {
					return &Status{
						Condition: string(appsv1.DeploymentAvailable),
						Reason:    cond.Reason,
						Type:      StatusTypeSuccess,
					}
				} else {
					return &Status{
						Condition: string(appsv1.DeploymentAvailable),
						Reason:    cond.Reason,
						Type:      StatusTypeWarning,
					}
				}
			}
		}
	case *appsv1.ReplicaSet:
		if object.Status.ReadyReplicas == object.Status.Replicas {
			return &Status{
				Type: StatusTypeSuccess,
			}
		} else {
			return &Status{
				Type: StatusTypeWarning,
			}
		}
	case *appsv1.StatefulSet:
		if object.Status.ReadyReplicas == object.Status.Replicas {
			return &Status{
				Type: StatusTypeSuccess,
			}
		} else {
			return &Status{
				Type: StatusTypeWarning,
			}
		}
	case *corev1.PersistentVolumeClaim:
		if object.Status.Phase == corev1.ClaimBound {
			return &Status{
				Type: StatusTypeSuccess,
			}
		} else {
			return &Status{
				Type: StatusTypeWarning,
			}
		}
	}
	return &Status{
		Type: StatusTypeError,
	}
}

// func (status *Status) Label() *gtk.Label {
// 	label := gtk.NewLabel(fmt.Sprintf("%v", status.Condition))
// 	label.SetHAlign(gtk.AlignStart)
// 	label.AddCSSClass(string(status.Type))
// 	return label
// }

func (status *Status) Icon() *gtk.Image {
	switch status.Type {
	case StatusTypeSuccess, StatusTypeInfo:
		icon := gtk.NewImageFromIconName("emblem-ok-symbolic")
		icon.AddCSSClass(string(status.Type))
		icon.SetHAlign(gtk.AlignStart)
		return icon
	case StatusTypeWarning:
		icon := gtk.NewImageFromIconName("dialog-warning")
		icon.AddCSSClass(string(status.Type))
		icon.SetHAlign(gtk.AlignStart)
		return icon
	case StatusTypeError:
		icon := gtk.NewImageFromIconName("dialog-error")
		icon.AddCSSClass(string(status.Type))
		icon.SetHAlign(gtk.AlignStart)
		return icon
	default:
		return nil
	}
}

func (status *Status) Icons() []*gtk.Image {
	if len(status.Children) == 0 {
		return []*gtk.Image{status.Icon()}
	}

	var icons []*gtk.Image
	for _, s := range status.Children {
		icons = append(icons, s.Icon())
	}
	return icons
}
