package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/behavior"
	corev1 "k8s.io/api/core/v1"
)

type LogPage struct {
	*adw.NavigationPage
	parent    *gtk.Window
	pod       *corev1.Pod
	container *corev1.Container
}

func NewLogPage(parent *gtk.Window, behavior *behavior.DetailBehavior, pod *corev1.Pod, container string) *LogPage {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	p := LogPage{NavigationPage: adw.NewNavigationPage(box, container)}
	p.SetSizeRequest(350, 350)

	header := adw.NewHeaderBar()
	header.SetTitleWidget(gtk.NewLabel(container))
	header.AddCSSClass("flat")
	box.Append(header)

	buffer := gtksource.NewBuffer(nil)
	buffer.SetStyleScheme(gtksource.StyleSchemeManagerGetDefault().Scheme("Adwaita-dark"))
	view := gtksource.NewViewWithBuffer(buffer)
	view.SetMarginBottom(8)
	view.SetMarginTop(8)
	view.SetMarginStart(8)
	view.SetMarginEnd(8)
	view.SetEditable(false)
	view.SetVExpand(true)

	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetChild(view)
	box.Append(scrolledWindow)

	logs, err := behavior.PodLogs(pod, container)
	if err != nil {
		ShowErrorDialog(parent, "Could not load logs", err)
	} else {
		buffer.SetText(string(logs))
	}

	return &p
}
