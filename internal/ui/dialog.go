package ui

import (
	"context"
	"errors"
	"os/exec"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/getseabird/seabird/widget"
)

func showClusterPrefsErrorDialog(ctx context.Context, prefs api.ClusterPreferences) bool {
	if len(prefs.Name) == 0 {
		widget.ShowErrorDialog(ctx, "Error", errors.New("name is required"))
		return true
	}

	if ex := prefs.Exec; ex != nil {
		if _, err := exec.LookPath(ex.Command); err != nil {
			w, _ := ctxt.From[*gtk.Window](ctx)
			dialog := adw.NewMessageDialog(w, "Credential plugin not found", err.Error())
			dialog.AddResponse("cancel", "Cancel")
			dialog.AddResponse("docs", "Open documentation")
			dialog.SetResponseAppearance("docs", adw.ResponseSuggested)
			dialog.ConnectResponse(func(response string) {
				switch response {
				case "docs":
					gtk.ShowURI(w, "https://getseabird.github.io/docs/credential-plugins/", gdk.CURRENT_TIME)
				}
			})
			dialog.Show()
			return true
		}
	}

	return false
}
