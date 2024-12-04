package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/internal/ctxt"
)

type AdwToastOverlay struct {
	*Ref[*adw.ToastOverlay] `gtk:",ref"`
	Widget
	Child Model
}

func (m *AdwToastOverlay) Type() reflect.Type {
	return reflect.TypeFor[*adw.ToastOverlay]()
}

func (m *AdwToastOverlay) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewToastOverlay()
	m.Update(ctx, w)
	return w
}

func (m *AdwToastOverlay) Update(ctx context.Context, wi gtk.Widgetter) {
	w := wi.(*adw.ToastOverlay)
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))

	if m.Child != nil {
		if child := w.Child(); m.Child.Type() == reflect.TypeOf(child) {
			updateChild(child, m.Child)
		} else {
			w.SetChild(createChild(ctxt.With(ctx, w), m.Child))
		}
	}
}

// Add toast using overlay from context
func AddToast(ctx context.Context, toast *adw.Toast) {
	to := ctxt.MustFrom[*adw.ToastOverlay](ctx)
	to.AddToast(toast)
}
