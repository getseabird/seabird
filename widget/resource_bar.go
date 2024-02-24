package widget

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"k8s.io/apimachinery/pkg/api/resource"
)

func NewResourceBar(res *resource.Quantity, req *resource.Quantity, iconName string) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationVertical, 4)
	box.SetVAlign(gtk.AlignCenter)

	levelBar := gtk.NewLevelBar()
	levelBar.SetSizeRequest(50, -1)
	levelBar.SetHAlign(gtk.AlignCenter)
	levelBar.SetVAlign(gtk.AlignCenter)
	// down from offset, not up
	levelBar.RemoveOffsetValue(gtk.LEVEL_BAR_OFFSET_LOW)
	levelBar.RemoveOffsetValue(gtk.LEVEL_BAR_OFFSET_HIGH)
	levelBar.AddOffsetValue("lb-normal", .85)
	levelBar.AddOffsetValue("lb-warning", .95)
	levelBar.AddOffsetValue("lb-error", 1)

	if res != nil && req != nil && !req.IsZero() {
		percent := res.AsApproximateFloat64() / req.AsApproximateFloat64()
		levelBar.SetValue(min(percent, 1))
		box.SetTooltipText(fmt.Sprintf("%.0f%%", percent*100))
	}

	if iconName != "" {
		box.Append(gtk.NewImageFromIconName(iconName))
	}
	box.Append(levelBar)

	return box
}
