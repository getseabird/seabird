//go:build !windows

package ui

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/creack/pty"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/behavior"
	"github.com/jgillich/gotk4-vte/pkg/vte/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/remotecommand"
)

type TerminalPage struct {
	*adw.NavigationPage
	pty *os.File
}

func NewTerminalPage(parent *gtk.Window, behavior *behavior.DetailBehavior, pod *corev1.Pod, container string) (w *TerminalPage) {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	nav := adw.NewNavigationPage(box, container)
	w = &TerminalPage{NavigationPage: nav}

	ctx, cancel := context.WithCancel(context.Background())
	nav.ConnectHidden(cancel)

	header := adw.NewHeaderBar()
	header.AddCSSClass("flat")
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
			ShowErrorDialog(parent, "Unable to open pty", err)
			return
		}
		defer tty.Close()

		go func() {
			// TODO there is a race condition here, not sure why
			time.Sleep(500 * time.Millisecond)
			glib.IdleAdd(func() {
				vtePty, err := vte.NewPtyForeignSync(ctx, int(pty.Fd()))
				if err != nil {
					ShowErrorDialog(parent, "Unable to open pty", err)
					return
				}
				terminal.SetPty(vtePty)
			})
		}()

		if err := behavior.PodExec(ctx, pod, container, []string{"/bin/sh"}, tty, tty, tty, nil); err != nil {
			if !errors.Is(err, context.Canceled) {
				glib.IdleAdd(func() {
					ShowErrorDialog(parent, "Exec failed", err)
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
