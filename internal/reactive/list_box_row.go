package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type ListBoxRow struct {
	Widget
	Activatable bool  `gtk:"activatable"`
	Child       Model `gtk:"child"`
	Selected    bool
}

func (m *ListBoxRow) Type() reflect.Type {
	return reflect.TypeFor[*gtk.ListBoxRow]()
}

func (m *ListBoxRow) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewListBoxRow()
	m.Update(ctx, w)
	return w
}

func (m *ListBoxRow) Update(ctx context.Context, wi gtk.Widgetter) {
	w := wi.(*gtk.ListBoxRow)
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))
}

func (m *ListBoxRow) PostUpdate(node Node) {
	w := node.Widget.(*gtk.ListBoxRow)
	p := node.Parent.Widget.(*gtk.ListBox)
	if m.Selected != w.IsSelected() {
		if m.Selected {
			p.SelectRow(w)
		} else {
			p.UnselectRow(w)
		}
	}
}
