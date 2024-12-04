package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type ApplicationWindow struct {
	Window      `gtk:",parent"`
	Application *gtk.Application
}

func (m *ApplicationWindow) Type() reflect.Type {
	return reflect.TypeFor[*adw.ApplicationWindow]()
}

func (m *ApplicationWindow) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewApplicationWindow(m.Application)
	m.Update(ctx, w)
	return w
}

func (m *ApplicationWindow) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.Window, &w.(*gtk.ApplicationWindow).Window)
}
