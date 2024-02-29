package ui

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/getseabird/seabird/internal/util"
	"github.com/getseabird/seabird/widget"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EditorWindow struct {
	*widget.UniversalWindow
	tabview  *adw.TabView
	resource *metav1.APIResource
	schema   *openapi3.T
	ref      string
}

func NewEditorWindow(ctx context.Context, resource *metav1.APIResource, object client.Object) (*EditorWindow, error) {
	schema, err := loadSchema(ctx, resource)
	if err != nil {
		return nil, err
	}

	s := strings.Split(resource.Group, ".")
	slices.Reverse(s)
	rdns := strings.Join(s, ".")
	if rdns == "" {
		rdns = "io.k8s.api.core"
	}

	w := EditorWindow{
		UniversalWindow: widget.NewUniversalWindow(),
		resource:        resource,
		ref:             fmt.Sprintf("#/components/schemas/%s.%s.%s", rdns, resource.Version, resource.Kind),
		schema:          schema,
	}
	w.SetTransientFor(ctxt.MustFrom[*gtk.Window](ctx))
	w.SetModal(true)
	w.SetDefaultSize(1000, 600)
	w.SetTitle("Editor")

	ctx = ctxt.With[*gtk.Window](ctx, w.Window)

	content := gtk.NewBox(gtk.OrientationVertical, 0)
	content.AddCSSClass("view")

	toast := adw.NewToastOverlay()
	toast.SetChild(content)
	w.SetContent(toast)

	toolbar := adw.NewToolbarView()
	toolbar.SetSizeRequest(200, -1)

	header := adw.NewHeaderBar()
	header.SetShowEndTitleButtons(false)
	header.AddCSSClass("flat")
	toolbar.AddTopBar(header)

	createButton := gtk.NewButton()
	createButton.SetLabel("Create")
	createButton.AddCSSClass("suggested-action")
	createButton.ConnectClicked(func() {
		buf := w.tabview.SelectedPage().Child().(*gtk.ScrolledWindow).Child().(*gtksource.View).Buffer()
		text := buf.Text(buf.StartIter(), buf.EndIter(), true)
		obj, err := util.YamlToUnstructured([]byte(text))
		if err != nil {
			widget.ShowErrorDialog(ctx, "Error decoding object", err)
			return
		}

		cluster := ctxt.MustFrom[*api.Cluster](ctx)

		if err := cluster.Create(ctx, obj); err != nil {
			widget.ShowErrorDialog(ctx, "Error creating object", err)
			return
		}

		toast.AddToast(adw.NewToast("Object created."))
	})
	header.PackEnd(createButton)

	w.tabview = adw.NewTabView()
	toolbar.SetContent(w.tabview)

	w.tabview.NotifyProperty("selected-page", func() {
		// w.SetTitle(w.tabview.SelectedPage().Title())
	})

	tabbar := adw.NewTabBar()
	tabbar.SetView(w.tabview)
	toolbar.AddTopBar(tabbar)

	paned := gtk.NewPaned(gtk.OrientationHorizontal)
	paned.SetStartChild(toolbar)
	docView := adw.NewNavigationView()
	docView.Push(w.newDocumentationPage(schema, w.ref))
	docView.SetSizeRequest(200, -1)
	paned.SetEndChild(docView)
	content.Append(paned)

	newButton := gtk.NewButton()
	newButton.SetIconName("tab-new-symbolic")
	newButton.ConnectClicked(w.newSource)
	header.PackStart(newButton)
	w.newSource()

	return &w, nil
}

func (w *EditorWindow) newSource() {
	buf := gtksource.NewBufferWithLanguage(gtksource.LanguageManagerGetDefault().Language("yaml"))

	buf.SetText(fmt.Sprintf("apiVersion: %s\nkind: %s\nmetadata:\n  name: example", util.ResourceGVR(w.resource).GroupVersion().String(), w.resource.Kind))
	util.SetSourceColorScheme(buf)
	source := gtksource.NewViewWithBuffer(buf)
	source.SetMarginStart(8)
	source.SetMarginEnd(8)
	source.SetMarginTop(8)
	source.SetMarginBottom(8)
	source.SetVExpand(true)

	// TODO highlight errors in sourceview
	// schema := resolveRef(w.schema, w.ref)
	// buf.ConnectChanged(func() {
	// 	text := buf.Text(buf.StartIter(), buf.EndIter(), true)
	// 	var data any
	// 	if err := yaml.Unmarshal([]byte(text), &data); err != nil {
	// 		log.Print(err.Error())
	// 		return
	// 	}
	// 	var opts []openapi3.SchemaValidationOption
	// 	opts = append(opts, openapi3.MultiErrors())
	// 	if err := schema.Value.VisitJSON(data, opts...); err != nil {
	// 		log.Print(err.Error())
	// 	}
	// })

	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetChild(source)
	scrolledWindow.SetVExpand(true)

	tabpage := w.tabview.Append(scrolledWindow)
	tabpage.SetTitle("New Object")
	w.tabview.SetSelectedPage(tabpage)
}

func (w *EditorWindow) newDocumentationPage(t *openapi3.T, ref string) *adw.NavigationPage {
	schema := resolveRef(t, ref)

	content := gtk.NewBox(gtk.OrientationVertical, 0)
	page := adw.NewNavigationPage(content, refName(ref))

	header := adw.NewHeaderBar()
	header.SetShowStartTitleButtons(false)
	header.AddCSSClass("flat")
	content.Append(header)

	prefs := adw.NewPreferencesPage()
	content.Append(prefs)

	group := adw.NewPreferencesGroup()
	// group.SetTitle(refName(ref))
	group.SetDescription(schema.Value.Description)
	prefs.Add(group)

	for prop, schema := range schema.Value.Properties {
		var required bool
		for _, r := range schema.Value.Required {
			if prop == r {
				required = true
				break
			}
		}
		row := adw.NewActionRow()
		row.SetUseMarkup(false)
		row.SetSubtitle(schema.Value.Description)

		var vtype string
		switch schema.Value.Type {
		case "array":
			vtype = fmt.Sprintf("[]%s", schema.Value.Items.Value.Type)
		default:
			vtype = schema.Value.Type
		}

		if inner := innerType(schema.Value); inner != nil {
			switch schema.Value.Type {
			case "array":
				vtype = fmt.Sprintf("[]%s", refName(inner.Ref))
			default:
				vtype = refName(inner.Ref)
			}
			row.SetActivatable(true)
			row.AddSuffix(gtk.NewImageFromIconName("go-next-symbolic"))
			row.ConnectActivated(func() {
				page.Parent().(*adw.NavigationView).Push(w.newDocumentationPage(t, inner.Ref))
			})
		}
		title := fmt.Sprintf("%s %s", prop, vtype)
		if required {
			title = fmt.Sprintf("%s* %s", prop, vtype)
		}
		row.SetTitle(title)
		group.Add(row)
	}

	return page
}

func innerType(schema *openapi3.Schema) *openapi3.SchemaRef {
	if schema.Items != nil {
		return innerType(schema.Items.Value)
	}
	switch {
	case len(schema.AllOf) > 0:
		return schema.AllOf[0]
	case len(schema.AnyOf) > 0:
		return schema.AnyOf[0]
	}

	return nil
}

func resolveRef(t *openapi3.T, ref string) *openapi3.SchemaRef {
	ref = strings.TrimPrefix(ref, "#/components/schemas/")
	return t.Components.Schemas[ref]
}

func refName(ref string) string {
	s := strings.Split(ref, ".")
	return s[len(s)-1]
}

func loadSchema(ctx context.Context, resource *metav1.APIResource) (*openapi3.T, error) {
	cluster := ctxt.MustFrom[*api.Cluster](ctx)

	paths, err := cluster.OpenAPIV3().Paths()
	if err != nil {
		return nil, err
	}

	var resourcePath string
	if len(resource.Group) == 0 {
		resourcePath = fmt.Sprintf("api/%s", resource.Version)
	} else {
		resourcePath = fmt.Sprintf("apis/%s/%s", resource.Group, resource.Version)
	}

	gv, exists := paths[resourcePath]
	if !exists {
		return nil, fmt.Errorf("couldn't find resource")
	}

	bytes, err := gv.Schema(runtime.ContentTypeJSON)
	if err != nil {
		return nil, err
	}

	return openapi3.NewLoader().LoadFromData(bytes)
}
