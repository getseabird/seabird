package behavior

import (
	"context"
	"log"

	"github.com/getseabird/seabird/util"
	"github.com/imkira/go-observer/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ListBehavior struct {
	*ClusterBehavior

	Objects observer.Property[[]client.Object]
}

func (b *ClusterBehavior) NewListBehavior() *ListBehavior {
	listView := ListBehavior{
		ClusterBehavior: b,
		Objects:         observer.NewProperty[[]client.Object](nil),
	}

	onChange(listView.SelectedResource, listView.onSelectedResourceChange)

	return &listView
}

// We want typed objects for known resources so we can type switch them
func (b *ListBehavior) onSelectedResourceChange(resource *metav1.APIResource) {
	gvr := util.ResourceGVR(resource)
	var res []client.Object
	var list client.ObjectList
	switch gvr.String() {
	case corev1.SchemeGroupVersion.WithResource("pods").String():
		list = &corev1.PodList{}
	case corev1.SchemeGroupVersion.WithResource("configmaps").String():
		list = &corev1.ConfigMapList{}
	case corev1.SchemeGroupVersion.WithResource("secrets").String():
		list = &corev1.SecretList{}
	case appsv1.SchemeGroupVersion.WithResource("deployments").String():
		list = &appsv1.DeploymentList{}
	case appsv1.SchemeGroupVersion.WithResource("statefulsets").String():
		list = &appsv1.StatefulSetList{}
	}
	if list != nil {
		if err := b.client.List(context.TODO(), list); err != nil {
			log.Printf(err.Error())
			return
		}
		switch list := list.(type) {
		case *corev1.PodList:
			for _, i := range list.Items {
				ii := i
				res = append(res, &ii)
			}
		case *corev1.ConfigMapList:
			for _, i := range list.Items {
				ii := i
				res = append(res, &ii)
			}
		case *corev1.SecretList:
			for _, i := range list.Items {
				ii := i
				res = append(res, &ii)
			}
		case *appsv1.DeploymentList:
			for _, i := range list.Items {
				ii := i
				res = append(res, &ii)
			}
		case *appsv1.StatefulSetList:
			for _, i := range list.Items {
				ii := i
				res = append(res, &ii)
			}
		}

		b.Objects.Update(res)
	} else {
		list, err := b.dynamic.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf(err.Error())
			return
		}
		for _, i := range list.Items {
			ii := i
			res = append(res, &ii)
		}
		b.Objects.Update(res)
	}
}
