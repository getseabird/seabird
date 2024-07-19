package editor

import (
	"fmt"
	"html"
	"slices"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/exp/maps"
	"k8s.io/klog/v2"
)

func newDocumentationPage(t *openapi3.T, ref string, breadcrumbs []string) *adw.NavigationPage {
	resolver := resolver{t}
	schema := resolver.resolve(ref)
	if schema == nil {
		klog.Warningf("failed to resolve schema '%v'", ref)
		return nil
	}

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
	group.SetDescription(html.EscapeString(schema.Value.Description))
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
		row := createPropRow(resolver, field, schema, required, func(ref string) {
			p := newDocumentationPage(t, ref, append(breadcrumbs, page.Title()))
			if p != nil {
				page.Parent().(*adw.NavigationView).Push(p)
			}
		})
		group.Add(row)
	}

	return page
}

// TODO this can create nested expander rows, which are not supported in adw
func createPropRow(resolver resolver, field string, schema *openapi3.SchemaRef, required bool, onActivate func(ref string)) gtk.Widgetter {
	title := fmt.Sprintf("<b>%s</b> %s", field, resolver.typeName(schema))
	if required {
		title = fmt.Sprintf("<b>%s*</b> %s", field, resolver.typeName(schema))
	}

	if len(schema.Value.Properties) > 0 ||
		schema.Value.Items != nil && schema.Value.Items.Value.Type.Includes(openapi3.TypeObject) && schema.Value.Items.Value.Properties != nil {
		expander := adw.NewExpanderRow()
		expander.SetSubtitle(html.EscapeString(schema.Value.Description))

		expander.SetTitle(title)

		properties := schema.Value.Properties
		if len(properties) == 0 {
			properties = schema.Value.Items.Value.Properties
		}

		keys := maps.Keys(properties)
		slices.Sort(keys)
		for _, field := range keys {
			schema := properties[field]
			var required bool
			for _, r := range schema.Value.Required {
				if field == r {
					required = true
					break
				}
			}
			expander.AddRow(createPropRow(resolver, field, schema, required, onActivate))
		}
		return expander
	} else {
		row := adw.NewActionRow()
		row.SetSubtitle(html.EscapeString(schema.Value.Description))
		row.SetTitle(title)

		for _, subtype := range resolver.subtypes(schema) {
			if subtype.Ref == "" {
				continue
			}
			// TODO selection for multiple types?
			row.SetActivatable(true)
			row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
			row.ConnectActivated(func() {
				onActivate(subtype.Ref)
			})
			break
		}
		return row
	}
}
