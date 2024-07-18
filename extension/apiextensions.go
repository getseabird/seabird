package extension

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/getseabird/seabird/api"
	"github.com/imkira/go-observer/v2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/jsonpath"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	Extensions = append(Extensions, NewApiextensions)
}

func NewApiextensions(ctx context.Context, cluster *api.Cluster) (Extension, error) {
	if err := apiextensionsv1.AddToScheme(cluster.Scheme); err != nil {
		return nil, err
	}

	// TODO apiextensions v1 is not part of default client-go
	// api.InformerConnectProperty(ctx, cluster, apiextensionsv1.SchemeGroupVersion.WithResource("customresourcedefinitions"), crds)

	var crds apiextensionsv1.CustomResourceDefinitionList
	if err := cluster.Client.List(ctx, &crds); err != nil {
		return nil, err
	}

	return &Apiextensions{Cluster: cluster, crds: observer.NewProperty(crds.Items)}, nil
}

type Apiextensions struct {
	Noop
	*api.Cluster
	crds observer.Property[[]apiextensionsv1.CustomResourceDefinition]
}

func (e *Apiextensions) CreateColumns(ctx context.Context, resource *metav1.APIResource, columns []api.Column) []api.Column {
	for column, path := range e.getAdditionalColumnPaths(resource) {
		columns = append(columns, api.Column{
			Name:     column.Name,
			Priority: column.Priority * -1,
			Bind: func(cell *gtk.ColumnViewCell, object client.Object) {
				value, err := e.resolvePath(path, object)
				if err != nil {
					return
				}
				label := gtk.NewLabel(ptr.Deref(value, ""))
				label.SetHAlign(gtk.AlignStart)
				label.SetEllipsize(pango.EllipsizeEnd)
				cell.SetChild(label)
			},
		})
	}

	return columns
}

func (e *Apiextensions) CreateObjectProperties(ctx context.Context, resource *metav1.APIResource, object client.Object, props []api.Property) []api.Property {
	group := api.GroupProperty{Name: resource.Kind}
	for column, path := range e.getAdditionalColumnPaths(resource) {
		value, err := e.resolvePath(path, object)
		if err != nil {
			continue
		}
		group.Children = append(group.Children, &api.TextProperty{
			Name:     column.Name,
			Value:    ptr.Deref(value, ""),
			Priority: column.Priority * -1,
		})
	}
	props = append(props, &group)

	return props
}

func (e *Apiextensions) getAdditionalColumnPaths(resource *metav1.APIResource) map[apiextensionsv1.CustomResourceColumnDefinition]*jsonpath.JSONPath {
	var crd *apiextensionsv1.CustomResourceDefinitionVersion
	for _, c := range e.crds.Value() {
		if resource.Group == c.Spec.Group && resource.Kind == c.Spec.Names.Kind {
			for _, v := range c.Spec.Versions {
				if v.Name == resource.Version {
					crd = &v
					break
				}
			}
		}
	}
	if crd == nil {
		return nil
	}

	paths := map[apiextensionsv1.CustomResourceColumnDefinition]*jsonpath.JSONPath{}
	for _, column := range crd.AdditionalPrinterColumns {
		path := jsonpath.New(column.Name)
		if err := path.Parse(fmt.Sprintf("{%s}", column.JSONPath)); err != nil {
			klog.Warningf("invalid jsonnpath on kind '%s': %s", resource.Kind, err)
			continue
		}
		paths[column] = path
	}
	return paths
}

func (e *Apiextensions) resolvePath(path *jsonpath.JSONPath, object client.Object) (*string, error) {
	var data interface{}

	if j, err := e.Encoder.Encode(object); err != nil {
		return nil, err
	} else if err := json.Unmarshal(j, &data); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := path.Execute(&buf, data); err != nil {
		return nil, err
	}

	return ptr.To(buf.String()), nil
}
