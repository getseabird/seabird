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

func (model *AdwPreferencesWindow) Create(ctx context.Context) gtk.Widgetter {
	return adw.NewPreferencesWindow()
}

func (model *AdwPreferencesWindow) Update(ctx context.Context, w gtk.Widgetter) {
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))
}
