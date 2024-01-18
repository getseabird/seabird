package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DetailView struct {
	*adw.PreferencesPage
	object client.Object
}

func NewDetailView(object client.Object) *DetailView {
	page := adw.NewPreferencesPage()
	page.SetHExpand(true)

	group := adw.NewPreferencesGroup()
	group.SetTitle("Metadata")
	group.Add(actionRow("Name", gtk.NewLabel(object.GetName())))
	group.Add(actionRow("Namespace", gtk.NewLabel(object.GetNamespace())))
	page.Add(group)

	return &DetailView{PreferencesPage: page, object: object}
}

func (d *DetailView) init() {
	// for {
	// 	child := d.FirstChild()
	// 	if child == nil {
	// 		break
	// 	}
	// 	d.Remove(child.(*adw.PreferencesGroup))
	// }

}

func actionRow(title string, suffix gtk.Widgetter) *adw.ActionRow {
	row := adw.NewActionRow()
	row.SetTitle(title)
	row.AddSuffix(suffix)
	return row
}
