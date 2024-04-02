package api

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func objectFromUnstructured(scheme *runtime.Scheme, gvk schema.GroupVersionKind, obj *unstructured.Unstructured) (client.Object, error) {
	target, err := scheme.New(gvk)
	if err != nil {
		return nil, err
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, target)
	return target.(client.Object), err
}
