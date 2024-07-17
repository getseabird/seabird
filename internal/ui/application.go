package ui

import (
	"context"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/icon"
	"github.com/getseabird/seabird/internal/style"
	"github.com/getseabird/seabird/internal/ui/common"
	"k8s.io/klog/v2"
)

const ApplicationName = "Seabird"

type Application struct {
	*adw.Application
	version string
}

func NewApplication(version string) (*Application, error) {
	gtk.Init()

	switch style.Get() {
	case style.Darwin:
		gtk.SettingsGetDefault().SetObjectProperty("gtk-decoration-layout", "close,minimize,maximize")
	}

	if err := icon.Register(); err != nil {
		klog.Infof("failed to load icons: %v", err)
	}

	ctx := context.Background()

	state, err := common.NewState()
	if err != nil {
		return nil, err
	}

	common.OnChange(ctx, state.Preferences, func(p api.Preferences) {
		adw.StyleManagerGetDefault().SetColorScheme(adw.ColorScheme(p.ColorScheme))
	})

	style.Load()

	a := Application{
		Application: adw.NewApplication("dev.skynomads.Seabird", gio.ApplicationFlagsNone),
		version:     version,
	}

	a.ConnectActivate(func() {
		NewWelcomeWindow(ctx, &a.Application.Application, state).Present()
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
