package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

var Version string = "dev"

type AboutWindow struct {
	*adw.AboutWindow
}

func NewAboutWindow(parent *gtk.Window) *AboutWindow {
	w := AboutWindow{adw.NewAboutWindow()}
	w.SetApplicationIcon("seabird")
	w.SetApplicationName(ApplicationName)
	w.SetVersion(Version)
	w.SetTransientFor(parent)
	w.SetWebsite("https://github.com/getseabird/seabird")
	w.SetIssueURL("https://github.com/getseabird/seabird/issues")
	w.SetSupportURL("https://github.com/getseabird/seabird/discussions")
	w.SetLicenseType(gtk.LicenseMPL20)
	return &w
}
