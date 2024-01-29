package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func ShowErrorDialog(parent *gtk.Window, title string, err error) *adw.MessageDialog {
	dialog := adw.NewMessageDialog(parent, title, err.Error())
	dialog.AddResponse("Ok", "Ok")
	dialog.Show()
	return dialog
}
