package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/internal/ctxt"
)

type AdwAlertDialog struct {
	Ref[*adw.AlertDialog]
	Widget
	Visible         bool
	Responses       map[string]string
	Heading         string                           `gtk:"heading"`
	Body            string                           `gtk:"body"`
	DefaultResponse string                           `gtk:"default-response"`
	Closed          func(actionRow *adw.AlertDialog) `gtk:"closed,signal"`
}

func (m *AdwAlertDialog) Type() reflect.Type {
	return reflect.TypeFor[*adw.AlertDialog]()
}

func (m *AdwAlertDialog) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewAlertDialog(m.Heading, m.Body)
	m.Update(ctx, w)
	return w
}

func (m *AdwAlertDialog) Update(ctx context.Context, wi gtk.Widgetter) {
	w := wi.(*adw.AlertDialog)
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))

	if m.Visible {
		w.Present(ctxt.MustFrom[*Node](ctx).Parent.Widget)
	}
}
