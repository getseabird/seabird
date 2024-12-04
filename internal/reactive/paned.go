package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Paned struct {
	Widget
	Orientation gtk.Orientation
	StartChild  Model `gtk:"start-child"`
	EndChild    Model `gtk:"end-child"`
}

func (m *Paned) Type() reflect.Type {
	return reflect.TypeFor[*gtk.Paned]()
}

func (m *Paned) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewPaned(m.Orientation)
	m.Update(ctx, w)
	return w
}

func (m *Paned) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))
}
