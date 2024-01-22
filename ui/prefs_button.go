package ui

import (
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type PrefsButton struct {
	*gtk.MenuButton
}

func NewPrefsButton() *PrefsButton {
	b := PrefsButton{gtk.NewMenuButton()}
	b.SetIconName("open-menu-symbolic")
	menu := gio.NewMenu()
	menu.Append("New Window", "app.new")
	menu.Append("Disconnect", "app.new")
	menu.Append("Preferences", "app.preferences")
	menu.Append("Keyboard Shortcuts", "app.shortcuts")
	menu.Append("About", "app.about")
	popover := gtk.NewPopoverMenuFromModel(menu)
	b.SetPopover(popover)

	return &b
}
