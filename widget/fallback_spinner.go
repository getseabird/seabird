package widget

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type FallbackSpinner struct {
	*adw.Bin
	spinner  *gtk.Spinner
	fallback gtk.Widgetter
}

func NewFallbackSpinner(fallback gtk.Widgetter) *FallbackSpinner {
	f := &FallbackSpinner{
		Bin:      adw.NewBin(),
		spinner:  gtk.NewSpinner(),
		fallback: fallback,
	}
	f.SetChild(fallback)
	return f
}

func (f *FallbackSpinner) Start() {
	f.SetChild(f.spinner)
	f.spinner.Start()
}

func (f *FallbackSpinner) Stop() {
	f.SetChild(f.fallback)
	f.spinner.Stop()
}
