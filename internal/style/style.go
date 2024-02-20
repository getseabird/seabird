package style

import (
	"embed"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

//go:embed *.css
var fs embed.FS

func Load() {
	provider := gtk.NewCSSProvider()
	style, _ := fs.ReadFile("style.css")
	provider.LoadFromData(string(style))
	gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
}
