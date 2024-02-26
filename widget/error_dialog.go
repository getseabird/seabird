package widget

import (
	"context"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/internal/ctxt"
)

func ShowErrorDialog(ctx context.Context, title string, err error) *adw.MessageDialog {
	dialog := adw.NewMessageDialog(ctxt.MustFrom[*gtk.Window](ctx), title, err.Error())
	dialog.AddResponse("Ok", "Ok")
	dialog.Show()
	return dialog
}
