package extension

import (
	"context"

	"github.com/getseabird/seabird/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Noop struct{}

func (e *Noop) CreateColumns(ctx context.Context, resource *metav1.APIResource, columns []api.Column) []api.Column {
	return columns
}

func (e *Noop) CreateObjectProperties(ctx context.Context, _ *metav1.APIResource, object client.Object, props []api.Property) []api.Property {
	return props
}
