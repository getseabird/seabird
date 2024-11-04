package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AdwPreferencesGroup struct {
	Widget
	Title        string `gtk:"title"`
	Children     []Model
	HeaderSuffix Model
}

func (m *AdwPreferencesGroup) Type() reflect.Type {
	return reflect.TypeFor[*adw.PreferencesGroup]()
}

func (model *AdwPreferencesGroup) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewPreferencesGroup()
	model.Update(ctx, w)
	return w
}

func (model *AdwPreferencesGroup) Update(ctx context.Context, w gtk.Widgetter) {
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))

	group := w.(*adw.PreferencesGroup)
	mergeChildren(ctx, w, model.Children, func(w gtk.Widgetter, pos int) {
		group.Add(w)
	}, func(w gtk.Widgetter) {
		group.Remove(w)
	})
}
