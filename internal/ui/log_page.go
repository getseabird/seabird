package ui

import (
	"runtime"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/internal/behavior"
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

	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(runtime.GOOS != "windows")
	header.AddCSSClass("flat")
	box.Append(header)

	buffer := gtksource.NewBuffer(nil)
	buffer.SetStyleScheme(gtksource.StyleSchemeManagerGetDefault().Scheme("Adwaita-dark"))
	view := gtksource.NewViewWithBuffer(buffer)
	view.SetMarginBottom(8)
	view.SetMarginTop(8)
	view.SetMarginEnd(8)
	view.SetEditable(false)
	view.SetWrapMode(gtk.WrapWord)
	view.SetShowLineNumbers(true)
	view.SetMonospace(true)

	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetChild(view)
	scrolledWindow.SetVExpand(true)
	box.Append(scrolledWindow)

	logs, err := behavior.PodLogs(pod, container)
	if err != nil {
		ShowErrorDialog(parent, "Could not load logs", err)
	} else {
		buffer.SetText(string(logs))
	}

	return &p
}
