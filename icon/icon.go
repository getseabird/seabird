package icon

import (
	_ "embed"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
)

//go:generate glib-compile-resources --target=icon.gresource gresource.xml

//go:embed icon.gresource
var data []byte

func Register() error {
	res, err := gio.NewResourceFromData(glib.NewBytesWithGo(data))
	if err != nil {
		return err
	}
	gio.ResourcesRegister(res)

	return nil
}
