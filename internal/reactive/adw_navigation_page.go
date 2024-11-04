package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AdwNavigationPage struct {
	Widget
	Title string `gtk:"title"`
	Child Model  `gtk:"child"`
}

func (m *AdwNavigationPage) Type() reflect.Type {
	return reflect.TypeFor[*adw.NavigationPage]()
}

func (model *AdwNavigationPage) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewNavigationPage(createChild(ctx, model.Child), model.Title)
	model.Update(ctx, w)
	return w
}

func (model *AdwNavigationPage) Update(ctx context.Context, wi gtk.Widgetter) {
	w := wi.(*adw.NavigationPage)
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))
}
