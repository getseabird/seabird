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

func (m *ToggleButton) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewToggleButton()
	m.Update(ctx, w)
	return w
}

func (m *ToggleButton) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.Button, &w.(*gtk.ToggleButton).Button)
}
