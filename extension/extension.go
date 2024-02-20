package extension

import (
	"github.com/getseabird/seabird/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var Extensions []Constructor

type Constructor func(*api.Cluster) Extension

type Extension interface {
	CreateObjectProperties(client.Object, []api.Property) []api.Property
}
