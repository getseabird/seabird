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

func (m *Window) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewWindow()
	m.Update(ctx, w)
	return w
}

func (m *Window) Update(ctx context.Context, wi gtk.Widgetter) {
	w := wi.(*gtk.Window)
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))

	if m.Child != nil {
		if child := w.Child(); m.Child.Type() == reflect.TypeOf(child) {
			updateChild(child, m.Child)
		} else {
			w.SetChild(createChild(ctxt.With(ctx, w), m.Child))
		}
	}
}
