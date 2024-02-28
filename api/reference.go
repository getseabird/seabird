package api

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectReference struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace,omitempty"`

	object client.Object
}

func (r *ObjectReference) GetObject(ctx context.Context, cluster *Cluster) (client.Object, error) {
	if r.object != nil {
		return r.object, nil
	}

	gvk := schema.FromAPIVersionAndKind(r.APIVersion, r.Kind).String()
	for key, t := range cluster.Scheme.AllKnownTypes() {
		if key.String() == gvk {
			r.object = reflect.New(t).Interface().(client.Object)
			break
		}
	}

	if err := cluster.Get(ctx, types.NamespacedName{Namespace: r.Namespace, Name: r.Name}, r.object); err != nil {
		return nil, err
	}

	return r.object, nil
}

func NewObjectReference(object client.Object) *ObjectReference {
	return &ObjectReference{
		APIVersion: object.GetObjectKind().GroupVersionKind().GroupKind().String(),
		Kind:       object.GetObjectKind().GroupVersionKind().Kind,
		Name:       object.GetName(),
		Namespace:  object.GetNamespace(),
		object:     object,
	}
}
