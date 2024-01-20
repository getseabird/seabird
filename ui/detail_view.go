package ui

import (
	"encoding/json"
	"log"
	"strconv"

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

func NewDetailView() *DetailView {
	d := DetailView{Box: gtk.NewBox(gtk.OrientationVertical, 0)}
	d.SetHExpand(true)

	stack := adw.NewViewStack()
	d.prefPage = d.properties()
	stack.AddTitledWithIcon(d.prefPage, "properties", "Properties", "document-properties-symbolic")
	stack.AddTitledWithIcon(d.source(), "source", "Source", "accessories-text-editor-symbolic")

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

	defer d.sourceBuffer.SetText(string(encodeToYaml(d.object)))

	d.nameLabel.SetText(object.GetName())
	d.namespaceLabel.SetText(object.GetNamespace())

	for _, r := range d.dynamicLabels {
		d.labelsRow.Remove(r)
	}
	d.dynamicLabels = []*adw.ActionRow{}

	for key, value := range object.GetLabels() {
		row := adw.NewActionRow()
		row.SetTitle(key)
		row.SetSubtitle(value)
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
		row := adw.NewActionRow()
		row.SetTitle(key)
		row.SetSubtitle(value)
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
		group = podProperties(object)
	case *corev1.ConfigMap:
		group = configMapProperties(object)
	case *corev1.Secret:
		group = secretProperties(object)
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

func (d *DetailView) properties() *adw.PreferencesPage {
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

func (d *DetailView) source() gtk.Widgetter {
	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetVExpand(true)
	// TODO collapse instead of remove
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

func encodeToYaml(object client.Object) []byte {
	codec := serializer.NewCodecFactory(application.cluster.Scheme).LegacyCodec(application.cluster.Scheme.PreferredVersionAllGroups()...)
	encoded, err := runtime.Encode(codec, object)
	if err != nil {
		log.Printf("failed to encode object: %v", err)
		return []byte{}
	}

	return jsonToYaml(encoded)
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
