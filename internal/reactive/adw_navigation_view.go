package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AdwNavigationView struct {
	Widget
	Pages []AdwNavigationPage
}

func (m *AdwNavigationView) Type() reflect.Type {
	return reflect.TypeFor[*adw.NavigationView]()
}

func (model *AdwNavigationView) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewNavigationView()
	model.Update(ctx, w)
	return w
}

func (model *AdwNavigationView) Update(ctx context.Context, wi gtk.Widgetter) {
	w := wi.(*adw.NavigationView)
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))

	mergeChildren(
		ctx, w, Map(model.Pages, func(p AdwNavigationPage) Model { return &p }),
		func(c gtk.Widgetter, pos int) { w.Add(c.(*adw.NavigationPage)) },
		func(c gtk.Widgetter) { w.Remove(c.(*adw.NavigationPage)) },
	)
}
