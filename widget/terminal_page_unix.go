//go:build !windows

package widget

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/creack/pty"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/jgillich/gotk4-vte/pkg/vte/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

type TerminalPage struct {
	*adw.NavigationPage
	pty *os.File
}

func NewTerminalPage(ctx context.Context, cluster *api.Cluster, pod *corev1.Pod, container string) (w *TerminalPage) {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	nav := adw.NewNavigationPage(box, container)
	w = &TerminalPage{NavigationPage: nav}

	ctx, cancel := context.WithCancel(ctx)
	nav.ConnectHidden(cancel)

	header := adw.NewHeaderBar()
	header.AddCSSClass("flat")
	header.SetShowStartTitleButtons(false)
	box.Append(header)

	terminal := vte.NewTerminal()
	terminal.SetHExpand(true)
	terminal.SetVExpand(true)
	box.Append(terminal)

	// TODO detect size changes and forward
	// sizeQueue := make(sizeQueue, 1)
	// go func() {
	// 	<-ctx.Done()
	// 	close(sizeQueue)
	// }()
	// sizeQueue <- remotecommand.TerminalSize{Width: uint16(terminal.Width()), Height: uint16(terminal.Height())}
	// parent.NotifyProperty("default-width", func() {
	// 	size := remotecommand.TerminalSize{Width: uint16(terminal.Width()), Height: uint16(terminal.Height())}
	// 	select {
	// 	case sizeQueue.ch <- size:
	// 	default:
	// 	}
	// })

	go func() {
		pty, tty, err := pty.Open()
		if err != nil {
			ShowErrorDialog(ctx, "Unable to open pty", err)
			return
		}
		defer tty.Close()

		go func() {
			// TODO there is a race condition here, not sure why
			time.Sleep(500 * time.Millisecond)
			glib.IdleAdd(func() {
				vtePty, err := vte.NewPtyForeignSync(ctx, int(pty.Fd()))
				if err != nil {
					ShowErrorDialog(ctx, "Unable to open pty", err)
					return
				}
				terminal.SetPty(vtePty)
			})
		}()

		if err := podExec(ctx, cluster, pod, container, []string{"/bin/sh"}, tty, tty, tty, nil); err != nil {
			if !errors.Is(err, context.Canceled) {
				glib.IdleAdd(func() {
					ShowErrorDialog(ctx, "Exec failed", err)
				})
			}
		}
	}()

	return
}

type sizeQueue chan remotecommand.TerminalSize

func (s *sizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-*s
	if !ok {
		return nil
	}
	return &size
}

func podExec(ctx context.Context, cluster *api.Cluster, pod *corev1.Pod, container string, command []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, sizeQueue remotecommand.TerminalSizeQueue) error {
	req := cluster.CoreV1().RESTClient().Post().Resource("pods").Name(pod.Name).Namespace(pod.Namespace).SubResource("exec")
	option := &corev1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)

	spdy, err := remotecommand.NewSPDYExecutor(cluster.Config, "POST", req.URL())
	if err != nil {
		return err
	}
	ws, err := remotecommand.NewWebSocketExecutor(cluster.Config, "GET", req.URL().String())
	if err != nil {
		return err
	}
	exec, err := remotecommand.NewFallbackExecutor(ws, spdy, httpstream.IsUpgradeFailure)
	if err != nil {
		return err
	}

	return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stderr,
		Tty:               true,
		TerminalSizeQueue: sizeQueue,
	})
}
