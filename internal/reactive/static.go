package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func Static[T gtk.Widgetter](w T) Model {
	return &static[T]{w: w}
}

type static[T gtk.Widgetter] struct {
	Widget
	w T
}

func (m *static[T]) Type() reflect.Type {
	return reflect.TypeFor[T]()
}

func (model *static[T]) Create(ctx context.Context) gtk.Widgetter {
	return model.w
}

func (model *static[T]) Update(ctx context.Context, w gtk.Widgetter) {}
