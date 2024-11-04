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

func (model *AdwToastOverlay) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewToastOverlay()
	model.Update(ctx, w)
	return w
}

func (model *AdwToastOverlay) Update(ctx context.Context, wi gtk.Widgetter) {
	w := wi.(*adw.ToastOverlay)
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))

	if model.Child != nil {
		if child := w.Child(); model.Child.Type() == reflect.TypeOf(child) {
			updateChild(child, model.Child)
		} else {
			w.SetChild(createChild(ctxt.With(ctx, w), model.Child))
		}
	}
}

// Add toast using overlay from context
func AddToast(ctx context.Context, toast *adw.Toast) {
	to := ctxt.MustFrom[*adw.ToastOverlay](ctx)
	to.AddToast(toast)
}
