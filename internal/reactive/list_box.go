package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type ListBox struct {
	Widget
	RowSelected func(listBox *gtk.ListBox, listBoxRow *gtk.ListBoxRow) `gtk:"row-selected,signal"`
	Children    []Model
}

func (m *ListBox) Type() reflect.Type {
	return reflect.TypeFor[*gtk.ListBox]()
}

func (m *ListBox) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewListBox()
	m.Update(ctx, w)
	return w
}

func (m *ListBox) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))
	box := w.(*gtk.ListBox)

	next := box.FirstChild()
	for _, child := range m.Children {
		if next == nil {
			new := createChild(ctx, child)
			box.Append(new)
			continue
		}

		if child.Type() == reflect.TypeOf(next) {
			child.Update(ctx, next)
			next = gtk.BaseWidget(next).NextSibling()
		} else {
			new := createChild(ctx, child)
			gtk.BaseWidget(new).InsertBefore(box, next)
			removeChild(next)
			box.Remove(next)
			next = gtk.BaseWidget(new).NextSibling()
		}
	}

	for {
		if next == nil {
			break
		}
		sibling := gtk.BaseWidget(next).NextSibling()
		box.Remove(next)
		next = sibling
	}
}
