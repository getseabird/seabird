package ui

import (
	"encoding/json"
	"fmt"
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
	object client.Object
}

func NewDetailView(object client.Object) *DetailView {
	detailView := DetailView{Box: gtk.NewBox(gtk.OrientationVertical, 0), object: object}
	detailView.SetHExpand(true)

	viewStack := adw.NewViewStack()
	_ = viewStack.AddTitledWithIcon(detailView.properties(), "properties", "Properties", "document-properties-symbolic")
	_ = viewStack.AddTitledWithIcon(detailView.source(), "source", "Source", "accessories-text-editor-symbolic")
	viewSwitcherBar := adw.NewViewSwitcherBar()

	viewSwitcherBar.SetStack(viewStack)
	viewSwitcherBar.SetReveal(true)
	viewSwitcherBar.AddCSSClass("bg-red")
	detailView.Append(viewSwitcherBar)
	detailView.Append(viewStack)

	return &detailView
}

func actionRow(title string, suffix gtk.Widgetter) *adw.ActionRow {
	row := adw.NewActionRow()
	row.SetTitle(title)
	row.AddSuffix(suffix)
	return row
}

func (d *DetailView) properties() *adw.PreferencesPage {
	page := adw.NewPreferencesPage()
	page.SetSizeRequest(400, 100)
	page.SetHExpand(false)
	group := adw.NewPreferencesGroup()
	group.SetTitle("Metadata")
	group.Add(actionRow("Name", gtk.NewLabel(d.object.GetName())))
	group.Add(actionRow("Namespace", gtk.NewLabel(d.object.GetNamespace())))
	page.Add(group)

	switch object := d.object.(type) {
	case *corev1.Pod:
		group := adw.NewPreferencesGroup()
		group.SetTitle("Containers")
		page.Add(group)
		for _, container := range object.Spec.Containers {
			row := adw.NewExpanderRow()
			row.SetTitle(container.Name)
			status := gtk.NewImageFromIconName("emblem-default-symbolic")
			status.AddCSSClass("container-status-ok")
			row.AddAction(status)
			group.Add(row)

			ar := adw.NewActionRow()
			ar.AddCSSClass("property") // 1.4+
			ar.SetTitle("Image")
			ar.SetSubtitle(container.Image)
			row.AddRow(ar)
			if len(container.Command) > 0 {
				ar = adw.NewActionRow()
				ar.AddCSSClass("property")
				ar.SetTitle("Command")
				ar.SetSubtitle(strings.Join(container.Command, " "))
				row.AddRow(ar)
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
				ar = adw.NewActionRow()
				ar.AddCSSClass("property")
				ar.SetTitle("Env")
				ar.SetSubtitle(strings.Join(env, " "))
				row.AddRow(ar)
			}
		}
	}
	return page
}

func (d *DetailView) source() gtk.Widgetter {
	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetVExpand(true)
	viewport := gtk.NewViewport(nil, nil)
	scrolledWindow.SetChild(viewport)

	// TODO collapse instead of remove
	// d.object.SetManagedFields([]metav1.ManagedFieldsEntry{})

	buffer := gtksource.NewBufferWithLanguage(gtksource.LanguageManagerGetDefault().Language("yaml"))
	buffer.SetText(string(encodeToYaml(d.object)))
	sourceView := gtksource.NewViewWithBuffer(buffer)
	sourceView.SetEditable(false)
	viewport.SetChild(sourceView)

	return scrolledWindow
}

func encodeToYaml(object client.Object) []byte {
	codec := serializer.NewCodecFactory(application.cluster.Scheme).LegacyCodec(corev1.SchemeGroupVersion)
	encoded, err := runtime.Encode(codec, object)
	if err != nil {
		panic(err)
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
