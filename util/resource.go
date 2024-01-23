package util

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func ResourceGVR(resource *v1.APIResource) schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: resource.Group, Version: resource.Version, Resource: resource.Name}
}

func ResourceEquals(r1, r2 *v1.APIResource) bool {
	return r1.Group == r2.Group && r1.Version == r2.Version && r1.Name == r2.Name
}
