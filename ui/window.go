package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func Window(app *gtk.Application) *gtk.ApplicationWindow {
	window := gtk.NewApplicationWindow(app)
	window.SetTitle("gotk4 Example")
	window.SetDefaultSize(1000, 800)

	header := adw.NewHeaderBar()
	leftBox := gtk.NewBox(gtk.OrientationHorizontal, 4)
	rightBox := gtk.NewBox(gtk.OrientationHorizontal, 4)
	header.PackStart(leftBox)
	header.PackEnd(rightBox)
	window.SetTitlebar(header)

	cb := gtk.NewComboBoxText()
	cb.AppendText("k3s")
	cb.SetActive(0)
	addButton := gtk.NewButtonFromIconName("document-new")
	leftBox.Append(cb)
	leftBox.Append(addButton)

	searchButton := gtk.NewButton()
	searchButton.SetIconName("system-search-symbolic")
	rightBox.Append(searchButton)

	return window
}
