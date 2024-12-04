package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AdwPreferencesWindow struct {
	Widget
	AdwWindow
}

func (m *AdwPreferencesWindow) Type() reflect.Type {
	return reflect.TypeFor[*adw.PreferencesWindow]()
}

func (m *AdwPreferencesWindow) Create(ctx context.Context) gtk.Widgetter {
	return adw.NewPreferencesWindow()
}

func (m *AdwPreferencesWindow) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))
}
