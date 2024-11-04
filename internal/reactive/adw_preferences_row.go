package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AdwPreferencesRow struct {
	ListBoxRow
	Title           string `gtk:"title"`
	TitleSelectable bool   `gtk:"title-selectable"`
}

func (m *AdwPreferencesRow) Type() reflect.Type {
	return reflect.TypeFor[*adw.PreferencesRow]()
}

func (model *AdwPreferencesRow) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewPreferencesRow()
	model.Update(ctx, w)
	return w
}

func (model *AdwPreferencesRow) Update(ctx context.Context, w gtk.Widgetter) {
	model.update(ctx, model, w, &model.ListBoxRow, &w.(*adw.PreferencesRow).ListBoxRow)
}
