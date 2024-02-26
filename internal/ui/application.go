package ui

import (
	"context"
	"log"
	"os"
	"runtime"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
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
	gtk.Init()

	switch runtime.GOOS {
	case "windows":
		os.Setenv("GTK_CSD", "0")
	case "darwin":
		gtk.SettingsGetDefault().SetObjectProperty("gtk-decoration-layout", "close,minimize,maximize")
	}

	if err := icon.Register(); err != nil {
		log.Printf("failed to load icons: %v", err)
	}

	ctx := context.Background()

	b, err := behavior.NewBehavior()
	if err != nil {
		return nil, err
	}

	adw.StyleManagerGetDefault().SetColorScheme(b.Preferences.Value().ColorScheme)
	onChange(ctx, b.Preferences, func(p api.Preferences) {
		adw.StyleManagerGetDefault().SetColorScheme(adw.ColorScheme(p.ColorScheme))
	})

	style.Load()

	a := Application{
		Application: adw.NewApplication("dev.skynomads.Seabird", gio.ApplicationFlagsNone),
		version:     version,
	}

	a.ConnectActivate(func() {
		NewWelcomeWindow(ctx, &a.Application.Application, b).Show()
	})

	return &a, nil
}

func (a *Application) Run() {
	// TODO doesn't work
	// debug.SetPanicOnFault(true)
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		NewPanicWindow(err).Present()
	// 	}
	// }()

	if code := a.Application.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
