package api

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (c *Cluster) GVKToR(gvk schema.GroupVersionKind) (*schema.GroupVersionResource, error) {
	m, err := c.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	return &m.Resource, nil
}

func (c *Cluster) GVRToK(gvr schema.GroupVersionResource) (*schema.GroupVersionKind, error) {
	kinds, err := c.RESTMapper.KindsFor(gvr)
	if err != nil {
		return nil, err
	}

	if len(kinds) == 0 {
		return nil, fmt.Errorf("%v not found", gvr.String())
	}

	return &kinds[0], nil
}
