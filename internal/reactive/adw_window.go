package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AdwWindow struct {
	Widget
	Content Model `gtk:"content"`
}

func (m *AdwWindow) Type() reflect.Type {
	return reflect.TypeFor[*adw.Window]()
}

func (model *AdwWindow) Create(ctx context.Context) gtk.Widgetter {
	return adw.NewWindow()
}

func (model *AdwWindow) Update(ctx context.Context, w gtk.Widgetter) {
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))
}
