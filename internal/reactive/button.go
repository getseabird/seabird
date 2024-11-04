package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Button struct {
	Widget
	Label    string                   `gtk:"label"`
	IconName string                   `gtk:"icon-name"`
	Clicked  func(button *gtk.Button) `gtk:"clicked,signal"`
}

func (m *Button) Type() reflect.Type {
	return reflect.TypeFor[*gtk.Button]()
}

func (model *Button) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewButton()
	model.Update(ctx, w)
	return w
}

func (model *Button) Update(ctx context.Context, w gtk.Widgetter) {
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))
}
