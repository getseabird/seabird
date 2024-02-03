package ui

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/jgillich/gotk4-vte/pkg/vte/v3"
)

type Terminal struct {
	*gtk.Revealer
}

func NewTerminal() *Terminal {
	t := Terminal{gtk.NewRevealer()}

	term := vte.NewTerminal()
	// pty, err := term.PtyNewSync(context.TODO(), vte.PtyDefault)
	// if err != nil {
	// 	panic(err)
	// }
	term.SetSizeRequest(0, 200)
	t.SetChild(term)

	return &t
}
