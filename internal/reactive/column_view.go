package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type ColumnView struct {
	*Ref[*gtk.ColumnView] `gtk:",ref"`
	Widget
	Model gtk.SelectionModeller
	// ColumnView columns. ID must be set for merging to work
	Columns              []*gtk.ColumnViewColumn
	SingleClickActivate  bool                                           `gtk:"single-click-activate"`
	ShowColumnSeperators bool                                           `gtk:"show-column-separators"`
	ShowRowSeperators    bool                                           `gtk:"show-row-separators"`
	Activate             func(columnView *gtk.ColumnView, position int) `gtk:"activate,signal"`
}

func (m *ColumnView) Type() reflect.Type {
	return reflect.TypeFor[*gtk.ColumnView]()
}

func (m *ColumnView) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewColumnView(m.Model)
	m.Update(ctx, w)
	return w
}

func (m *ColumnView) Update(ctx context.Context, wi gtk.Widgetter) {
	w := wi.(*gtk.ColumnView)
	m.update(ctx, m, w, &m.Widget, &w.Widget)
	m.mergeColumns(w)
}

func (m *ColumnView) mergeColumns(w *gtk.ColumnView) {
	current := w.Columns()
	for i, c := range m.Columns {
		if int(current.NItems()) > i && current.Item(uint(i)).Cast().(*gtk.ColumnViewColumn).ID() == c.ID() {
			continue
		}
		w.InsertColumn(uint(i), c)
	}

	for i := w.Columns().NItems() - 1; int(i) > len(m.Columns); i-- {
		w.RemoveColumn(current.Item(uint(i)).Cast().(*gtk.ColumnViewColumn))
	}
}
