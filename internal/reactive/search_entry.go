package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type SearchEntry struct {
	Widget
	Editable        `gtk:",interface"`
	PlaceholderText string `gtk:"placeholder-text"`
}

func (m *SearchEntry) Type() reflect.Type {
	return reflect.TypeFor[*gtk.SearchEntry]()
}

func (m *SearchEntry) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewSearchEntry()
	m.Update(ctx, w)
	return w
}

func (m *SearchEntry) Update(ctx context.Context, wi gtk.Widgetter) {
	w := wi.(*gtk.SearchEntry)
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))
}
