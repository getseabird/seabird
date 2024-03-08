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
	for _, prop := range keys {
		schema := schema.Value.Properties[prop]
		var required bool
		for _, r := range schema.Value.Required {
			if prop == r {
				required = true
				break
			}
		}
		row := adw.NewActionRow()
		row.SetUseMarkup(false)
		row.SetSubtitle(schema.Value.Description)

		var vtype string
		switch schema.Value.Type {
		case "array":
			vtype = fmt.Sprintf("[]%s", schema.Value.Items.Value.Type)
		default:
			vtype = schema.Value.Type
		}

		if inner := innerType(schema.Value); inner != nil {
			switch schema.Value.Type {
			case "array":
				vtype = fmt.Sprintf("[]%s", refName(inner.Ref))
			default:
				vtype = refName(inner.Ref)
			}
			row.SetActivatable(true)
			row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
			row.ConnectActivated(func() {
				page.Parent().(*adw.NavigationView).Push(newDocumentationPage(t, inner.Ref, append(breadcrumbs, page.Title())))
			})
		}
		title := fmt.Sprintf("%s %s", prop, vtype)
		if required {
			title = fmt.Sprintf("%s* %s", prop, vtype)
		}
		row.SetTitle(title)
		group.Add(row)
	}

	return page
}
