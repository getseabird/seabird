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

func (model *AdwViewStack) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewViewStack()
	model.Update(ctx, w)
	return w
}

func (model *AdwViewStack) Update(ctx context.Context, w gtk.Widgetter) {
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))
}
