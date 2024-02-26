package behavior

import (
	"context"

	"github.com/getseabird/seabird/internal/util"
	"github.com/imkira/go-observer/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ListBehavior struct {
	*ClusterBehavior
	ctx         context.Context
	watchCancel context.CancelFunc
	Objects     observer.Property[[]client.Object]
}

func (b *ClusterBehavior) NewListBehavior(ctx context.Context) *ListBehavior {
	listView := ListBehavior{
		ClusterBehavior: b,
		ctx:             ctx,
		Objects:         observer.NewProperty[[]client.Object](nil),
	}

	onChange(ctx, listView.SelectedResource, listView.onSelectedResourceChange)

	return &listView
}

func (b *ListBehavior) onSelectedResourceChange(resource *metav1.APIResource) {
	if b.watchCancel != nil {
		b.watchCancel()
	}
	var ctx context.Context
	ctx, b.watchCancel = context.WithCancel(b.ctx)
	util.ObjectWatcher(ctx, b.Cluster, util.ResourceGVR(resource), b.Objects)
}
