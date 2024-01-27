package widget

import (
	"runtime"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// Adwaita makes client side decorations mandatory, which causes some problems on the Windows platform
// See e.g. https://gitlab.gnome.org/GNOME/gtk/-/issues/3749
// This wrapper uses GtkWindow on Windows and AdwWindow everywhere else
type UniversalApplicationWindow struct {
	*gtk.ApplicationWindow
	AdwWindow *adw.ApplicationWindow
}

func NewUniversalApplicationWindow(app *gtk.Application) *UniversalApplicationWindow {
	switch runtime.GOOS {
	case "windows":
		w := gtk.NewApplicationWindow(app)
		w.SetDecorated(true)
		return &UniversalApplicationWindow{ApplicationWindow: w}
	default:
		w := adw.NewApplicationWindow(app)
		// causes odd visual glitches around corners
		// w.Window.SetDecorated(false)
		return &UniversalApplicationWindow{ApplicationWindow: &w.ApplicationWindow, AdwWindow: w}
	}
}

func (w *UniversalApplicationWindow) SetContent(content gtk.Widgetter) {
	if w.AdwWindow != nil {
		w.AdwWindow.SetContent(content)
	} else {
		w.SetChild(content)
	}
}
