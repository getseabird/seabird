package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AdwToolbarView struct {
	Widget
	Content        Model            `gtk:"content"`
	TopBarStyle    adw.ToolbarStyle `gtk:"top-bar-style"`
	BottomBarStyle adw.ToolbarStyle `gtk:"top-bar-style"`
	TopBars        []Model
	BottomBars     []Model
}

func (m *AdwToolbarView) Type() reflect.Type {
	return reflect.TypeFor[*adw.ToolbarView]()
}

func (model *AdwToolbarView) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewToolbarView()
	model.Update(ctx, w)

	for _, b := range model.TopBars {
		child := createChild(ctx, b)
		w.AddTopBar(child)
	}

	return w
}

func (model *AdwToolbarView) Update(ctx context.Context, wi gtk.Widgetter) {
	w := wi.(*adw.ToolbarView)
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))

}
