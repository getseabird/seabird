package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AdwViewSwitcher struct {
	Widget
	ViewStack *adw.ViewStack
	Policy    adw.ViewSwitcherPolicy `gtk:"policy"`
}

func (m *AdwViewSwitcher) Type() reflect.Type {
	return reflect.TypeFor[*adw.ViewSwitcher]()
}

func (m *AdwViewSwitcher) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewViewSwitcher()
	m.Update(ctx, w)
	return w
}

func (m *AdwViewSwitcher) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))

	switcher := w.(*adw.ViewSwitcher)

	glib.IdleAdd(func() {
		if m.ViewStack != nil {
			switcher.SetStack(m.ViewStack)
		}
	})
}
