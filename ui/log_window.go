package ui

import (
	"context"
	"io"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	corev1 "k8s.io/api/core/v1"
)

type LogWindow struct {
	*adw.PreferencesWindow
	pod       *corev1.Pod
	container *corev1.Container
}

func NewLogWindow(parent *gtk.Window, pod *corev1.Pod, container *corev1.Container) *LogWindow {
	w := LogWindow{PreferencesWindow: adw.NewPreferencesWindow()}
	w.SetTransientFor(parent)
	w.SetDefaultSize(800, 800)

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	w.SetContent(box)

	header := adw.NewHeaderBar()
	header.SetTitleWidget(gtk.NewLabel(container.Name))
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

	w.ConnectShow(func() {
		req := application.cluster.Clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
		r, err := req.Stream(context.TODO())
		if err != nil {
			ShowErrorDialog(&w.Window.Window, "Could not load logs", err)
			return
		}
		logs, err := io.ReadAll(r)
		if err != nil {
			ShowErrorDialog(&w.Window.Window, "Could not load logs", err)
			return
		}
		buffer.SetText(string(logs))
	})

	return &w
}
