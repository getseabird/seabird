package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AdwComboRow struct {
	Widget
	AdwPreferencesRow
	Model    gio.ListModeller
	Selected uint `gtk:"selected"`
}

func (m *AdwComboRow) Type() reflect.Type {
	return reflect.TypeFor[*adw.ComboRow]()
}

func (m *AdwComboRow) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewComboRow()
	w.SetModel(m.Model)
	m.Update(ctx, w)
	return w
}

func (m *AdwComboRow) Update(ctx context.Context, w gtk.Widgetter) {
	row := w.(*adw.ComboRow)
	m.update(ctx, m, w, &m.AdwPreferencesRow, &row.PreferencesRow)
}
