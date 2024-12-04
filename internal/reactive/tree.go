package reactive

import (
	"context"
	"slices"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/internal/ctxt"
)

func NewTree(ctx context.Context, model Model) gtk.Widgetter {
	root := &Node{
		ch:  make(chan any),
		ctx: ctx,
	}
	root.ctx = ctxt.With[*Node](ctx, root)

	root.widget = model.Create(root.ctx)
	if c := model.Component(); c != nil {
		root.component = c
	}

	glib.Bind[*Node](root.widget, root)

	go func() {
		for {
			select {
			case msg := <-root.ch:
				glib.IdleAdd(func() {
					root.message(msg, true)
				})
			case <-ctx.Done():
				return
			}
		}
	}()

	return root.widget
}

type Node struct {
	ch     chan any
	parent *Node
	ctx    context.Context
	cancel context.CancelFunc
	// model     Model
	widget         gtk.Widgetter
	component      Component
	children       []*Node
	signalHandlers map[string]glib.SignalHandle
	// state          []any
}

func (n *Node) CreateChild(ctx context.Context, model Model) gtk.Widgetter {
	child := &Node{parent: n, ch: n.ch}
	child.ctx, child.cancel = context.WithCancel(ctxt.With[*Node](ctx, child))

	child.widget = model.Create(child.ctx)
	glib.Bind[*Node](child.widget, child)
	n.children = append(n.children, child)

	if c := model.Component(); c != nil {
		child.component = c
	}

	return child.widget
}

func (n *Node) RemoveChild(widget gtk.Widgetter) {
	node := *glib.Bounded[*Node](widget)
	node.cancel()
	if p := node.parent; p != nil {
		for i, n := range node.children {
			if n.widget == p.widget {
				node.children = slices.Delete(node.children, i, i+1)
				break
			}
		}
	}
}

func (n *Node) message(msg any, rerender bool) {
	if n.component != nil {
		if n.component.Update(n.ctx, msg) && rerender {
			rerender = false
			n.component.View(n.ctx).Update(n.ctx, n.widget)
		}
	}
	for _, c := range n.children {
		c.message(msg, rerender)
	}
}
