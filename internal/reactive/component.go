package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/internal/ctxt"
)

type Hook int

const (
	HookCreate Hook = iota
	HookUpdate
)

func CreateComponent(component Component) Model {
	return &ComponentModel{component: component}
}

type Component interface {
	Init(ctx context.Context)
	Update(ctx context.Context, message any) bool
	View(ctx context.Context) Model
	On(hook Hook, widget gtk.Widgetter)
}

type Sender[T any] interface {
	// Update self
	Self(ctx context.Context, updater func(component T))
	// stop after one?
	// Up(message any)
	// Down(message any)
	Broadcast(ctx context.Context, message any)
}

type BaseComponent[T any] struct {
}

func (c *BaseComponent[T]) Init(ctx context.Context) {}
func (c *BaseComponent[T]) Update(ctx context.Context, message any) bool {
	return false
}
func (c *BaseComponent[T]) On(hook Hook, widget gtk.Widgetter) {}

func (c *BaseComponent[T]) SetState(ctx context.Context, updater func(component T)) {
	node := ctxt.MustFrom[*Node](ctx)
	updater(node.component.(T))
	node.component.View(ctx).Update(ctx, node.Widget)
}

func (m *BaseComponent[T]) Broadcast(ctx context.Context, message any) {
	node := ctxt.MustFrom[*Node](ctx)
	node.ch <- message
}

type ComponentModel struct {
	Widget
	component Component
}

func (m *ComponentModel) Type() reflect.Type {
	return reflect.TypeFor[Component]()
}

func (m *ComponentModel) Create(ctx context.Context) gtk.Widgetter {
	// node := ctxt.MustFrom[*Node](ctx)
	// node.component = m.component
	m.component.Init(ctx)
	w := m.component.View(ctx).Create(ctx)
	// node.widget = w
	m.component.On(HookCreate, w)
	return w
}

func (m *ComponentModel) Update(ctx context.Context, w gtk.Widgetter) {
	// node := ctxt.MustFrom[*Node](ctx)
	// m.component = node.component
	model := m.component.View(ctx)
	// if model.Component() != nil && reflect.TypeOf(m.component) != reflect.TypeOf(model.Component()) {

	// }
	model.Update(ctx, w)
	// updateChild(w, )
	m.component.On(HookUpdate, w)
}

func (m *ComponentModel) Component() Component {
	return m.component
}
