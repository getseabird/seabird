package editor

import (
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getkin/kin-openapi/openapi3"
)

func newDocumentationPage(t *openapi3.T, ref string) *adw.NavigationPage {
	schema := resolveRef(t, ref)

	content := gtk.NewBox(gtk.OrientationVertical, 0)
	page := adw.NewNavigationPage(content, refName(ref))

	header := adw.NewHeaderBar()
	header.AddCSSClass("flat")
	header.SetShowStartTitleButtons(false)
	header.SetShowEndTitleButtons(false)
	content.Append(header)

	prefs := adw.NewPreferencesPage()
	content.Append(prefs)
	group := adw.NewPreferencesGroup()
	group.SetDescription(schema.Value.Description)
	prefs.Add(group)

	for prop, schema := range schema.Value.Properties {
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
				page.Parent().(*adw.NavigationView).Push(newDocumentationPage(t, inner.Ref))
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
