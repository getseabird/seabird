package widget

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/internal/style"
)

// Adwaita makes client side decorations mandatory, which causes some problems on the Windows platform
// See e.g. https://gitlab.gnome.org/GNOME/gtk/-/issues/3749
// This wrapper uses GtkWindow on Windows and AdwWindow everywhere else
type UniversalWindow struct {
	*gtk.Window
	AdwWindow *adw.Window
}

func NewUniversalWindow() *UniversalWindow {
	switch style.Get() {
	case style.Windows:
		w := gtk.NewWindow()
		w.SetDecorated(true)
		return &UniversalWindow{Window: w}
	default:
		w := adw.NewWindow()
		return &UniversalWindow{Window: &w.Window, AdwWindow: w}
	}
}

func (w *UniversalWindow) SetContent(content gtk.Widgetter) {
	if w.AdwWindow != nil {
		w.AdwWindow.SetContent(content)
	} else {
		w.SetChild(content)
	}
}
