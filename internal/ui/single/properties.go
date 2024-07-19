package single

import (
	"context"
	"sort"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/pubsub"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type propertiesView struct {
	*api.Cluster
}

func (p *propertiesView) Render(ctx context.Context, level int, prop api.Property, single *SingleView) gtk.Widgetter {
	switch prop := prop.(type) {
	case *api.TextProperty:
		switch level {
		case 0, 1, 2:
			row := adw.NewActionRow()
			row.SetTitle(prop.Name)
			row.SetUseMarkup(false)
			row.AddCSSClass("property")
			// *Very* long labels cause a segfault in GTK. Limiting lines prevents it, but they're still
			// slow and CPU-intensive to render. https://gitlab.gnome.org/GNOME/gtk/-/issues/1332
			// TODO explore alternative rendering options such as TextView
			row.SetSubtitleLines(5)
			row.SetSubtitle(prop.Value)

			if prop.Widget != nil {
				prop.Widget(row, single.navView)
			}
			if prop.Reference == nil {
				copy := gtk.NewButton()
				copy.SetIconName("edit-copy-symbolic")
				copy.AddCSSClass("flat")
				copy.AddCSSClass("dim-label")
				copy.SetVAlign(gtk.AlignCenter)
				copy.ConnectClicked(func() {
					gdk.DisplayGetDefault().Clipboard().SetText(prop.Value)
				})
				row.AddSuffix(copy)
			} else {
				row.SetActivatable(true)
				row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
				row.ConnectActivated(func() {
					obj, err := p.GetReference(ctx, *prop.Reference)
					if err != nil {
						klog.Infof("error resolving reference '%v': %v", prop.Reference, err.Error())
						return
					}
					ctx, cancel := context.WithCancel(ctx)
					state := *single.ClusterState
					state.SelectedObject = pubsub.NewProperty[client.Object](obj)
					sv := NewSingleView(ctx, &state, single.editor, single.navView)
					sv.PinAdded.Sub(ctx, single.PinAdded.Pub)
					sv.PinRemoved.Sub(ctx, single.PinRemoved.Pub)
					sv.Deleted.Sub(ctx, func(o client.Object) {
						if visible := single.navView.VisiblePage(); visible != nil && visible.Tag() == sv.Tag() {
							single.navView.Pop()
						}
						single.navView.Remove(sv.NavigationPage)
					})
					single.navView.Push(sv.NavigationPage)
					single.navView.ConnectReplaced(cancel)
				})
			}
			return row
		case 3:
			box := gtk.NewBox(gtk.OrientationHorizontal, 4)
			box.SetHAlign(gtk.AlignStart)

			label := gtk.NewLabel(prop.Name)
			label.AddCSSClass("dim-label")
			label.AddCSSClass("monospace")
			label.SetEllipsize(pango.EllipsizeEnd)
			box.Append(label)

			label = gtk.NewLabel(prop.Value)
			label.SetWrap(true)
			label.AddCSSClass("monospace")
			label.SetEllipsize(pango.EllipsizeEnd)
			box.Append(label)

			if prop.Widget != nil {
				prop.Widget(box, single.navView)
			}
			return box
		}

	case *api.GroupProperty:
		sort.Slice(prop.Children, func(i, j int) bool {
			return prop.Children[i].GetPriority() > prop.Children[j].GetPriority()
		})
		switch level {
		case 0:
			group := adw.NewPreferencesGroup()
			group.SetTitle(prop.Name)
			for _, child := range prop.Children {
				group.Add(p.Render(ctx, level+1, child, single))
			}
			if prop.Widget != nil {
				prop.Widget(group, single.navView)
			}
			return group
		case 1:
			row := adw.NewExpanderRow()
			row.SetTitle(prop.Name)
			for _, child := range prop.Children {
				row.AddRow(p.Render(ctx, level+1, child, single))
			}
			row.SetSensitive(len(prop.Children) > 0)
			if prop.Widget != nil {
				prop.Widget(row, single.navView)
			}
			return row
		case 2:
			row := adw.NewActionRow()
			row.SetTitle(prop.Name)
			row.SetUseMarkup(false)
			row.AddCSSClass("property")

			box := gtk.NewFlowBox()
			box.SetColumnSpacing(8)
			box.SetSelectionMode(gtk.SelectionNone)
			row.FirstChild().(*gtk.Box).FirstChild().(*gtk.Box).NextSibling().(*gtk.Image).NextSibling().(*gtk.Box).Append(box)
			for _, child := range prop.Children {
				box.Insert(p.Render(ctx, level+1, child, single), -1)
			}
			if prop.Widget != nil {
				prop.Widget(row, single.navView)
			}
			return row
		}
	}
	return nil

}
