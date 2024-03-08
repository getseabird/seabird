package ui

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sort"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/ui/editor"
	"github.com/getseabird/seabird/internal/util"
	"github.com/getseabird/seabird/widget"
	"github.com/google/uuid"
	"github.com/imkira/go-observer/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DetailView struct {
	*adw.NavigationPage
	*common.ClusterState
	ctx          context.Context
	prefPage     *adw.PreferencesPage
	groups       []*adw.PreferencesGroup
	sourceBuffer *gtksource.Buffer
	sourceView   *gtksource.View
	expanded     map[string]bool
	editor       *editor.EditorWindow
}

func NewDetailView(ctx context.Context, state *common.ClusterState, editor *editor.EditorWindow) *DetailView {
	content := gtk.NewBox(gtk.OrientationVertical, 0)
	content.AddCSSClass("background")
	d := DetailView{
		NavigationPage: adw.NewNavigationPage(content, "Object"),
		ClusterState:   state,
		prefPage:       adw.NewPreferencesPage(),
		expanded:       map[string]bool{},
		ctx:            ctx,
		editor:         editor,
	}
	d.SetTag(uuid.NewString())

	clamp := d.prefPage.FirstChild().(*gtk.ScrolledWindow).FirstChild().(*gtk.Viewport).FirstChild().(*adw.Clamp)
	clamp.SetMaximumSize(5000)

	header := adw.NewHeaderBar()
	header.AddCSSClass("flat")
	content.Append(header)
	switch runtime.GOOS {
	case "windows", "darwin":
		header.SetShowStartTitleButtons(false)
		header.SetShowEndTitleButtons(false)
	}

	delete := gtk.NewButton()
	delete.SetIconName("edit-delete-symbolic")
	delete.SetTooltipText("Delete")
	delete.ConnectClicked(func() {
		selected := d.SelectedObject.Value()
		dialog := adw.NewMessageDialog(ctxt.MustFrom[*gtk.Window](ctx), "Delete object?", selected.GetName())
		defer dialog.Show()
		dialog.AddResponse("cancel", "Cancel")
		dialog.AddResponse("delete", "Delete")
		dialog.SetResponseAppearance("delete", adw.ResponseDestructive)
		dialog.ConnectResponse(func(response string) {
			switch response {
			case "delete":
				if err := d.Delete(ctx, selected); err != nil {
					widget.ShowErrorDialog(ctx, "Failed to delete object", err)
				}
			}
		})
	})
	header.PackEnd(delete)

	edit := gtk.NewButton()
	edit.SetIconName("document-edit-symbolic")
	edit.SetTooltipText("Edit")
	edit.ConnectClicked(func() {
		gvk := util.ResourceGVK(d.SelectedResource.Value())
		if err := d.editor.AddPage(&gvk, d.SelectedObject.Value()); err != nil {
			widget.ShowErrorDialog(d.ctx, "Error loading editor", err)
		} else {
			d.editor.Present()
		}
	})
	header.PackEnd(edit)

	stack := adw.NewViewStack()
	stack.AddTitledWithIcon(d.prefPage, "properties", "Properties", "info-outline-symbolic")
	stack.AddTitledWithIcon(d.createSource(), "source", "Yaml", "code-symbolic")
	content.Append(stack)

	switcher := adw.NewViewSwitcher()
	switcher.SetPolicy(adw.ViewSwitcherPolicyWide)
	switcher.SetStack(stack)
	header.SetTitleWidget(switcher)

	common.OnChange(ctx, d.ClusterPreferences, func(prefs api.ClusterPreferences) {
		edit.SetVisible(!prefs.ReadOnly)
		delete.SetVisible(!prefs.ReadOnly)
	})
	common.OnChange(ctx, d.SelectedObject, func(object client.Object) {
		switch parent := d.Parent().(type) {
		case *adw.NavigationView:
			for parent.Pop() {
				// empty
			}
		}

		if object == nil {
			d.sourceBuffer.SetText("")
			d.updateProperties([]api.Property{})
			return
		}

		yaml, err := d.Encoder.EncodeYAML(object)
		if err != nil {
			d.sourceBuffer.SetText(fmt.Sprintf("error: %v", err))
		} else {
			d.sourceBuffer.SetText(string(yaml))
		}

		var props []api.Property
		for _, ext := range d.Extensions {
			props = ext.CreateObjectProperties(ctx, object, props)
		}
		sort.Slice(props, func(i, j int) bool {
			return props[i].GetPriority() > props[j].GetPriority()
		})
		d.updateProperties(props)
	})

	return &d
}

func (d *DetailView) updateProperties(properties []api.Property) {
	for _, g := range d.groups {
		d.prefPage.Remove(g)
	}
	d.groups = nil

	for i, prop := range properties {
		group := d.renderObjectProperty(0, i, prop).(*adw.PreferencesGroup)
		d.groups = append(d.groups, group)
		d.prefPage.Add(group)
	}
}

func (d *DetailView) renderObjectProperty(level, index int, prop api.Property) gtk.Widgetter {
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
				if parent, ok := d.Parent().(*adw.NavigationView); ok {
					prop.Widget(row, parent)
				}
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
					obj, err := prop.Reference.GetObject(d.ctx, d.Cluster)
					if err != nil {
						log.Print(err.Error())
						return
					}
					ctx, cancel := context.WithCancel(d.ctx)
					state := *d.ClusterState
					state.SelectedObject = observer.NewProperty[client.Object](obj)
					dv := NewDetailView(ctx, &state, d.editor)
					d.Parent().(*adw.NavigationView).Push(dv.NavigationPage)
					d.Parent().(*adw.NavigationView).ConnectPopped(func(page *adw.NavigationPage) {
						if page.Tag() == dv.Tag() {
							cancel()
						}
					})
				})
			}
			return row
		case 3:
			box := gtk.NewBox(gtk.OrientationHorizontal, 4)
			box.SetHAlign(gtk.AlignStart)

			label := gtk.NewLabel(prop.Name)
			label.AddCSSClass("caption")
			label.AddCSSClass("dim-label")
			label.SetVAlign(gtk.AlignStart)
			label.SetEllipsize(pango.EllipsizeEnd)
			box.Append(label)

			label = gtk.NewLabel(prop.Value)
			label.AddCSSClass("caption")
			label.SetName("button")
			label.AddCSSClass("pill")
			label.SetWrap(true)
			label.SetEllipsize(pango.EllipsizeEnd)
			box.Append(label)

			if prop.Widget != nil {
				if parent, ok := d.Parent().(*adw.NavigationView); ok {
					prop.Widget(box, parent)
				}
			}
			return box
		}

	case *api.GroupProperty:
		switch level {
		case 0:
			group := adw.NewPreferencesGroup()
			group.SetTitle(prop.Name)
			for i, child := range prop.Children {
				group.Add(d.renderObjectProperty(level+1, i, child))
			}
			if prop.Widget != nil {
				if parent, ok := d.Parent().(*adw.NavigationView); ok {
					prop.Widget(group, parent)
				}
			}
			return group
		case 1:
			row := adw.NewExpanderRow()
			id := fmt.Sprintf("%s-%d-%d", util.ResourceGVR(d.SelectedResource.Value()).String(), level, index)
			if e, ok := d.expanded[id]; ok && e {
				row.SetExpanded(true)
			}
			row.Connect("state-flags-changed", func() {
				d.expanded[id] = row.Expanded()
			})
			row.SetTitle(prop.Name)
			for i, child := range prop.Children {
				row.AddRow(d.renderObjectProperty(level+1, i, child))
			}
			row.SetSensitive(len(prop.Children) > 0)
			if prop.Widget != nil {
				if parent, ok := d.Parent().(*adw.NavigationView); ok {
					prop.Widget(row, parent)
				}
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
			for i, child := range prop.Children {
				box.Insert(d.renderObjectProperty(level+1, i, child), -1)
				// prop.Value += fmt.Sprintf("%s: %s\n", child.Name, child.Value)
			}
			if prop.Widget != nil {
				if parent, ok := d.Parent().(*adw.NavigationView); ok {
					prop.Widget(row, parent)
				}
			}
			return row
		}
	}

	return nil
}

func (d *DetailView) createSource() *gtk.ScrolledWindow {
	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetVExpand(true)

	d.sourceBuffer = gtksource.NewBufferWithLanguage(gtksource.LanguageManagerGetDefault().Language("yaml"))
	d.setSourceColorScheme()
	gtk.SettingsGetDefault().NotifyProperty("gtk-application-prefer-dark-theme", d.setSourceColorScheme)
	d.sourceView = gtksource.NewViewWithBuffer(d.sourceBuffer)
	d.sourceView.SetEditable(false)
	d.sourceView.SetWrapMode(gtk.WrapWord)
	d.sourceView.SetShowLineNumbers(true)
	d.sourceView.SetMonospace(true)
	scrolledWindow.SetChild(d.sourceView)

	return scrolledWindow
}

func (d *DetailView) setSourceColorScheme() {
	util.SetSourceColorScheme(d.sourceBuffer)
}
