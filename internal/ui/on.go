package ui

import (
	"context"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/imkira/go-observer/v2"
)

func onChange[T any](ctx context.Context, prop observer.Property[T], f func(T)) {
	go func() {
		stream := prop.Observe()
		for {
			select {
			case <-stream.Changes():
				stream.Next()
				glib.IdleAdd(func() {
					f(stream.Value())
				})
			case <-ctx.Done():
				return
			}
		}
	}()
}
