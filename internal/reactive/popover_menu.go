package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type PopoverMenu struct {
	Widget
	Model gio.MenuModeller `gtk:"menu-model"`
}

func (m *PopoverMenu) Type() reflect.Type {
	return reflect.TypeFor[*gtk.PopoverMenu]()
}

func (m *PopoverMenu) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewPopoverMenuFromModel(nil)
	m.Update(ctx, w)
	return w
}

func (m *PopoverMenu) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))
}
