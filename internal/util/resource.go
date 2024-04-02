package util

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func GVRForResource(resource *v1.APIResource) schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: resource.Group, Version: resource.Version, Resource: resource.Name}
}

func GVKForResource(resource *v1.APIResource) schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: resource.Group, Version: resource.Version, Kind: resource.Kind}
}

func ResourceEquals(r1, r2 *v1.APIResource) bool {
	if r1 == nil || r2 == nil {
		return false
	}
	return r1.Group == r2.Group && r1.Version == r2.Version && r1.Name == r2.Name
}

func GVREquals(r1, r2 schema.GroupVersionResource) bool {
	return r1.Group == r2.Group && r1.Version == r2.Version && r1.Resource == r2.Resource
}

func GVKEquals(r1, r2 schema.GroupVersionKind) bool {
	return r1.Group == r2.Group && r1.Version == r2.Version && r1.Kind == r2.Kind
}
