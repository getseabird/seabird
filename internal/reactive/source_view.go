package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type SourceView struct {
	Widget
	Buffer          *gtksource.Buffer
	Editable        bool         `gtk:"editable"`
	ShowLineNumbers bool         `gtk:"show-line-numbers"`
	Monospace       bool         `gtk:"monospace"`
	WrapMode        gtk.WrapMode `gtk:"wrap-mode"`
}

func (m *SourceView) Type() reflect.Type {
	return reflect.TypeFor[*gtksource.View]()
}

func (model *SourceView) Create(ctx context.Context) gtk.Widgetter {
	w := gtksource.NewViewWithBuffer(model.Buffer)
	model.Update(ctx, w)
	return w
}

func (model *SourceView) Update(ctx context.Context, w gtk.Widgetter) {
	model.update(ctx, model, w, &model.Widget, gtk.BaseWidget(w))
}
