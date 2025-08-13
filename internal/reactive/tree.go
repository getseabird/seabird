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

	root.Widget = model.Create(root.ctx)
	if c := model.Component(); c != nil {
		root.component = c
	}

	glib.Bind[*Node](root.Widget, root)

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

	return root.Widget
}

type Node struct {
	Parent *Node
	Widget gtk.Widgetter

	ch     chan any
	ctx    context.Context
	cancel context.CancelFunc
	// model     Model
	component      Component
	children       []*Node
	signalHandlers map[string]glib.SignalHandle
	// state          []any
}

func (n *Node) CreateChild(ctx context.Context, model Model) gtk.Widgetter {
	child := &Node{Parent: n, ch: n.ch}
	child.ctx, child.cancel = context.WithCancel(ctxt.With[*Node](ctx, child))

	child.Widget = model.Create(child.ctx)
	glib.Bind[*Node](child.Widget, child)
	n.children = append(n.children, child)

	if c := model.Component(); c != nil {
		child.component = c
	}

	return child.Widget
}

func (n *Node) RemoveChild(widget gtk.Widgetter) {
	node := *glib.Bounded[*Node](widget)
	node.cancel()
	if p := node.Parent; p != nil {
		for i, n := range node.children {
			if n.Widget == p.Widget {
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
			n.component.View(n.ctx).Update(n.ctx, n.Widget)
		}
	}
	for _, c := range n.children {
		c.message(msg, rerender)
	}
}
