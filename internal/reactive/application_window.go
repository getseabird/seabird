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

func (model *ApplicationWindow) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewApplicationWindow(model.Application)
	model.Update(ctx, w)
	return w
}

func (model *ApplicationWindow) Update(ctx context.Context, w gtk.Widgetter) {
	model.update(ctx, model, w, &model.Window, &w.(*gtk.ApplicationWindow).Window)
}
