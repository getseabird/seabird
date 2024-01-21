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
	w.SetDeveloperName("Jakob Gillich")
	w.SetVersion("0.1-dev")
	w.SetTransientFor(parent)
	w.SetWebsite("https://github.com/jgillich/kubegtk")
	w.SetIssueURL("https://github.com/jgillich/kubegtk/issues")
	w.SetLicenseType(gtk.LicenseGPL30)
	return &w
}
