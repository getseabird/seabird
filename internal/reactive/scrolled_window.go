package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type ScrolledWindow struct {
	Widget
	Child Model `gtk:"child"`
}

func (m *ScrolledWindow) Type() reflect.Type {
	return reflect.TypeFor[*gtk.ScrolledWindow]()
}

func (m *ScrolledWindow) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewScrolledWindow()
	m.Update(ctx, w)
	return w
}

func (m *ScrolledWindow) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))
}
