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

func (model *AdwBin) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewBin()
	model.Update(ctx, w)
	return w
}

func (model *AdwBin) Update(ctx context.Context, w gtk.Widgetter) {
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))
}
