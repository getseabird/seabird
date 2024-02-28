package ui

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/behavior"
	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/getseabird/seabird/internal/util"
	"github.com/getseabird/seabird/widget"
	"github.com/google/uuid"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DetailView struct {
	*adw.NavigationPage
	ctx          context.Context
	behavior     *behavior.DetailBehavior
	prefPage     *adw.PreferencesPage
	groups       []*adw.PreferencesGroup
	sourceBuffer *gtksource.Buffer
	sourceView   *gtksource.View
	expanded     map[string]bool
}

func NewDetailView(ctx context.Context, behavior *behavior.DetailBehavior) *DetailView {
	content := gtk.NewBox(gtk.OrientationVertical, 0)
	content.AddCSSClass("view")
	d := DetailView{
		NavigationPage: adw.NewNavigationPage(content, "Object"),
		prefPage:       adw.NewPreferencesPage(),
		behavior:       behavior,
		expanded:       map[string]bool{},
		ctx:            ctx,
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

	stack := adw.NewViewStack()
	stack.AddTitledWithIcon(d.prefPage, "properties", "Properties", "info-outline-symbolic")
	stack.AddTitledWithIcon(d.createSource(), "source", "Yaml", "code-symbolic")
	content.Append(stack)

	switcher := adw.NewViewSwitcher()
	switcher.SetPolicy(adw.ViewSwitcherPolicyWide)
	switcher.SetStack(stack)
	header.SetTitleWidget(switcher)

	editable := gio.NewSimpleActionStateful("editable", nil, glib.NewVariantBoolean(false))
	save := gio.NewSimpleAction("save", nil)
	save.SetEnabled(false)
	editable.ConnectActivate(func(parameter *glib.Variant) {
		isEditable := !d.sourceView.Editable()
		d.sourceView.SetEditable(isEditable)
		editable.SetState(glib.NewVariantBoolean(isEditable))
		save.SetEnabled(isEditable)
	})
	save.ConnectActivate(func(parameter *glib.Variant) {
		text := d.sourceBuffer.Text(d.sourceBuffer.StartIter(), d.sourceBuffer.EndIter(), true)
		d.showSaveDialog(d.behavior.SelectedObject.Value(), d.behavior.Yaml.Value(), text)
	})

	// TODO should be local shortcuts, not global. how?

	ctxt.MustFrom[*gtk.Window](ctx).Application().SetAccelsForAction("detail.editable", []string{"<Ctrl>E"})
	ctxt.MustFrom[*gtk.Window](ctx).Application().SetAccelsForAction("detail.save", []string{"<Ctrl>S"})

	delete := gio.NewSimpleAction("delete", nil)
	delete.ConnectActivate(func(parameter *glib.Variant) {
		selected := d.behavior.SelectedObject.Value()
		dialog := adw.NewMessageDialog(ctxt.MustFrom[*gtk.Window](ctx), "Delete object?", selected.GetName())
		defer dialog.Show()
		dialog.AddResponse("cancel", "Cancel")
		dialog.AddResponse("delete", "Delete")
		dialog.SetResponseAppearance("delete", adw.ResponseDestructive)
		dialog.ConnectResponse(func(response string) {
			if response == "delete" {
				if err := behavior.DeleteObject(selected); err != nil {
					widget.ShowErrorDialog(ctx, "Failed to delete object", err)
				}
			}
		})
	})
	actionGroup := gio.NewSimpleActionGroup()
	actionGroup.AddAction(editable)
	actionGroup.AddAction(save)
	actionGroup.AddAction(delete)
	ctxt.MustFrom[*gtk.Window](ctx).InsertActionGroup("detail", actionGroup)

	onChange(ctx, d.behavior.SelectedObject, func(_ client.Object) {
		for d.Parent().(*adw.NavigationView).Pop() {
			// empty
		}
		if editable.State().Boolean() {
			editable.Activate(nil)
		}
	})
	onChange(ctx, d.behavior.Yaml, func(yaml string) {
		d.sourceBuffer.SetText(string(yaml))
	})
	onChange(ctx, d.behavior.Properties, d.onPropertiesChange)

	return &d
}

func (d *DetailView) onPropertiesChange(properties []api.Property) {
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
				prop.Widget(row, d.Parent().(*adw.NavigationView))
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
				row.SetSubtitleSelectable(false)
				row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
				row.ConnectActivated(func() {
					obj, err := prop.Reference.GetObject(d.ctx, d.behavior.Cluster)
					if err != nil {
						log.Print(err.Error())
						return
					}
					ctx, cancel := context.WithCancel(d.ctx)
					dv := NewDetailView(ctx, d.behavior.NewDetailBehavior(ctx))
					dv.behavior.SelectedObject.Update(obj)
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
			label.SetWrap(true)
			label.SetEllipsize(pango.EllipsizeEnd)
			box.Append(label)

			if prop.Widget != nil {
				prop.Widget(box, d.Parent().(*adw.NavigationView))
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
				prop.Widget(group, d.Parent().(*adw.NavigationView))
			}
			return group
		case 1:
			row := adw.NewExpanderRow()
			id := fmt.Sprintf("%s-%d-%d", util.ResourceGVR(d.behavior.SelectedResource.Value()).String(), level, index)
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
				prop.Widget(row, d.Parent().(*adw.NavigationView))
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
				prop.Widget(row, d.Parent().(*adw.NavigationView))
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

	windowSection := gio.NewMenu()
	windowSection.Append("Editable", "detail.editable")
	windowSection.Append("Save", "detail.save")
	d.sourceView.SetExtraMenu(windowSection)

	return scrolledWindow
}

func (d *DetailView) setSourceColorScheme() {
	util.SetSourceColorScheme(d.sourceBuffer)
}

func (d *DetailView) showSaveDialog(object client.Object, current, next string) *adw.MessageDialog {
	json, err := util.YamlToJson([]byte(next))
	if err != nil {
		return widget.ShowErrorDialog(d.ctx, "Error decoding object", err)
	}
	var obj unstructured.Unstructured
	_, _, err = unstructured.UnstructuredJSONScheme.Decode(json, nil, &obj)
	if err != nil {
		return widget.ShowErrorDialog(d.ctx, "Error decoding object", err)
	}

	dialog := adw.NewMessageDialog(ctxt.MustFrom[*gtk.Window](d.ctx), fmt.Sprintf("Saving %s", object.GetName()), "The following changes will be made")
	dialog.AddResponse("cancel", "Cancel")
	dialog.AddResponse("save", "Save")
	dialog.SetResponseAppearance("save", adw.ResponseSuggested)
	dialog.SetSizeRequest(600, 500)
	defer dialog.Show()

	box := dialog.Child().(*gtk.WindowHandle).Child().(*gtk.Box).FirstChild().(*gtk.Box)

	box.FirstChild().(*gtk.Label).NextSibling().(*gtk.Label).SetVExpand(false)

	edits := myers.ComputeEdits(span.URIFromPath(object.GetName()), current, next)

	buf := gtksource.NewBufferWithLanguage(gtksource.LanguageManagerGetDefault().Language("diff"))
	buf.SetText(strings.TrimPrefix(fmt.Sprint(gotextdiff.ToUnified("", "", current, edits)), "--- \n+++ \n"))
	util.SetSourceColorScheme(buf)
	view := gtksource.NewViewWithBuffer(buf)
	view.SetEditable(false)
	view.SetWrapMode(gtk.WrapWord)
	view.SetShowLineNumbers(false)
	view.SetMonospace(true)

	sw := gtk.NewScrolledWindow()
	sw.SetChild(view)
	sw.SetVExpand(true)

	box.Append(sw)

	dialog.ConnectResponse(func(response string) {
		if response == "save" {
			if err := d.behavior.UpdateObject(&obj); err != nil {
				widget.ShowErrorDialog(d.ctx, "Error updating object", err)
			}
		}
	})

	return dialog
}
