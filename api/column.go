package api

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Column struct {
	Name     string
	Priority int8
	Bind     func(cell *gtk.ColumnViewCell, object client.Object)
	Compare  func(a, b client.Object) int
}
