package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AdwBin struct {
	Widget
	Child Model `gtk:"child"`
}

func (m *AdwBin) Type() reflect.Type {
	return reflect.TypeFor[*adw.Bin]()
}

func (m *AdwBin) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewBin()
	m.Update(ctx, w)
	return w
}

func (m *AdwBin) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))
}
