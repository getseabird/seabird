package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AdwViewStack struct {
	Ref[*adw.ViewStack]
	Widget
	Pages []AdwViewStackPage
}

func (m *AdwViewStack) Type() reflect.Type {
	return reflect.TypeFor[*adw.ViewStack]()
}

func (m *AdwViewStack) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewViewStack()
	m.Update(ctx, w)
	return w
}

func (m *AdwViewStack) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))
}
