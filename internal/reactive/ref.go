package reactive

import (
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// type Ref[T gtk.Widgetter] = T

type Ref[T gtk.Widgetter] struct {
	Ref T
}

func (r *Ref[T]) Type() reflect.Type {
	return reflect.TypeFor[T]()
}
