package behavior

import (
	"context"
	"fmt"
	"sort"

	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/util"
	"github.com/imkira/go-observer/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DetailBehavior struct {
	*ClusterBehavior

	SelectedObject observer.Property[client.Object]
	Yaml           observer.Property[string]
	Properties     observer.Property[[]api.Property]
}

func (b *ClusterBehavior) NewRootDetailBehavior() *DetailBehavior {
	db := b.NewDetailBehavior()
	b.RootDetailBehavior = db
	return db
}

func (b *ClusterBehavior) NewDetailBehavior() *DetailBehavior {
	d := DetailBehavior{
		ClusterBehavior: b,
		SelectedObject:  observer.NewProperty[client.Object](nil),
		Yaml:            observer.NewProperty[string](""),
		Properties:      observer.NewProperty[[]api.Property](nil),
	}

	onChange(d.SelectedObject, d.onObjectChange)

	return &d
}

func (b *DetailBehavior) onObjectChange(object client.Object) {
	if object == nil {
		b.Properties.Update([]api.Property{})
		b.Yaml.Update("")
		return
	}

	codec := unstructured.NewJSONFallbackEncoder(serializer.NewCodecFactory(b.Scheme).LegacyCodec(b.Scheme.PreferredVersionAllGroups()...))
	objWithoutManagedFields := object.DeepCopyObject().(client.Object)
	objWithoutManagedFields.SetManagedFields(nil)
	encoded, err := runtime.Encode(codec, objWithoutManagedFields)
	if err != nil {
		b.Yaml.Update(fmt.Sprintf("error: %v", err))
	} else {
		yaml, err := util.JsonToYaml(encoded)
		if err != nil {
			b.Yaml.Update(fmt.Sprintf("error: %v", err))
		} else {
			b.Yaml.Update(string(yaml))
		}
	}

	var props []api.Property

	for _, ext := range b.Extensions {
		props = ext.CreateObjectProperties(object, props)
	}
	sort.Slice(props, func(i, j int) bool {
		return props[i].GetPriority() > props[j].GetPriority()
	})

	b.Properties.Update(props)
}

func (b *DetailBehavior) UpdateObject(obj *unstructured.Unstructured) error {
	m, err := b.RESTMapper.RESTMapping(obj.GetObjectKind().GroupVersionKind().GroupKind(), obj.GetObjectKind().GroupVersionKind().Version)
	if err != nil {
		return err
	}
	var iface dynamic.ResourceInterface = b.DynamicClient.Resource(m.Resource)
	if len(obj.GetNamespace()) > 0 {
		iface = iface.(dynamic.NamespaceableResourceInterface).Namespace(obj.GetNamespace())
	}
	_, err = iface.Update(context.TODO(), obj, metav1.UpdateOptions{})
	return err
}

func (b *DetailBehavior) DeleteObject(obj client.Object) error {
	return b.Delete(context.TODO(), obj)
}
