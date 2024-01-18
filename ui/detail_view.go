package ui

import (
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DetailView struct {
	*adw.PreferencesPage
	object client.Object
}

func NewDetailView(object client.Object) *DetailView {
	page := adw.NewPreferencesPage()
	page.SetSizeRequest(400, 100)
	page.SetHExpand(false)
	group := adw.NewPreferencesGroup()
	group.SetTitle("Metadata")
	group.Add(actionRow("Name", gtk.NewLabel(object.GetName())))
	group.Add(actionRow("Namespace", gtk.NewLabel(object.GetNamespace())))
	page.Add(group)

	switch object := object.(type) {
	case *corev1.Pod:
		group := adw.NewPreferencesGroup()
		group.SetTitle("Containers")
		page.Add(group)
		for _, container := range object.Spec.Containers {
			row := adw.NewExpanderRow()
			row.SetTitle(container.Name)
			status := gtk.NewImageFromIconName("emblem-default-symbolic")
			status.AddCSSClass("container-status-ok")
			row.AddAction(status)
			group.Add(row)

			ar := adw.NewActionRow()
			ar.SetTitle("Image")
			ar.SetSubtitle(container.Image)
			row.AddRow(ar)
			if len(container.Command) > 0 {
				ar = adw.NewActionRow()
				ar.SetTitle("Command")
				ar.SetSubtitle(strings.Join(container.Command, " "))
				row.AddRow(ar)
			}
			if len(container.Env) > 0 {
				var env []string
				for _, e := range container.Env {
					if e.ValueFrom != nil {
						// TODO
					} else {
						env = append(env, fmt.Sprintf("%s=%v", e.Name, e.Value))
					}
				}
				ar = adw.NewActionRow()
				ar.SetTitle("Env")
				ar.SetSubtitle(strings.Join(env, " "))
				row.AddRow(ar)
			}
		}
	}

	return &DetailView{PreferencesPage: page, object: object}
}

func actionRow(title string, suffix gtk.Widgetter) *adw.ActionRow {
	row := adw.NewActionRow()
	row.SetTitle(title)
	row.AddSuffix(suffix)
	return row
}
