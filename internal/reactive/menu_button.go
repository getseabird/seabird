package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type MenuButton struct {
	Widget
	Label    string `gtk:"label"`
	IconName string `gtk:"icon-name"`
	Popover  Model  `gtk:"popover"`
}

func (m *MenuButton) Type() reflect.Type {
	return reflect.TypeFor[*gtk.MenuButton]()
}

func (m *MenuButton) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewMenuButton()
	m.Update(ctx, w)
	return w
}

func (m *MenuButton) Update(ctx context.Context, wi gtk.Widgetter) {
	w := wi.(*gtk.MenuButton)
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(&w.Widget))
}
