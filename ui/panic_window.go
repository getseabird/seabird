package ui

import (
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// TODO catch panics; not sure how to with gotk

type PanicWindow struct {
	*gtk.Window
}

func NewPanicWindow(err any) *PanicWindow {
	w := gtk.NewWindow()
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	title := gtk.NewLabel("Seabird has crashed")
	title.AddCSSClass("title-2")
	box.Append(title)
	w.SetChild(box)

	textView := gtk.NewTextView()
	textView.Buffer().Insert(textView.Buffer().EndIter(),
		strings.Join([]string{
			fmt.Sprintf("err: %v", err),
			fmt.Sprintf("version: %s", Version),
		}, "\n"))
	box.Append(textView)

	return &PanicWindow{w}
}
