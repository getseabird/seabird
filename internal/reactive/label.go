package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Label struct {
	Widget
	Label string `gtk:"label"`
}

func (m *Label) Type() reflect.Type {
	return reflect.TypeFor[*gtk.Label]()
}

func (model *Label) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewLabel(model.Label)
	model.Update(ctx, w)
	return w
}

func (model *Label) Update(ctx context.Context, w gtk.Widgetter) {
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))
}
