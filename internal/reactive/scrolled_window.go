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

func (model *ScrolledWindow) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewScrolledWindow()
	model.Update(ctx, w)
	return w
}

func (model *ScrolledWindow) Update(ctx context.Context, w gtk.Widgetter) {
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))
}
