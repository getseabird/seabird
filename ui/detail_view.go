package ui

import (
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/behavior"
	corev1 "k8s.io/api/core/v1"
)

type DetailView struct {
	*gtk.Box
	parent   *gtk.Window
	behavior *behavior.DetailBehavior
	prefPage *adw.PreferencesPage
	groups   []*adw.PreferencesGroup

	sourceBuffer *gtksource.Buffer
}

func NewDetailView(parent *gtk.Window, behavior *behavior.DetailBehavior) *DetailView {
	d := DetailView{
		Box:      gtk.NewBox(gtk.OrientationVertical, 0),
		behavior: behavior,
		parent:   parent,
	}

	stack := adw.NewViewStack()

	d.prefPage = adw.NewPreferencesPage()
	d.prefPage.SetSizeRequest(350, 350)

	stack.AddTitledWithIcon(d.prefPage, "properties", "Properties", "document-properties-symbolic")
	stack.AddTitledWithIcon(d.createSource(), "source", "Source", "accessories-text-editor-symbolic")

	header := adw.NewHeaderBar()
	header.AddCSSClass("flat")
	switcher := adw.NewViewSwitcher()
	switcher.SetPolicy(adw.ViewSwitcherPolicyWide)
	switcher.SetStack(stack)
	header.SetTitleWidget(switcher)

	d.Append(header)
	d.Append(stack)

	onChange(d.behavior.Yaml, func(yaml string) {
		d.sourceBuffer.SetText(string(yaml))
	})
	onChange(d.behavior.Properties, d.onPropertiesChange)

	return &d
}

func (d *DetailView) onPropertiesChange(properties []behavior.ObjectProperty) {
	for _, g := range d.groups {
		d.prefPage.Remove(g)
	}
	d.groups = nil

	for _, p1 := range properties {
		g := adw.NewPreferencesGroup()
		g.SetTitle(p1.Name)
		d.groups = append(d.groups, g)
		for _, p2 := range p1.Children {
			if len(p2.Children) > 0 {
				r2 := adw.NewExpanderRow()
				r2.SetTitle(p2.Name)
				g.Add(r2)
				for _, p3 := range p2.Children {
					if len(p3.Children) > 0 {
						r3 := adw.NewActionRow()
						r3.SetTitle(p3.Name)
						r3.AddCSSClass("property")
						r2.AddRow(r3)

						box := gtk.NewFlowBox()
						box.SetColumnSpacing(2)
						box.SetRowSpacing(2)
						r3.FirstChild().(*gtk.Box).FirstChild().(*gtk.Box).NextSibling().(*gtk.Image).NextSibling().(*gtk.Box).Append(box)

						for _, p4 := range p3.Children {
							label := gtk.NewLabel(fmt.Sprintf("%s: %s", p4.Name, p4.Value))
							label.SetSelectable(true)
							label.AddCSSClass("badge")
							box.Insert(label, -1)
						}
					} else {
						r3 := adw.NewActionRow()
						r3.SetTitle(p3.Name)
						r3.SetSubtitle(p3.Value)
						r3.SetSubtitleSelectable(true)
						r3.AddCSSClass("property")
						r2.AddRow(r3)
					}

				}
				d.extendRow([]string{p1.Name, p2.Name}, r2)
			} else {
				r2 := adw.NewActionRow()
				r2.SetTitle(p2.Name)
				r2.SetSubtitle(p2.Value)
				r2.SetSubtitleSelectable(true)
				r2.AddCSSClass("property")
				g.Add(r2)
			}
		}
	}

	for _, g := range d.groups {
		d.prefPage.Add(g)
	}
}

func (d *DetailView) extendRow(path []string, widget gtk.Widgetter) {
	switch obj := d.behavior.SelectedObject.Value().(type) {
	case *corev1.Pod:
		if len(path) == 2 && path[0] == "Containers" {
			var status corev1.ContainerStatus
			for _, s := range obj.Status.ContainerStatuses {
				if s.Name == path[1] {
					status = s
				}
			}
			if status.Ready {
				icon := gtk.NewImageFromIconName("emblem-ok-symbolic")
				icon.AddCSSClass("success")
				widget.(*adw.ExpanderRow).AddSuffix(icon)
			} else {
				icon := gtk.NewImageFromIconName("dialog-warning-symbolic")
				icon.AddCSSClass("warning")
				widget.(*adw.ExpanderRow).AddSuffix(icon)
			}

			row := adw.NewActionRow()
			row.SetActivatable(true)
			row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
			row.SetTitle("Logs")
			row.ConnectActivated(func() {
				var container corev1.Container
				for _, c := range obj.Spec.Containers {
					if c.Name == path[1] {
						container = c
					}
				}
				NewLogWindow(d.parent, d.behavior, &container).Show()
			})
			widget.(*adw.ExpanderRow).AddRow(row)
		}

	}
}

func (d *DetailView) createSource() *gtk.ScrolledWindow {
	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetVExpand(true)
	// TODO collapse instead of remove
	// https://gitlab.gnome.org/swilmet/tepl
	// d.object.SetManagedFields([]metav1.ManagedFieldsEntry{})

	d.sourceBuffer = gtksource.NewBufferWithLanguage(gtksource.LanguageManagerGetDefault().Language("yaml"))
	d.setSourceColorScheme()
	gtk.SettingsGetDefault().NotifyProperty("gtk-application-prefer-dark-theme", d.setSourceColorScheme)
	sourceView := gtksource.NewViewWithBuffer(d.sourceBuffer)
	sourceView.SetMarginBottom(8)
	sourceView.SetMarginTop(8)
	sourceView.SetMarginStart(8)
	sourceView.SetMarginEnd(8)
	sourceView.SetEditable(false)
	scrolledWindow.SetChild(sourceView)

	return scrolledWindow
}

func (d *DetailView) setSourceColorScheme() {
	if gtk.SettingsGetDefault().ObjectProperty("gtk-application-prefer-dark-theme").(bool) {
		d.sourceBuffer.SetStyleScheme(gtksource.StyleSchemeManagerGetDefault().Scheme("Adwaita-dark"))
	} else {
		d.sourceBuffer.SetStyleScheme(gtksource.StyleSchemeManagerGetDefault().Scheme("Adwaita"))
	}
}
