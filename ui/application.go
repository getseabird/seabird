package ui

import (
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/jgillich/kubegio/state"
	"github.com/jgillich/kubegio/style"
)

const ApplicationName = "kubegtk"

type Application struct {
	*adw.Application
	version string
}

func NewApplication(version string) (*Application, error) {
	gtk.Init()

	prefs, err := state.LoadPreferences()
	if err != nil {
		return nil, err
	}
	prefs.Defaults()

	adw.StyleManagerGetDefault().SetColorScheme(adw.ColorScheme(prefs.ColorScheme))

	a := Application{
		Application: adw.NewApplication("io.github.jgillich.kubegtk", gio.ApplicationFlagsNone),
		version:     version,
	}

	provider := gtk.NewCSSProvider()
	theme, _ := style.FS.ReadFile("theme.css")
	provider.LoadFromData(string(theme))
	gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	a.ConnectActivate(func() {
		NewWelcomeWindow(&a.Application.Application, prefs).Show()
	})

	return &a, nil
}

func (a *Application) Run() {
	if code := a.Application.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
