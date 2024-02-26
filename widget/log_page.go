package widget

import (
	"context"
	"io"
	"runtime"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/util"
	corev1 "k8s.io/api/core/v1"
)

type LogPage struct {
	*adw.NavigationPage
}

func NewLogPage(ctx context.Context, cluster *api.Cluster, pod *corev1.Pod, container string) *LogPage {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	p := LogPage{NavigationPage: adw.NewNavigationPage(box, container)}

	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(runtime.GOOS != "windows")
	header.AddCSSClass("flat")
	box.Append(header)

	buffer := gtksource.NewBuffer(nil)
	util.SetSourceColorScheme(buffer)
	view := gtksource.NewViewWithBuffer(buffer)
	view.SetEditable(false)
	view.SetWrapMode(gtk.WrapWord)
	view.SetShowLineNumbers(true)
	view.SetMonospace(true)

	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetChild(view)
	scrolledWindow.SetVExpand(true)
	box.Append(scrolledWindow)

	logs, err := podLogs(ctx, cluster, pod, container)
	if err != nil {
		ShowErrorDialog(ctx, "Could not load logs", err)
	} else {
		buffer.SetText(string(logs))
	}

	return &p
}

func podLogs(ctx context.Context, cluster *api.Cluster, pod *corev1.Pod, container string) ([]byte, error) {
	req := cluster.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Container: container})
	r, err := req.Stream(ctx)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}
