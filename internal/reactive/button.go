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

func (m *Button) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewButton()
	m.Update(ctx, w)
	return w
}

func (m *Button) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))
}
