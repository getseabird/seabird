package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AdwPreferencesPage struct {
	Widget
	Groups []AdwPreferencesGroup
}

func (m *AdwPreferencesPage) Type() reflect.Type {
	return reflect.TypeFor[*adw.PreferencesPage]()
}

func (m *AdwPreferencesPage) Create(ctx context.Context) gtk.Widgetter {
	w := adw.NewPreferencesPage()
	m.Update(ctx, w)
	return w
}

func (m *AdwPreferencesPage) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))
	page := w.(*adw.PreferencesPage)

	mergeChildren[*adw.PreferencesGroup](
		ctx, page, Map(m.Groups, func(g AdwPreferencesGroup) Model { return &g }),
		func(w *adw.PreferencesGroup, pos int) { page.Add(w) },
		func(w *adw.PreferencesGroup) { page.Remove(w) },
	)

	// var groups []*adw.PreferencesGroup
	// if g := glib.Bounded[[]*adw.PreferencesGroup](page); g != nil {
	// 	groups = *g
	// }

	// for i, group := range model.Groups {
	// 	if len(groups) > i {
	// 		group.Update(ctx, groups[i])
	// 	} else {
	// 		new := createChild(ctx, &group).(*adw.PreferencesGroup)
	// 		page.Add(new)
	// 		groups = append(groups, new)
	// 	}
	// }
	// for i := len(groups); i > len(model.Groups); i-- {
	// 	page.Remove(groups[i-1])
	// 	groups = groups[:i-1]
	// }
	// glib.Bind(page, groups)
}
