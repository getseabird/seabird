package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type ListBoxRow struct {
	Widget
	Activatable bool  `gtk:"activatable"`
	Child       Model `gtk:"child"`
}

func (m *ListBoxRow) Type() reflect.Type {
	return reflect.TypeFor[*gtk.ListBoxRow]()
}

func (model *ListBoxRow) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewListBoxRow()
	model.Update(ctx, w)
	return w
}

func (model *ListBoxRow) Update(ctx context.Context, w gtk.Widgetter) {
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))
}
