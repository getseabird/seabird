package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

type Label struct {
	Widget
	Label     string              `gtk:"label"`
	Ellipsize pango.EllipsizeMode `gtk:"ellipsize"`
	Justify   gtk.Justification   `gtk:"justify"`
}

func (m *Label) Type() reflect.Type {
	return reflect.TypeFor[*gtk.Label]()
}

func (m *Label) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewLabel(m.Label)
	m.Update(ctx, w)
	return w
}

func (m *Label) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))
}
