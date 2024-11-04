package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AdwActionRow struct {
	Ref[*adw.ActionRow]
	AdwPreferencesRow
	Subitle   string                         `gtk:"subtitle"`
	Activated func(actionRow *adw.ActionRow) `gtk:"activated,signal"`
	Prefixes  []Model
	Suffixes  []Model
}

func (m *AdwActionRow) Type() reflect.Type {
	return reflect.TypeFor[*adw.ActionRow]()
}

func (model *AdwActionRow) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewActionRow()
	model.Update(ctx, w)
	return w
}

func (model *AdwActionRow) Update(ctx context.Context, w gtk.Widgetter) {
	row := w.(*adw.ActionRow)
	model.update(ctx, model, w, &model.AdwPreferencesRow, &row.PreferencesRow)

	mergeChildren(
		ctx, row, model.Prefixes,
		func(w gtk.Widgetter, pos int) { row.AddPrefix(w) },
		func(w gtk.Widgetter) {}, // TODO remove
	)

	mergeChildren(
		ctx, row, model.Suffixes,
		func(w gtk.Widgetter, pos int) { row.AddSuffix(w) },
		func(w gtk.Widgetter) {}, // TODO remove
	)
}
