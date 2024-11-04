package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type ToggleButton struct {
	Button
}

func (m *ToggleButton) Type() reflect.Type {
	return reflect.TypeFor[*gtk.ToggleButton]()
}

func (model *ToggleButton) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewToggleButton()
	model.Update(ctx, w)
	return w
}

func (model *ToggleButton) Update(ctx context.Context, w gtk.Widgetter) {
	model.update(ctx, model, w, &model.Button, &w.(*gtk.ToggleButton).Button)
}
