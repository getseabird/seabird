package editor

import (
	"fmt"
	"slices"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/exp/maps"
)

func newDocumentationPage(t *openapi3.T, ref string, breadcrumbs []string) *adw.NavigationPage {
	schema := resolveRef(t, ref)

	content := gtk.NewBox(gtk.OrientationVertical, 0)
	page := adw.NewNavigationPage(content, refName(ref))

	prefs := adw.NewPreferencesPage()
	content.Append(prefs)
	group := adw.NewPreferencesGroup()

	header := gtk.NewBox(gtk.OrientationHorizontal, 0)
	header.AddCSSClass("linked")
	header.SetMarginBottom(12)
	group.FirstChild().(*gtk.Box).Prepend(header)
	for i, b := range breadcrumbs {
		btn := gtk.NewToggleButtonWithLabel(b)
		btn.ConnectClicked(func() {
			for range len(breadcrumbs) - i {
				page.Parent().(*adw.NavigationView).Pop()
			}
		})
		btn.FirstChild().(*gtk.Label).SetEllipsize(pango.EllipsizeEnd)
		header.Append(btn)
	}
	title := gtk.NewToggleButtonWithLabel(page.Title())
	title.SetActive(true)
	title.ConnectClicked(func() { title.SetActive(true) })
	title.FirstChild().(*gtk.Label).SetEllipsize(pango.EllipsizeEnd)
	header.Append(title)

	group.SetTitle(page.Title())
	group.SetDescription(schema.Value.Description)
	prefs.Add(group)

	keys := maps.Keys(schema.Value.Properties)
	slices.Sort(keys)
	for _, field := range keys {
		schema := schema.Value.Properties[field]
		var required bool
		for _, r := range schema.Value.Required {
			if field == r {
				required = true
				break
			}
		}
		row := createPropRow(field, schema, required, func(ref string) {
			page.Parent().(*adw.NavigationView).Push(newDocumentationPage(t, ref, append(breadcrumbs, page.Title())))
		})
		group.Add(row)
	}

	return page
}

// TODO this can create nested expander rows, which are not supported in adw
func createPropRow(field string, schema *openapi3.SchemaRef, required bool, onActivate func(ref string)) gtk.Widgetter {
	if typeName(schema.Value) == "object" && len(schema.Value.Properties) > 0 {
		expander := adw.NewExpanderRow()
		expander.SetUseMarkup(false)
		expander.SetSubtitle(schema.Value.Description)
		title := fmt.Sprintf("%s %s", field, typeName(schema.Value))
		if required {
			title = fmt.Sprintf("%s* %s", field, typeName(schema.Value))
		}
		expander.SetTitle(title)
		keys := maps.Keys(schema.Value.Properties)
		slices.Sort(keys)
		for _, field := range keys {
			schema := schema.Value.Properties[field]
			var required bool
			for _, r := range schema.Value.Required {
				if field == r {
					required = true
					break
				}
			}
			expander.AddRow(createPropRow(field, schema, required, onActivate))
		}
		return expander
	} else {
		row := adw.NewActionRow()
		row.SetUseMarkup(false)
		row.SetSubtitle(schema.Value.Description)
		title := fmt.Sprintf("%s %s", field, typeName(schema.Value))
		if required {
			title = fmt.Sprintf("%s* %s", field, typeName(schema.Value))
		}
		row.SetTitle(title)
		if inner := innerType(schema.Value); inner != nil {
			row.SetActivatable(true)
			row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
			row.ConnectActivated(func() {
				onActivate(inner.Ref)
			})
		}
		return row
	}
}

func typeName(v *openapi3.Schema) string {
	if inner := innerType(v); inner != nil {
		switch {
		case v.Type.Is(openapi3.TypeArray):
			return fmt.Sprintf("[]%s", refName(inner.Ref))
		default:
			return refName(inner.Ref)
		}
	}

	switch {
	case v.Type.Is(openapi3.TypeArray):
		return fmt.Sprintf("[]%s", v.Items.Value.Type)
	default:
		return v.Type.Slice()[0]
	}
}
