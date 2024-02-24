package extension

import (
	"github.com/getseabird/seabird/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var Extensions []Constructor

type Constructor func(*api.Cluster) Extension

type Extension interface {
	CreateColumns(resource *metav1.APIResource, columns []api.Column) []api.Column
	CreateObjectProperties(object client.Object, props []api.Property) []api.Property
}
