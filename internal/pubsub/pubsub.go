package pubsub

import (
	"context"
	"sync"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/google/uuid"
)

type EventHandler[T any] func(T)

type Topic[T any] interface {
	Pub(value T)
	Sub(ctx context.Context, fn EventHandler[T])
}

func NewTopic[T any]() Topic[T] {
	return &topic[T]{}
}

type topic[T any] struct {
	mutex sync.RWMutex
	subs  map[string]EventHandler[T]
}

func (t *topic[T]) Sub(ctx context.Context, fn EventHandler[T]) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	uid := uuid.NewString()
	if t.subs == nil {
		t.subs = map[string]EventHandler[T]{}
	}
	t.subs[uid] = fn
	go func() {
		<-ctx.Done()
		t.mutex.Lock()
		defer t.mutex.Unlock()
		delete(t.subs, uid)
	}()
}

func (t *topic[T]) Pub(value T) {
	var mutex sync.Mutex
	for _, ev := range t.subs {
		glib.IdleAdd(func() {
			mutex.Lock()
			defer mutex.Unlock()
			ev(value)
		})
	}
}

type Property[T any] interface {
	Pub(value T)
	Sub(ctx context.Context, fn EventHandler[T])
	Value() T
}

func NewProperty[T any](value T) Property[T] {
	return &property[T]{
		value: value,
	}
}

type property[T any] struct {
	topic[T]
	value T
}

func (p *property[T]) Sub(ctx context.Context, fn EventHandler[T]) {
	p.topic.Sub(ctx, fn)
	fn(p.value)
}

func (p *property[T]) Pub(value T) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.value = value
	p.topic.Pub(value)
}

func (p *property[T]) Value() T {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.value
}
