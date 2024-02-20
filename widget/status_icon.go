package widget

import "github.com/diamondburned/gotk4/pkg/gtk/v4"

func NewStatusIcon(ok bool) *gtk.Image {
	if ok {
		icon := gtk.NewImageFromIconName("emblem-ok-symbolic")
		icon.AddCSSClass("success")
		icon.SetHAlign(gtk.AlignStart)
		return icon
	}
	icon := gtk.NewImageFromIconName("dialog-warning")
	icon.AddCSSClass("warning")
	icon.SetHAlign(gtk.AlignStart)
	return icon
}
