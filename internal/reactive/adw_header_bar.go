package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AdwHeaderBar struct {
	Widget
	TitleWidget Model
}

func (m *AdwHeaderBar) Type() reflect.Type {
	return reflect.TypeFor[*adw.HeaderBar]()
}

func (model *AdwHeaderBar) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewHeaderBar()
	model.Update(ctx, w)
	return w
}

func (model *AdwHeaderBar) Update(ctx context.Context, w gtk.Widgetter) {
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))
	// bar := w.(*adw.HeaderBar)

}
