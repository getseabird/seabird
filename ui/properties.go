package ui

import (
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	corev1 "k8s.io/api/core/v1"
)

func podProperties(object *corev1.Pod) *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()
	group.SetTitle("Containers")

	for _, container := range object.Spec.Containers {
		row := adw.NewExpanderRow()
		row.SetTitle(container.Name)
		status := gtk.NewImageFromIconName("emblem-default-symbolic")
		status.AddCSSClass("container-status-ok")
		row.AddSuffix(status)
		group.Add(row)

		ar := adw.NewActionRow()
		ar.AddCSSClass("property")
		ar.SetTitle("Image")
		ar.SetSubtitle(container.Image)
		row.AddRow(ar)
		if len(container.Command) > 0 {
			ar = adw.NewActionRow()
			ar.AddCSSClass("property")
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
			ar.AddCSSClass("property")
			ar.SetTitle("Env")
			ar.SetSubtitle(strings.Join(env, " "))
			row.AddRow(ar)
		}
	}

	return group
}

func secretProperties(object *corev1.Secret) *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()
	group.SetTitle("Data")

	for key, value := range object.Data {
		row := adw.NewActionRow()
		row.AddCSSClass("property")
		row.SetTitle(key)
		row.SetSubtitle(string(value))
		group.Add(row)
	}

	return group
}

func configMapProperties(object *corev1.ConfigMap) *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()
	group.SetTitle("Data")

	for key, value := range object.Data {
		row := adw.NewActionRow()
		row.AddCSSClass("property")
		row.SetTitle(key)
		row.SetSubtitle(value)
		group.Add(row)
	}

	return group
}
