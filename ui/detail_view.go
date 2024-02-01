package ui

import (
	"fmt"
	"log"
	"runtime"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/getseabird/seabird/behavior"
	"github.com/getseabird/seabird/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		prefPage:       adw.NewPreferencesPage(),
		behavior:       behavior,
		parent:         parent,
		expanded:       map[string]bool{},
	}

	clamp := d.prefPage.FirstChild().(*gtk.ScrolledWindow).FirstChild().(*gtk.Viewport).FirstChild().(*adw.Clamp)
	clamp.SetTighteningThreshold(300)
	clamp.SetMaximumSize(10000)

	stack := adw.NewViewStack()
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

	onChange(d.behavior.SelectedObject, func(_ client.Object) {
		for {
			if !d.Parent().(*adw.NavigationView).Pop() {
				break
			}
		}
	})
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

	for i, prop := range properties {
		d.groups = append(d.groups, d.renderObjectProperty(0, i, prop).(*adw.PreferencesGroup))
	}

	for _, g := range d.groups {
		d.prefPage.Add(g)
	}
}

func (d *DetailView) renderObjectProperty(level, index int, prop behavior.ObjectProperty) gtk.Widgetter {
	switch level {
	case 0:
		g := adw.NewPreferencesGroup()
		g.SetTitle(prop.Name)
		for i, child := range prop.Children {
			g.Add(d.renderObjectProperty(level+1, i, child))
		}
		return g
	case 1:
		if len(prop.Children) > 0 {
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
			d.extendRow(row, level, prop)
			return row
		}
		fallthrough
	case 2:
		row := adw.NewActionRow()
		row.SetTitle(prop.Name)
		row.SetUseMarkup(false)
		row.AddCSSClass("property")

		if len(prop.Children) > 0 {
			box := gtk.NewFlowBox()
			box.SetColumnSpacing(8)
			box.SetSelectionMode(gtk.SelectionNone)
			row.FirstChild().(*gtk.Box).FirstChild().(*gtk.Box).NextSibling().(*gtk.Image).NextSibling().(*gtk.Box).Append(box)
			for i, child := range prop.Children {
				box.Insert(d.renderObjectProperty(level+1, i, child), -1)
				prop.Value += fmt.Sprintf("%s: %s\n", child.Name, child.Value)
			}
		} else {
			// *Very* long labels cause a segfault in GTK. Limiting lines prevents it, but they're still
			// slow and CPU-intensive to render. https://gitlab.gnome.org/GNOME/gtk/-/issues/1332
			// TODO explore alternative rendering options such as TextView
			row.SetSubtitleLines(5)
			row.SetSubtitle(prop.Value)
		}

		copy := gtk.NewButton()
		copy.SetIconName("edit-copy-symbolic")
		copy.AddCSSClass("flat")
		copy.AddCSSClass("dim-label")
		copy.SetVAlign(gtk.AlignCenter)
		copy.ConnectClicked(func() {
			gdk.DisplayGetDefault().Clipboard().SetText(prop.Value)
		})
		row.AddSuffix(copy)

		d.extendRow(row, level, prop)
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

		return box
	}

	return nil
}

// This is a bit of a hack, should probably extend ObjectProperty with this stuff...
func (d *DetailView) extendRow(widget gtk.Widgetter, level int, prop behavior.ObjectProperty) {
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

			for _, p := range prop.Children {
				if p.Name == "Memory" {
					v, err := resource.ParseQuantity(p.Value)
					if err != nil {
						log.Printf(err.Error())
					} else {
						widget.(*adw.ExpanderRow).AddSuffix(createMemoryBar(v, object.Resources))
					}
				}
			}
			for _, p := range prop.Children {
				if p.Name == "CPU" {
					v, err := resource.ParseQuantity(p.Value)
					if err != nil {
						log.Printf(err.Error())
					} else {
						widget.(*adw.ExpanderRow).AddSuffix(createCPUBar(v, object.Resources))
					}
				}
			}

			logs := adw.NewActionRow()
			logs.SetActivatable(true)
			logs.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
			logs.SetTitle("Logs")
			logs.ConnectActivated(func() {
				d.Parent().(*adw.NavigationView).Push(NewLogPage(d.parent, d.behavior, selected, object.Name).NavigationPage)
			})
			widget.(*adw.ExpanderRow).AddRow(logs)
		}

	case *appsv1.Deployment:
		switch object := prop.Object.(type) {
		case *corev1.Pod:
			for _, cond := range object.Status.Conditions {
				if cond.Type == corev1.ContainersReady {
					row := widget.(*adw.ActionRow)
					row.AddPrefix(createStatusIcon(cond.Status == corev1.ConditionTrue || cond.Reason == "PodCompleted"))
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

	case *corev1.Node:
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
	scrolledWindow.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)

	// TODO collapse managed fields
	// https://gitlab.gnome.org/swilmet/tepl
	// d.object.SetManagedFields([]metav1.ManagedFieldsEntry{})

	d.sourceBuffer = gtksource.NewBufferWithLanguage(gtksource.LanguageManagerGetDefault().Language("yaml"))
	d.setSourceColorScheme()
	gtk.SettingsGetDefault().NotifyProperty("gtk-application-prefer-dark-theme", d.setSourceColorScheme)
	sourceView := gtksource.NewViewWithBuffer(d.sourceBuffer)
	sourceView.SetMarginBottom(8)
	sourceView.SetMarginTop(8)
	sourceView.SetMarginEnd(8)
	sourceView.SetEditable(false)
	sourceView.SetWrapMode(gtk.WrapWord)
	sourceView.SetShowLineNumbers(true)
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

func createMemoryBar(actual resource.Quantity, res corev1.ResourceRequirements) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationVertical, 4)
	box.SetVAlign(gtk.AlignCenter)
	req := res.Requests.Memory()
	if req == nil || req.IsZero() {
		req = res.Limits.Memory()
	}
	if req == nil || req.IsZero() {
		return box
	}

	percent := actual.AsApproximateFloat64() / req.AsApproximateFloat64()
	levelBar := gtk.NewLevelBar()
	levelBar.SetSizeRequest(50, -1)
	levelBar.SetHAlign(gtk.AlignCenter)
	levelBar.SetVAlign(gtk.AlignCenter)
	levelBar.SetValue(min(percent, 1))
	// down from offset, not up
	levelBar.AddOffsetValue("lb-warning", .9)
	levelBar.AddOffsetValue("lb-error", 1)
	box.SetTooltipText(fmt.Sprintf("%.0f%% Memory", percent*100))

	box.Append(gtk.NewImageFromIconName("memory-stick-symbolic"))
	box.Append(levelBar)

	return box
}

func createCPUBar(actual resource.Quantity, res corev1.ResourceRequirements) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationVertical, 4)
	box.SetVAlign(gtk.AlignCenter)
	req := res.Requests.Cpu()
	if req == nil || req.IsZero() {
		req = res.Limits.Cpu()
	}
	if req == nil || req.IsZero() {
		return box
	}

	percent := actual.AsApproximateFloat64() / req.AsApproximateFloat64()
	levelBar := gtk.NewLevelBar()
	levelBar.SetSizeRequest(50, -1)
	levelBar.SetHAlign(gtk.AlignCenter)
	levelBar.SetVAlign(gtk.AlignCenter)
	levelBar.SetValue(min(percent, 1))
	levelBar.AddOffsetValue("lb-warning", .9)
	levelBar.AddOffsetValue("lb-error", 1)
	box.SetTooltipText(fmt.Sprintf("%.0f%% CPU", percent*100))

	box.Append(gtk.NewImageFromIconName("cpu-symbolic"))
	box.Append(levelBar)

	return box
}
