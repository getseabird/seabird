package reactive

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/fatih/structtag"
	"github.com/getseabird/seabird/internal/ctxt"
)

type Model interface {
	Type() reflect.Type
	Create(ctx context.Context) gtk.Widgetter
	Update(ctx context.Context, widget gtk.Widgetter)
	Component() Component
	Connect(string, any)
}

type Widget struct {
	// margin top end bottom start
	Margin      [4]int
	VExpand     bool      `gtk:"vexpand"`
	HExpand     bool      `gtk:"hexpand"`
	HAlign      gtk.Align `gtk:"halign"`
	VAlign      gtk.Align `gtk:"valign"`
	Name        string    `gtk:"name"`
	Opacity     float64   `gtk:"opacity"`
	TooltipText string    `gtk:"tooltip-text"`
	//`gtk:"css-classes"` type []string not implemented
	CSSClasses []string
	Signals    map[string]any
}

func (m *Widget) Type() reflect.Type {
	return nil
}

func (m *Widget) Create(ctx context.Context) gtk.Widgetter {
	return nil
}

func (m *Widget) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, nil, nil)

	node := ctxt.MustFrom[*Node](ctx)

	w.SetObjectProperty("margin-top", m.Margin[0])
	w.SetObjectProperty("margin-end", m.Margin[1])
	w.SetObjectProperty("margin-bottom", m.Margin[2])
	w.SetObjectProperty("margin-start", m.Margin[3])

	if node.signalHandlers == nil {
		node.signalHandlers = map[string]glib.SignalHandle{}
	}

	for signal, callback := range m.Signals {
		if handler, ok := node.signalHandlers[signal]; ok {
			w.HandlerDisconnect(handler)
		}
		if connect := reflect.ValueOf(w).MethodByName("Connect"); connect.IsValid() {
			// pass widget reference to callbacks
			cb := reflect.TypeOf(callback)
			var in []reflect.Type
			for i := 1; i < cb.NumIn(); i++ {
				in = append(in, cb.In(i))
			}
			var out []reflect.Type
			for i := 0; i < cb.NumOut(); i++ {
				out = append(in, cb.Out(i))
			}
			ret := connect.Call(
				[]reflect.Value{reflect.ValueOf(signal),
					reflect.MakeFunc(reflect.FuncOf(in, out, false), func(args []reflect.Value) (results []reflect.Value) {
						return reflect.ValueOf(callback).Call(append([]reflect.Value{reflect.ValueOf(node.widget)}, args...))
					}),
				})
			node.signalHandlers[signal] = ret[0].Interface().(glib.SignalHandle)
		}
	}

	for _, class := range m.CSSClasses {
		reflect.ValueOf(w).MethodByName("AddCSSClass").Call([]reflect.Value{reflect.ValueOf(class)})
	}
}

func (m *Widget) Component() Component {
	return nil
}

func (m *Widget) Connect(s string, h any) {
	if m.Signals == nil {
		m.Signals = map[string]any{}
	}
	m.Signals[s] = h
}

func createChild(ctx context.Context, model Model) gtk.Widgetter {
	node := ctxt.MustFrom[*Node](ctx)
	return node.CreateChild(ctx, model)
}

func updateChild(widget gtk.Widgetter, model Model) {
	node := *glib.Bounded[*Node](widget)
	model.Update(node.ctx, widget)
}

func removeChild(widget gtk.Widgetter) {
	node := *glib.Bounded[*Node](widget)
	node.parent.RemoveChild(widget)
}

func (m *Widget) update(ctx context.Context, model Model, w gtk.Widgetter, parent Model, parentW gtk.Widgetter) {
	if parent != nil {
		defer parent.Update(ctx, parentW)
	}

	val := reflect.Indirect(reflect.ValueOf(model))
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tags, err := structtag.Parse(string(field.Tag))
		if err != nil {
			panic(err)
		}

		tag, err := tags.Get("gtk")
		if err != nil {
			continue
		}

		if val.Field(i).IsZero() {
			continue
		}

		if slices.Contains(tag.Options, "interface") {
			vv := val.Field(i)
			tt := vv.Type()
			for i := 0; i < tt.NumField(); i++ {
				field := tt.Field(i)
				tags, err := structtag.Parse(string(field.Tag))
				if err != nil {
					panic(err)
				}
				tag, err := tags.Get("gtk")
				if err != nil || vv.Field(i).IsZero() {
					continue
				}
				w.SetObjectProperty(tag.Name, vv.Field(i).Interface())
			}
			continue
		}

		if slices.Contains(tag.Options, "ref") {
			val.Field(i).Elem().FieldByName("Ref").Set(reflect.ValueOf(w))
			continue
		}

		if slices.Contains(tag.Options, "signal") {
			model.Connect(tag.Name, val.Field(i).Interface())
			continue
		}

		if slices.Contains(tag.Options, "parent") {
			// model := v.Field(i).Addr().Interface().(Model)
			// TODO reflect.Value.Addr of unaddressable value
			// model.Update(ctx, reflect.Indirect(reflect.ValueOf(w)).FieldByName(field.Name).Addr().Interface().(gtk.Widgetter))
			continue
		}

		if val.Field(i).Type() == reflect.TypeFor[Model]() {
			model := val.Field(i).Interface().(Model)
			if val := reflect.ValueOf(w).MethodByName(field.Name).Call([]reflect.Value{}); val[0].IsValid() && !val[0].IsNil() && reflect.ValueOf(val[0].Interface()).Type() == model.Type() {
				updateChild(val[0].Interface().(gtk.Widgetter), model)
			} else {
				val := reflect.ValueOf(createChild(ctx, model))
				reflect.ValueOf(w).MethodByName(fmt.Sprintf("Set%v", field.Name)).Call([]reflect.Value{val})
			}
		} else {
			val := val.Field(i).Interface()
			w.SetObjectProperty(tag.Name, val)
		}
	}
}

func mergeChildren[T gtk.Widgetter](ctx context.Context, w gtk.Widgetter, models []Model, add func(w T, pos int), remove func(w T)) {
	// var children []T
	// if g := glib.Bounded[[]T](w); g != nil {
	// 	children = *g
	// }
	node := ctxt.MustFrom[*Node](ctx)

	for i, model := range models {
		if len(node.children) > i && model.Type() == reflect.TypeOf(node.children[i].widget) {
			updateChild(node.children[i].widget, model)
		} else {
			new := createChild(ctx, model).(T)
			add(new, i)
			// node.children = append(children, new)
		}
	}
	for i := len(node.children); i > len(models); i-- {
		w := node.children[i-1].widget.(T)
		removeChild(w)
		remove(w)
	}
	// glib.Bind(w, children)
}
