package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/internal/ctxt"
)

type AdwApplicationWindow struct {
	ApplicationWindow `gtk:",parent"`
	Content           Model
}

func (m *AdwApplicationWindow) Type() reflect.Type {
	return reflect.TypeFor[*adw.ApplicationWindow]()
}

func (m *AdwApplicationWindow) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewApplicationWindow(m.Application)
	m.Update(ctx, w)
	return w
}

func (m *AdwApplicationWindow) Update(ctx context.Context, wi gtk.Widgetter) {
	w := wi.(*adw.ApplicationWindow)
	m.update(ctx, m, w, &m.ApplicationWindow, &w.ApplicationWindow)

	if m.Content != nil {
		if content := w.Content(); m.Content.Type() == reflect.TypeOf(content) {
			updateChild(content, m.Content)
		} else {
			w.SetContent(createChild(ctxt.With(ctx, w.ApplicationWindow.Window), m.Content))
		}
	}
}
