package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/internal/ctxt"
)

type Hook int

const (
	HookCreate Hook = iota
	HookUpdate
)

type Component interface {
	Init(ctx context.Context, ch chan<- any)
	Update(ctx context.Context, message any, ch chan<- any) bool
	View(ctx context.Context, ch chan<- any) Model
	On(hook Hook, widget gtk.Widgetter)
}

type BaseComponent struct {
}

func (c *BaseComponent) Init(ctx context.Context, ch chan<- any) {}
func (c *BaseComponent) Update(ctx context.Context, message any, ch chan<- any) bool {
	return false
}
func (c *BaseComponent) On(hook Hook, widget gtk.Widgetter) {}

type ComponentModel struct {
	Widget
	component Component
}

func (m *ComponentModel) Type() reflect.Type {
	return reflect.TypeFor[*adw.Bin]()
}

func (c *ComponentModel) Create(ctx context.Context) gtk.Widgetter {
	node := ctxt.MustFrom[*Node](ctx)
	c.component.Init(ctx, node.ch)

	w := c.component.View(ctx, node.ch).Create(ctx)
	c.component.On(HookCreate, node.widget)
	return w
}

func (c *ComponentModel) Update(ctx context.Context, w gtk.Widgetter) {
	node := ctxt.MustFrom[*Node](ctx)

	if component := glib.Bounded[Component](w); component != nil {
		c.component = *component
	} else {
		glib.Bind(w, c.Component)
	}

	c.component.View(ctx, node.ch).Update(ctx, w)
	c.component.On(HookUpdate, w)
}

func (m *ComponentModel) Component() Component {
	return m.component
}

func CreateComponent(component Component) Model {
	return &ComponentModel{component: component}
}
