package ui

import (
	"fmt"
	"log"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	corev1 "k8s.io/api/core/v1"
)

func podProperties(pod *corev1.Pod) *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()
	group.SetTitle("Containers")

	for _, container := range pod.Spec.Containers {
		var status corev1.ContainerStatus
		for _, s := range pod.Status.ContainerStatuses {
			if s.Name == container.Name {
				status = s
				break
			}
		}

		expander := adw.NewExpanderRow()
		expander.SetTitle(container.Name)
		group.Add(expander)

		if status.Ready {
			icon := gtk.NewImageFromIconName("emblem-ok-symbolic")
			icon.AddCSSClass("success")
			expander.AddSuffix(icon)
		} else {
			icon := gtk.NewImageFromIconName("dialog-warning")
			icon.AddCSSClass("warning")
			expander.AddSuffix(icon)
		}

		row := adw.NewActionRow()
		row.AddCSSClass("property")
		row.SetTitle("Image")
		row.SetSubtitle(container.Image)
		expander.AddRow(row)
		if len(container.Command) > 0 {
			row = adw.NewActionRow()
			row.AddCSSClass("property")
			row.SetTitle("Command")
			row.SetSubtitle(strings.Join(container.Command, " "))
			expander.AddRow(row)
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
			row = adw.NewActionRow()
			row.AddCSSClass("property")
			row.SetTitle("Env")
			row.SetSubtitle(strings.Join(env, " "))
			expander.AddRow(row)
		}

		row = adw.NewActionRow()
		row.AddCSSClass("property")
		row.SetTitle("State")
		log.Printf("%v", status.State)
		if status.State.Running != nil {
			row.SetSubtitle("Running")
		} else if status.State.Terminated != nil {
			row.SetSubtitle(fmt.Sprintf("Terminated: %s", status.State.Terminated.Reason))
		} else if status.State.Waiting != nil {
			row.SetSubtitle(fmt.Sprintf("Waiting: %s", status.State.Waiting.Reason))
		}
		expander.AddRow(row)

		row = adw.NewActionRow()
		row.SetActivatable(true)
		row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
		row.SetTitle("Logs")
		row.ConnectActivated(func() {
			NewLogWindow(&application.window.Window, pod, &container).Show()
		})
		expander.AddRow(row)
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
