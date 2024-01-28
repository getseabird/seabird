package ui

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/behavior"
	"github.com/getseabird/seabird/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type DetailView struct {
	*adw.NavigationPage
	parent       *gtk.Window
	behavior     *behavior.DetailBehavior
	prefPage     *adw.PreferencesPage
	groups       []*adw.PreferencesGroup
	sourceBuffer *gtksource.Buffer
	expanded     map[string]bool
}

func NewDetailView(parent *gtk.Window, behavior *behavior.DetailBehavior) *DetailView {
	content := gtk.NewBox(gtk.OrientationVertical, 0)

	d := DetailView{
		NavigationPage: adw.NewNavigationPage(content, "main"),
		behavior:       behavior,
		parent:         parent,
		expanded:       map[string]bool{},
	}

	stack := adw.NewViewStack()

	d.prefPage = adw.NewPreferencesPage()
	d.prefPage.SetSizeRequest(350, 350)

	stack.AddTitledWithIcon(d.prefPage, "properties", "Properties", "document-properties-symbolic")
	stack.AddTitledWithIcon(d.createSource(), "source", "Source", "accessories-text-editor-symbolic")

	header := adw.NewHeaderBar()
	header.AddCSSClass("flat")
	header.SetShowEndTitleButtons(runtime.GOOS != "windows")
	switcher := adw.NewViewSwitcher()
	switcher.SetPolicy(adw.ViewSwitcherPolicyWide)
	switcher.SetStack(stack)
	header.SetTitleWidget(switcher)

	content.Append(header)
	content.Append(stack)

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

	for _, prop := range properties {
		d.groups = append(d.groups, d.renderObjectProperty(0, prop).(*adw.PreferencesGroup))
	}

	for _, g := range d.groups {
		d.prefPage.Add(g)
	}
}

func (d *DetailView) renderObjectProperty(level uint, prop behavior.ObjectProperty) gtk.Widgetter {
	switch level {
	case 0:
		g := adw.NewPreferencesGroup()
		g.SetTitle(prop.Name)
		for _, child := range prop.Children {
			g.Add(d.renderObjectProperty(level+1, child))
		}
		return g
	case 1:
		if len(prop.Children) > 0 {
			row := adw.NewExpanderRow()
			id := strings.Join([]string{
				util.ResourceGVR(d.behavior.SelectedResource.Value()).String(),
				string(d.behavior.SelectedObject.Value().GetUID()),
				prop.Name,
				strconv.Itoa(int(level)),
			}, "-")
			if e, ok := d.expanded[id]; ok && e {
				row.SetExpanded(true)
			}
			row.Connect("state-flags-changed", func() {
				d.expanded[id] = row.Expanded()
			})
			row.SetTitle(prop.Name)
			for _, child := range prop.Children {
				row.AddRow(d.renderObjectProperty(level+1, child))
			}
			d.extendRow(row, level, prop)
			return row
		}
		fallthrough
	case 2:
		row := adw.NewActionRow()
		row.SetTitle(prop.Name)
		if len(prop.Children) > 0 {
			box := gtk.NewFlowBox()
			box.SetColumnSpacing(2)
			box.SetRowSpacing(2)
			row.FirstChild().(*gtk.Box).FirstChild().(*gtk.Box).NextSibling().(*gtk.Image).NextSibling().(*gtk.Box).Append(box)
			for _, child := range prop.Children {
				box.Insert(d.renderObjectProperty(level+1, child), -1)
			}
		} else {
			row.SetSubtitle(prop.Value)
			row.SetSubtitleSelectable(true)
		}
		row.AddCSSClass("property")

		d.extendRow(row, level, prop)
		return row
	case 3:
		label := gtk.NewLabel(fmt.Sprintf("%s: %s", prop.Name, prop.Value))
		label.SetSelectable(true)
		label.AddCSSClass("badge")
		return label
	}

	return nil
}

// This is a bit of a hack, should probably extend ObjectProperty with this stuff...
func (d *DetailView) extendRow(widget gtk.Widgetter, level uint, prop behavior.ObjectProperty) {
	switch selected := d.behavior.SelectedObject.Value().(type) {
	case *corev1.Pod:
		switch object := prop.Object.(type) {
		case *corev1.Container:
			var status corev1.ContainerStatus
			for _, s := range selected.Status.ContainerStatuses {
				if s.Name == object.Name {
					status = s
					break
				}
			}
			widget.(*adw.ExpanderRow).AddPrefix(createStatusIcon(status.Ready))

			var currentMem *resource.Quantity
			for _, p := range prop.Children {
				if p.Name == "Memory" {
					v, err := resource.ParseQuantity(p.Value)
					if err != nil {
						log.Printf(err.Error())
					} else {
						currentMem = &v
					}
				}
			}
			if currentMem != nil {
				targetMem := object.Resources.Requests.Memory()
				if targetMem == nil || targetMem.IsZero() {
					targetMem = object.Resources.Limits.Memory()
				}
				if targetMem != nil && !targetMem.IsZero() {
					percent := currentMem.AsApproximateFloat64() / targetMem.AsApproximateFloat64()
					levelBar := gtk.NewLevelBar()
					levelBar.SetSizeRequest(50, -1)
					levelBar.SetHAlign(gtk.AlignCenter)
					levelBar.SetVAlign(gtk.AlignCenter)
					levelBar.SetValue(min(percent, 1))
					// down from offset, not up
					levelBar.AddOffsetValue("lb-warning", .9)
					levelBar.AddOffsetValue("lb-error", 1)
					levelBar.SetTooltipText(fmt.Sprintf("%.0f%%", percent*100))
					widget.(*adw.ExpanderRow).AddSuffix(levelBar)
					widget.(*adw.ExpanderRow).AddSuffix(gtk.NewImageFromIconName("memory-stick-symbolic"))
				}
			}

			logs := adw.NewActionRow()
			logs.SetActivatable(true)
			logs.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
			logs.SetTitle("Logs")
			logs.ConnectActivated(func() {
				d.Parent().(*adw.NavigationView).Push(NewLogPage(d.parent, d.behavior, object).NavigationPage)
			})
			widget.(*adw.ExpanderRow).AddRow(logs)
		}

	case *appsv1.Deployment:
		switch object := prop.Object.(type) {
		case *corev1.Pod:
			for _, cond := range object.Status.Conditions {
				if cond.Type == corev1.ContainersReady {
					row := widget.(*adw.ActionRow)
					row.AddPrefix(createStatusIcon(cond.Status == corev1.ConditionTrue))
					row.SetActivatable(true)
					row.SetSubtitleSelectable(false)
					row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
					row.ConnectActivated(func() {
						dv := NewDetailView(d.parent, d.behavior.NewDetailBehavior())
						dv.behavior.SelectedObject.Update(object)
						d.Parent().(*adw.NavigationView).Push(dv.NavigationPage)
					})
				}
			}
		}

	case *appsv1.StatefulSet:
		switch object := prop.Object.(type) {
		case *corev1.Pod:
			for _, cond := range object.Status.Conditions {
				if cond.Type == corev1.ContainersReady {
					row := widget.(*adw.ActionRow)
					row.AddPrefix(createStatusIcon(cond.Status == corev1.ConditionTrue))
					row.SetActivatable(true)
					row.SetSubtitleSelectable(false)
					row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
					row.ConnectActivated(func() {
						dv := NewDetailView(d.parent, d.behavior.NewDetailBehavior())
						dv.behavior.SelectedObject.Update(object)
						d.Parent().(*adw.NavigationView).Push(dv.NavigationPage)
					})
				}
			}
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
