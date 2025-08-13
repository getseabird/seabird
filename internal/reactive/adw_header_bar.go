package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AdwHeaderBar struct {
	Widget
	TitleWidget           Model `gtk:"title-widget"`
	ShowStartTitleButtons *bool `gtk:"show-start-title-buttons,deref"`
	ShowEndTitleButtons   *bool `gtk:"show-end-title-buttons,deref"`
	Start                 []Model
	End                   []Model
}

func (m *AdwHeaderBar) Type() reflect.Type {
	return reflect.TypeFor[*adw.HeaderBar]()
}

func (m *AdwHeaderBar) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewHeaderBar()
	m.Update(ctx, w)
	for _, s := range m.Start {
		child := createChild(ctx, s)
		w.PackStart(child)
	}
	for _, s := range m.End {
		child := createChild(ctx, s)
		w.PackEnd(child)
	}

	return w
}

func (m *AdwHeaderBar) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))
	// bar := w.(*adw.HeaderBar)

}
