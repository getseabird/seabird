package api

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Column struct {
	Name     string
	Priority int32
	Bind     func(cell Cell, object client.Object)
	Compare  func(a, b client.Object) int
}

type Cell struct {
	*gtk.ColumnViewCell
}

func (cell *Cell) SetLabel(format string, a ...any) {
	label := gtk.NewLabel(fmt.Sprintf(format, a...))
	label.SetHAlign(gtk.AlignStart)
	label.SetEllipsize(pango.EllipsizeEnd)
	cell.SetChild(label)
}
