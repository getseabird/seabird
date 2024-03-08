package editor

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4-sourceview/pkg/gtksource/v5"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/getseabird/seabird/internal/util"
	"github.com/getseabird/seabird/widget"
	"github.com/imkira/go-observer/v2"
	"github.com/zmwangx/debounce"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type sourcePage struct {
	*gtk.Paned
	source *gtksource.View
	nav    *adw.NavigationView
	ctx    context.Context
	gvk    *schema.GroupVersionKind
	object client.Object
	schema *openapi3.T
	ref    string
	title  observer.Property[string]
}

func newSourcePage(ctx context.Context, gvk *schema.GroupVersionKind, object client.Object, title observer.Property[string]) (*sourcePage, error) {
	cluster := ctxt.MustFrom[*api.Cluster](ctx)
	buf := gtksource.NewBufferWithLanguage(gtksource.LanguageManagerGetDefault().Language("yaml"))
	if object != nil {
		title.Update(object.GetName())
		yaml, err := cluster.Encoder.EncodeYAML(object)
		if err != nil {
			widget.ShowErrorDialog(ctx, "Error encoding object", err)
		} else {
			buf.SetText(string(yaml))
		}
	} else {
		g := schema.GroupVersionKind{Version: "v1", Kind: "Pod"}
		if gvk != nil {
			g = *gvk
		}
		buf.SetText(fmt.Sprintf("apiVersion: %s\nkind: %s\nmetadata:\n  name: example\n  namespace: default", metav1.GroupVersion{Group: g.Group, Version: g.Version}, g.Kind))
	}

	paned := gtk.NewPaned(gtk.OrientationHorizontal)

	util.SetSourceColorScheme(buf)
	source := gtksource.NewViewWithBuffer(buf)
	source.SetMarginStart(8)
	source.SetMarginEnd(8)
	source.SetMarginTop(8)
	source.SetMarginBottom(8)
	source.SetMonospace(true)
	source.SetVExpand(true)

	scrolledWindow := gtk.NewScrolledWindow()
	scrolledWindow.SetChild(source)
	scrolledWindow.SetSizeRequest(500, -1)
	paned.SetStartChild(scrolledWindow)
	sw, _ := scrolledWindow.SizeRequest()
	paned.SetPosition(sw)

	nav := adw.NewNavigationView()
	nav.SetSizeRequest(500, -1)
	paned.SetEndChild(nav)

	page := &sourcePage{
		Paned:  paned,
		source: source,
		ctx:    ctx,
		gvk:    gvk,
		object: object,
		nav:    nav,
		title:  title,
	}

	if gvk != nil {
		if err := page.updateKind(*gvk); err != nil {
			return nil, err
		}
	}

	changed, ctrl := debounce.Debounce(func() {
		glib.IdleAdd(func() {
			text := buf.Text(buf.StartIter(), buf.EndIter(), true)
			obj, err := util.YamlToUnstructured([]byte(text))
			if err != nil {
				return
			}

			if title.Value() != obj.GetName() {
				title.Update(obj.GetName())
			}

			gvk := obj.GetObjectKind().GroupVersionKind()
			if page.gvk == nil || !util.GVKEquals(*page.gvk, gvk) {
				page.updateKind(gvk)
			}

			page.object = obj
		})

		// TODO highlight errors in sourceview
		// var opts []openapi3.SchemaValidationOption
		// opts = append(opts, openapi3.MultiErrors())
		// if err := schema.Value.VisitJSON(data, opts...); err != nil {
		// 	log.Print(err.Error())
		// }

	}, 500*time.Millisecond)
	buf.ConnectChanged(changed)
	changed()
	ctrl.Flush()

	return page, nil
}

func (page *sourcePage) updateKind(gvk schema.GroupVersionKind) error {
	page.gvk = &gvk

	var err error
	page.schema, err = loadSchema(page.ctx, gvk)
	if err != nil {
		return err
	}

	s := strings.Split(gvk.Group, ".")
	slices.Reverse(s)
	rdns := strings.Join(s, ".")
	switch {
	case gvk.Group == "":
		rdns = "io.k8s.api.core"
	case gvk.Group == "apps":
		rdns = "io.k8s.api.apps"
	case strings.HasSuffix(gvk.Group, "k8s.io"):
		rdns = strings.ReplaceAll(rdns, "io.k8s", "io.k8s.api")
	}

	page.ref = fmt.Sprintf("#/components/schemas/%s.%s.%s", rdns, gvk.Version, gvk.Kind)
	if resolveRef(page.schema, page.ref) == nil {
		return errors.New("component schema not found")
	}

	page.nav.Replace([]*adw.NavigationPage{newDocumentationPage(page.schema, page.ref, nil)})

	return nil
}

func loadSchema(ctx context.Context, gvk schema.GroupVersionKind) (*openapi3.T, error) {
	cluster := ctxt.MustFrom[*api.Cluster](ctx)

	paths, err := cluster.OpenAPIV3().Paths()
	if err != nil {
		return nil, err
	}

	var resourcePath string
	if len(gvk.Group) == 0 {
		resourcePath = fmt.Sprintf("api/%s", gvk.Version)
	} else {
		resourcePath = fmt.Sprintf("apis/%s/%s", gvk.Group, gvk.Version)
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
