package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type AboutWindow struct {
	*adw.AboutWindow
}

func NewAboutWindow(parent *gtk.Window) *AboutWindow {
	w := AboutWindow{adw.NewAboutWindow()}
	w.SetApplicationName(ApplicationName)
	w.SetVersion("0.0.1")
	w.SetTransientFor(parent)
	w.SetWebsite("https://github.com/jgillich/kubegtk")
	w.SetIssueURL("https://github.com/jgillich/kubegtk/issues")
	w.SetLicenseType(gtk.LicenseMPL20)
	return &w
}
