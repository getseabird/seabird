package extension

import (
	"context"

	"github.com/getseabird/seabird/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var Extensions []Constructor

type Constructor func(*api.Cluster) Extension

type Extension interface {
	CreateColumns(ctx context.Context, resource *metav1.APIResource, columns []api.Column) []api.Column
	CreateObjectProperties(ctx context.Context, object client.Object, props []api.Property) []api.Property
}
