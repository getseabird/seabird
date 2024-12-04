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

func (m *AdwPreferencesRow) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewPreferencesRow()
	m.Update(ctx, w)
	return w
}

func (m *AdwPreferencesRow) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.ListBoxRow, &w.(*adw.PreferencesRow).ListBoxRow)
}
