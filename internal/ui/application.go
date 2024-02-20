package ui

import (
	"log"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/behavior"
	"github.com/getseabird/seabird/internal/icon"
	"github.com/getseabird/seabird/internal/style"
)

const ApplicationName = "Seabird"

type Application struct {
	*adw.Application
	version string
}

func NewApplication(version string) (*Application, error) {
	switch runtime.GOOS {
	case "windows":
		os.Setenv("GTK_CSD", "0")
	}

	gtk.Init()

	if err := icon.Register(); err != nil {
		log.Printf("failed to load icons: %v", err)
	}

	b, err := behavior.NewBehavior()
	if err != nil {
		return nil, err
	}

	adw.StyleManagerGetDefault().SetColorScheme(b.Preferences.Value().ColorScheme)
	onChange(b.Preferences, func(p api.Preferences) {
		adw.StyleManagerGetDefault().SetColorScheme(adw.ColorScheme(p.ColorScheme))
	})

	a := Application{
		Application: adw.NewApplication("dev.skynomads.Seabird", gio.ApplicationFlagsNone),
		version:     version,
	}

	provider := gtk.NewCSSProvider()
	theme, _ := style.FS.ReadFile("theme.css")
	provider.LoadFromData(string(theme))
	gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	a.ConnectActivate(func() {
		NewWelcomeWindow(&a.Application.Application, b).Show()
	})

	return &a, nil
}

func (a *Application) Run() {
	debug.SetPanicOnFault(true)
	defer func() {
		if err := recover(); err != nil {
			NewPanicWindow(err).Present()
		}
	}()

	if code := a.Application.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
