package behavior

import (
	"github.com/getseabird/seabird/internal/util"
	"github.com/imkira/go-observer/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ListBehavior struct {
	*ClusterBehavior
	Objects   observer.Property[[]client.Object]
	stopWatch chan struct{}
}

func (b *ClusterBehavior) NewListBehavior() *ListBehavior {
	listView := ListBehavior{
		ClusterBehavior: b,
		Objects:         observer.NewProperty[[]client.Object](nil),
	}

	onChange(listView.SelectedResource, listView.onSelectedResourceChange)

	return &listView
}

func (b *ListBehavior) onSelectedResourceChange(resource *metav1.APIResource) {
	if b.stopWatch != nil {
		close(b.stopWatch)
	}
	b.stopWatch = make(chan struct{})
	util.ObjectWatcher(b.Cluster, util.ResourceGVR(resource), b.stopWatch, b.Objects)
}
