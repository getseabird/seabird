package ui

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DetailView struct {
	*gtk.Box
	root               *ClusterWindow
	object             client.Object
	prefPage           *adw.PreferencesPage
	dynamicGroups      []*adw.PreferencesGroup
	nameLabel          *gtk.Label
	namespaceLabel     *gtk.Label
	labelsRow          *adw.ExpanderRow
	labelsSuffix       *adw.Bin
	dynamicLabels      []*adw.ActionRow
	annotationsRow     *adw.ExpanderRow
	annotationsSuffix  *adw.Bin
	dynamicAnnotations []*adw.ActionRow
	sourceBuffer       *gtksource.Buffer
}

func NewDetailView(root *ClusterWindow) *DetailView {
	d := DetailView{Box: gtk.NewBox(gtk.OrientationVertical, 0), root: root}

	stack := adw.NewViewStack()
	d.prefPage = d.createProperties()
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

	return &d
}

func (d *DetailView) SetObject(object client.Object) {
	d.object = object

	defer d.sourceBuffer.SetText(string(jsonToYaml(d.encode(d.object))))

	d.nameLabel.SetText(object.GetName())
	d.namespaceLabel.SetText(object.GetNamespace())

	for _, r := range d.dynamicLabels {
		d.labelsRow.Remove(r)
	}
	d.dynamicLabels = []*adw.ActionRow{}

	for key, value := range object.GetLabels() {
		// workaround for annoying gtk warning (libadwaita bug?)
		if len(value) < 5 {
			value = fmt.Sprintf("%s     ", value)
		}
		row := adw.NewActionRow()
		row.SetTitle(key)
		row.SetSubtitle(value)
		row.SetTooltipText(value)
		row.SetSubtitleLines(1)
		row.AddCSSClass("property")
		d.labelsRow.AddRow(row)
		d.dynamicLabels = append(d.dynamicLabels, row)
	}

	d.labelsSuffix.SetChild(gtk.NewLabel(strconv.Itoa(len(d.dynamicLabels))))

	for _, r := range d.dynamicAnnotations {
		d.annotationsRow.Remove(r)
	}
	d.dynamicAnnotations = []*adw.ActionRow{}

	for key, value := range object.GetAnnotations() {
		// workaround for annoying gtk warning (libadwaita bug?)
		if len(value) < 5 {
			value = fmt.Sprintf("%s     ", value)
		}
		row := adw.NewActionRow()
		row.SetTitle(key)
		row.SetSubtitle(value)
		row.SetTooltipText(value)
		row.SetSubtitleLines(1)
		row.AddCSSClass("property")
		d.annotationsRow.AddRow(row)
		d.dynamicAnnotations = append(d.dynamicAnnotations, row)
	}

	d.annotationsSuffix.SetChild(gtk.NewLabel(strconv.Itoa(len(d.dynamicAnnotations))))

	for _, g := range d.dynamicGroups {
		d.prefPage.Remove(g)
	}
	d.dynamicGroups = []*adw.PreferencesGroup{}

	var group *adw.PreferencesGroup
	switch object := d.object.(type) {
	case *corev1.Pod:
		group = d.podProperties(object)
	case *corev1.ConfigMap:
		group = d.configMapProperties(object)
	case *corev1.Secret:
		group = d.secretProperties(object)
	}
	if group != nil {
		d.prefPage.Add(group)
		d.dynamicGroups = append(d.dynamicGroups, group)
	}
}

func actionRow(title string, suffix gtk.Widgetter) *adw.ActionRow {
	row := adw.NewActionRow()
	row.SetTitle(title)
	row.AddSuffix(suffix)
	return row
}

func (d *DetailView) createProperties() *adw.PreferencesPage {
	page := adw.NewPreferencesPage()
	page.SetSizeRequest(400, 400)
	group := adw.NewPreferencesGroup()
	group.SetTitle("Metadata")
	d.nameLabel = gtk.NewLabel("")
	group.Add(actionRow("Name", d.nameLabel))
	d.namespaceLabel = gtk.NewLabel("")
	group.Add(actionRow("Namespace", d.namespaceLabel))

	d.labelsRow = adw.NewExpanderRow()
	d.labelsRow.SetTitle("Labels")
	d.labelsSuffix = adw.NewBin()
	d.labelsRow.AddSuffix(d.labelsSuffix)
	group.Add(d.labelsRow)

	d.annotationsRow = adw.NewExpanderRow()
	d.annotationsRow.SetTitle("Annotations")
	d.annotationsSuffix = adw.NewBin()
	d.annotationsRow.AddSuffix(d.annotationsSuffix)
	group.Add(d.annotationsRow)

	page.Add(group)

	return page
}

func (d *DetailView) createSource() *gtk.ScrolledWindow {
	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetVExpand(true)
	// TODO collapse instead of remove
	// https://gitlab.gnome.org/swilmet/tepl
	// d.object.SetManagedFields([]metav1.ManagedFieldsEntry{})

	d.sourceBuffer = gtksource.NewBufferWithLanguage(gtksource.LanguageManagerGetDefault().Language("yaml"))
	d.sourceBuffer.SetStyleScheme(gtksource.StyleSchemeManagerGetDefault().Scheme("Adwaita-dark"))
	sourceView := gtksource.NewViewWithBuffer(d.sourceBuffer)
	sourceView.SetMarginBottom(8)
	sourceView.SetMarginTop(8)
	sourceView.SetMarginStart(8)
	sourceView.SetMarginEnd(8)
	sourceView.SetEditable(false)
	scrolledWindow.SetChild(sourceView)

	return scrolledWindow
}

func (d *DetailView) podProperties(pod *corev1.Pod) *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()
	group.SetTitle("Containers")

	for _, container := range pod.Spec.Containers {
		var status corev1.ContainerStatus
		for _, s := range pod.Status.ContainerStatuses {
			if s.Name == container.Name {
				status = s
				break
			}
		}

		expander := adw.NewExpanderRow()
		expander.SetTitle(container.Name)
		group.Add(expander)

		if status.Ready {
			icon := gtk.NewImageFromIconName("emblem-ok-symbolic")
			icon.AddCSSClass("success")
			expander.AddSuffix(icon)
		} else {
			icon := gtk.NewImageFromIconName("dialog-warning")
			icon.AddCSSClass("warning")
			expander.AddSuffix(icon)
		}

		row := adw.NewActionRow()
		row.AddCSSClass("property")
		row.SetTitle("Image")
		row.SetSubtitle(container.Image)
		expander.AddRow(row)
		if len(container.Command) > 0 {
			row = adw.NewActionRow()
			row.AddCSSClass("property")
			row.SetTitle("Command")
			row.SetSubtitle(strings.Join(container.Command, " "))
			expander.AddRow(row)
		}
		if len(container.Env) > 0 {
			var env []string
			for _, e := range container.Env {
				if e.ValueFrom != nil {
					// TODO
				} else {
					env = append(env, fmt.Sprintf("%s=%v", e.Name, e.Value))
				}
			}
			row = adw.NewActionRow()
			row.AddCSSClass("property")
			row.SetTitle("Env")
			row.SetSubtitle(strings.Join(env, " "))
			expander.AddRow(row)
		}

		row = adw.NewActionRow()
		row.AddCSSClass("property")
		row.SetTitle("State")
		if status.State.Running != nil {
			row.SetSubtitle("Running")
		} else if status.State.Terminated != nil {
			row.SetSubtitle(fmt.Sprintf("Terminated: %s", status.State.Terminated.Reason))
		} else if status.State.Waiting != nil {
			row.SetSubtitle(fmt.Sprintf("Waiting: %s", status.State.Waiting.Reason))
		}
		expander.AddRow(row)

		row = adw.NewActionRow()
		row.SetActivatable(true)
		row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
		row.SetTitle("Logs")
		row.ConnectActivated(func() {
			NewLogWindow(d.root, pod, &container).Show()
		})
		expander.AddRow(row)
	}

	return group
}

func (d *DetailView) secretProperties(object *corev1.Secret) *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()
	group.SetTitle("Data")

	for key, value := range object.Data {
		row := adw.NewActionRow()
		row.AddCSSClass("property")
		row.SetTitle(key)
		row.SetSubtitle(string(value))
		group.Add(row)
	}

	return group
}

func (d *DetailView) configMapProperties(object *corev1.ConfigMap) *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()
	group.SetTitle("Data")

	for key, value := range object.Data {
		row := adw.NewActionRow()
		row.AddCSSClass("property")
		row.SetTitle(key)
		row.SetSubtitle(value)
		group.Add(row)
	}

	return group
}

func (d *DetailView) encode(object client.Object) []byte {
	codec := serializer.NewCodecFactory(d.root.cluster.Scheme).LegacyCodec(d.root.cluster.Scheme.PreferredVersionAllGroups()...)
	encoded, err := runtime.Encode(codec, object)
	if err != nil {
		log.Printf("failed to encode object: %v", err)
		return []byte{}
	}
	return encoded
}

func jsonToYaml(data []byte) []byte {
	var o any
	if err := json.Unmarshal(data, &o); err != nil {
		panic(err)
	}
	ret, err := yaml.Marshal(o)
	if err != nil {
		panic(err)
	}
	return ret
}
