package util

import (
	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func SetSourceColorScheme(buf *gtksource.Buffer) {
	if gtk.SettingsGetDefault().ObjectProperty("gtk-application-prefer-dark-theme").(bool) {
		buf.SetStyleScheme(gtksource.StyleSchemeManagerGetDefault().Scheme("Adwaita-dark"))
	} else {
		buf.SetStyleScheme(gtksource.StyleSchemeManagerGetDefault().Scheme("Adwaita"))
	}
}
