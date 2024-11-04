package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/internal/ctxt"
)

type Window struct {
	Widget
	Title         string `gtk:"title"`
	IconName      string `gtk:"icon-name"`
	Child         Model
	DefaultHeight int `gtk:"default-height"`
	DefaultWidth  int `gtk:"default-width"`
}

func (m *Window) Type() reflect.Type {
	return reflect.TypeFor[*gtk.Window]()
}

func (model *Window) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewWindow()
	model.Update(ctx, w)
	return w
}

func (model *Window) Update(ctx context.Context, wi gtk.Widgetter) {
	w := wi.(*gtk.Window)
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))

	if model.Child != nil {
		if child := w.Child(); model.Child.Type() == reflect.TypeOf(child) {
			updateChild(child, model.Child)
		} else {
			w.SetChild(createChild(ctxt.With(ctx, w), model.Child))
		}
	}
}
