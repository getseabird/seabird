package style

import (
	"embed"
	"runtime"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

//go:embed *.css
var fs embed.FS

func Load() {
	switch runtime.GOOS {
	case "darwin":
		gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), getProvider("darwin.css"), gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
		dark := getProvider("darwin-dark.css")
		light := getProvider("darwin-light.css")
		addProviderWithColors(dark, light)
		gtk.SettingsGetDefault().NotifyProperty("gtk-application-prefer-dark-theme", func() {
			addProviderWithColors(dark, light)
		})
	case "windows":
		dark := getProvider("windows-dark.css")
		light := getProvider("windows-light.css")
		addProviderWithColors(dark, light)
		gtk.SettingsGetDefault().NotifyProperty("gtk-application-prefer-dark-theme", func() {
			addProviderWithColors(dark, light)
		})
	}

	gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), getProvider("style.css"), gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
}

func getProvider(name string) *gtk.CSSProvider {
	provider := gtk.NewCSSProvider()
	style, _ := fs.ReadFile(name)
	provider.LoadFromData(string(style))
	return provider
}

func addProviderWithColors(dark, light *gtk.CSSProvider) {
	if gtk.SettingsGetDefault().ObjectProperty("gtk-application-prefer-dark-theme").(bool) {
		gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), dark, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
		gtk.StyleContextRemoveProviderForDisplay(gdk.DisplayGetDefault(), light)
	} else {
		gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), light, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
		gtk.StyleContextRemoveProviderForDisplay(gdk.DisplayGetDefault(), dark)

	}
}
