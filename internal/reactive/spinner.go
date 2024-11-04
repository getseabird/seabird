package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Spinner struct {
	*Ref[*gtk.Spinner] `gtk:",ref"`
	Spinning           bool `gtk:"spinning"`
	Widget
}

func (m *Spinner) Type() reflect.Type {
	return reflect.TypeFor[*gtk.Spinner]()
}

func (m *Spinner) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewSpinner()
	m.Update(ctx, w)
	return w
}

func (m *Spinner) Update(ctx context.Context, wi gtk.Widgetter) {
	w := wi.(*gtk.Spinner)
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))
	if !m.Spinning {
		w.Stop()
	}
}
