package behavior

import (
	"github.com/imkira/go-observer/v2"
)

func onChange[T any](prop observer.Property[T], f func(T)) {
	go func() {
		stream := prop.Observe()
		for {
			select {
			case <-stream.Changes():
				stream.Next()
				f(stream.Value())
			}
		}
	}()
}
