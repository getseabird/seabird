package style

import (
	"embed"
	"os"
	"runtime"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

//go:embed *.css
var fs embed.FS

type Style string

const (
	Darwin  Style = "darwin"
	Windows Style = "windows"
	GNOME   Style = "gnome"
)

func Get() Style {
	if style := os.Getenv("SEABIRD_STYLE"); style != "" {
		return Style(style)
	}

	switch runtime.GOOS {
	case "darwin":
		return Darwin
	case "windows":
		return Windows
	default:
		return GNOME
	}
}

func Eq(styles ...Style) bool {
	s := Get()
	for _, style := range styles {
		if s == style {
			return true
		}
	}
	return false
}

func Load() {
	switch Get() {
	case Darwin:
		gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), getProvider("darwin.css"), gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
		dark := getProvider("darwin-dark.css")
		light := getProvider("darwin-light.css")
		addProviderWithColors(dark, light)
		gtk.SettingsGetDefault().NotifyProperty("gtk-application-prefer-dark-theme", func() {
			addProviderWithColors(dark, light)
		})
	case Windows:
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
	provider.LoadFromString(string(style))
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
