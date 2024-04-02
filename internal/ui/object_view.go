package ui

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/getseabird/seabird/internal/style"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/ui/editor"
	"github.com/getseabird/seabird/internal/util"
	"github.com/getseabird/seabird/widget"
	"github.com/google/uuid"
	"github.com/imkira/go-observer/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectView struct {
	*adw.NavigationPage
	*common.ClusterState
	ctx          context.Context
	prefPage     *adw.PreferencesPage
	groups       []*adw.PreferencesGroup
	sourceBuffer *gtksource.Buffer
	sourceView   *gtksource.View
	expanded     map[string]bool
	editor       *editor.EditorWindow
	navView      *adw.NavigationView
	navigation   *Navigation
}

func NewObjectView(ctx context.Context, state *common.ClusterState, editor *editor.EditorWindow, navView *adw.NavigationView, navigation *Navigation) *ObjectView {
	content := gtk.NewBox(gtk.OrientationVertical, 0)
	content.AddCSSClass("view")
	o := ObjectView{
		NavigationPage: adw.NewNavigationPage(content, "Object"),
		ClusterState:   state,
		prefPage:       adw.NewPreferencesPage(),
		expanded:       map[string]bool{},
		ctx:            ctx,
		editor:         editor,
		navView:        navView,
		navigation:     navigation,
	}
	o.SetTag(uuid.NewString())

	clamp := o.prefPage.FirstChild().(*gtk.ScrolledWindow).FirstChild().(*gtk.Viewport).FirstChild().(*adw.Clamp)
	clamp.SetMaximumSize(5000)

	header := adw.NewHeaderBar()
	header.AddCSSClass("flat")
	content.Append(header)
	header.SetShowStartTitleButtons(!style.Eq(style.Windows, style.Darwin))
	header.SetShowEndTitleButtons(!style.Eq(style.Windows, style.Darwin))

	delete := gtk.NewButton()
	delete.SetIconName("user-trash-symbolic")
	delete.SetTooltipText("Delete")
	delete.ConnectClicked(func() {
		selected := o.SelectedObject.Value()
		dialog := adw.NewMessageDialog(ctxt.MustFrom[*gtk.Window](ctx), "Delete object?", selected.GetName())
		defer dialog.Show()
		dialog.AddResponse("cancel", "Cancel")
		dialog.AddResponse("delete", "Delete")
		dialog.SetResponseAppearance("delete", adw.ResponseDestructive)
		dialog.ConnectResponse(func(response string) {
			switch response {
			case "delete":
				if err := o.Delete(ctx, selected); err != nil {
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
		gvk := o.SelectedObject.Value().GetObjectKind().GroupVersionKind()
		if err := o.editor.AddPage(&gvk, o.SelectedObject.Value()); err != nil {
			widget.ShowErrorDialog(o.ctx, "Error loading editor", err)
		} else {
			o.editor.Present()
		}
	})
	header.PackEnd(edit)

	pin := gtk.NewToggleButton()
	pin.SetIconName("view-pin-symbolic")
	pin.SetTooltipText("Pin")
	pin.ConnectClicked(func() {
		if pin.Active() {
			o.navigation.AddPin(o.SelectedObject.Value())
		} else {
			o.navigation.RemovePin(o.SelectedObject.Value())
		}
	})
	header.PackStart(pin)

	stack := adw.NewViewStack()
	stack.AddTitledWithIcon(o.prefPage, "properties", "Properties", "info-outline-symbolic")
	stack.AddTitledWithIcon(o.createSource(), "source", "Yaml", "code-symbolic")
	content.Append(stack)

	switcher := adw.NewViewSwitcher()
	switcher.SetPolicy(adw.ViewSwitcherPolicyWide)
	switcher.SetStack(stack)
	header.SetTitleWidget(switcher)

	common.OnChange(ctx, o.ClusterPreferences, func(prefs api.ClusterPreferences) {
		edit.SetVisible(!prefs.ReadOnly)
		delete.SetVisible(!prefs.ReadOnly)

		if object := o.SelectedObject.Value(); object != nil {
			pinned := false
			for _, p := range prefs.Navigation.Pins {
				if p.UID == object.GetUID() {
					pinned = true
					break
				}
			}
			pin.SetActive(pinned)
		}
	})

	watchCtx, cancelWatch := context.WithCancel(ctx)
	common.OnChange(ctx, o.SelectedObject, func(object client.Object) {
		for o.navView.Pop() {
			// empty
		}

		if object == nil {
			o.sourceBuffer.SetText("")
			o.updateProperties([]api.Property{})
			return
		}

		cancelWatch()
		watchCtx, cancelWatch = context.WithCancel(ctx)
		api.Watch(
			watchCtx,
			o.Cluster,
			o.Cluster.GetAPIResource(object.GetObjectKind().GroupVersionKind()),
			api.WatchOptions[client.Object]{
				ListOptions: v1.ListOptions{
					FieldSelector:   fields.OneTermEqualSelector("metadata.name", object.GetName()).String(),
					ResourceVersion: object.GetResourceVersion(),
				},
				UpdateFunc: func(obj client.Object) {
					o.SelectedObject.Update(obj)
				},
				DeleteFunc: func(obj client.Object) {
					o.SelectedObject.Update(nil)
					if pin.Active() {
						pin.Activate()
					}
				},
			},
		)

		resource := o.GetAPIResource(object.GetObjectKind().GroupVersionKind())

		yaml, err := o.Encoder.EncodeYAML(object)
		if err != nil {
			o.sourceBuffer.SetText(fmt.Sprintf("error: %v", err))
		} else {
			o.sourceBuffer.SetText(string(yaml))
		}

		var props []api.Property
		for _, ext := range o.Extensions {
			props = ext.CreateObjectProperties(ctx, resource, object, props)
		}
		sort.Slice(props, func(i, j int) bool {
			return props[i].GetPriority() > props[j].GetPriority()
		})
		o.updateProperties(props)

		pinned := false
		for _, p := range o.ClusterPreferences.Value().Navigation.Pins {
			if p.UID == object.GetUID() {
				pinned = true
				break
			}
		}
		pin.SetActive(pinned)
	})

	return &o
}

func (o *ObjectView) updateProperties(properties []api.Property) {
	for _, g := range o.groups {
		o.prefPage.Remove(g)
	}
	o.groups = nil

	for i, prop := range properties {
		group := o.renderObjectProperty(0, i, prop).(*adw.PreferencesGroup)
		o.groups = append(o.groups, group)
		o.prefPage.Add(group)
	}
}

func (o *ObjectView) renderObjectProperty(level, index int, prop api.Property) gtk.Widgetter {
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
				prop.Widget(row, o.navView)
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
					obj, err := o.GetReference(o.ctx, *prop.Reference)
					if err != nil {
						log.Print(err.Error())
						return
					}
					ctx, cancel := context.WithCancel(o.ctx)
					state := *o.ClusterState
					state.SelectedObject = observer.NewProperty[client.Object](obj)
					dv := NewObjectView(ctx, &state, o.editor, o.navView, o.navigation)
					o.navView.Push(dv.NavigationPage)
					o.navView.ConnectPopped(func(page *adw.NavigationPage) {
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
				prop.Widget(box, o.navView)
			}
			return box
		}

	case *api.GroupProperty:
		switch level {
		case 0:
			group := adw.NewPreferencesGroup()
			group.SetTitle(prop.Name)
			for i, child := range prop.Children {
				group.Add(o.renderObjectProperty(level+1, i, child))
			}
			if prop.Widget != nil {
				prop.Widget(group, o.navView)
			}
			return group
		case 1:
			row := adw.NewExpanderRow()
			id := fmt.Sprintf("%s-%d-%d", o.SelectedObject.Value().GetObjectKind().GroupVersionKind().String(), level, index)
			if e, ok := o.expanded[id]; ok && e {
				row.SetExpanded(true)
			}
			row.Connect("state-flags-changed", func() {
				o.expanded[id] = row.Expanded()
			})
			row.SetTitle(prop.Name)
			for i, child := range prop.Children {
				row.AddRow(o.renderObjectProperty(level+1, i, child))
			}
			row.SetSensitive(len(prop.Children) > 0)
			if prop.Widget != nil {
				prop.Widget(row, o.navView)
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
				box.Insert(o.renderObjectProperty(level+1, i, child), -1)
				// prop.Value += fmt.Sprintf("%s: %s\n", child.Name, child.Value)
			}
			if prop.Widget != nil {
				prop.Widget(row, o.navView)
			}
			return row
		}
	}

	return nil
}

func (o *ObjectView) createSource() *gtk.ScrolledWindow {
	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetVExpand(true)

	o.sourceBuffer = gtksource.NewBufferWithLanguage(gtksource.LanguageManagerGetDefault().Language("yaml"))
	o.setSourceColorScheme()
	gtk.SettingsGetDefault().NotifyProperty("gtk-application-prefer-dark-theme", o.setSourceColorScheme)
	o.sourceView = gtksource.NewViewWithBuffer(o.sourceBuffer)
	o.sourceView.SetEditable(false)
	o.sourceView.SetWrapMode(gtk.WrapWord)
	o.sourceView.SetShowLineNumbers(true)
	o.sourceView.SetMonospace(true)
	scrolledWindow.SetChild(o.sourceView)

	return scrolledWindow
}

func (o *ObjectView) setSourceColorScheme() {
	util.SetSourceColorScheme(o.sourceBuffer)
}
