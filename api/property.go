package api

import (
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Property interface {
	GetID() string
}

type TextProperty struct {
	ID     string
	Name   string
	Value  string
	Source client.Object
}

func (p *TextProperty) GetID() string {
	if p.ID == "" {
		return strings.ToLower(p.Name)
	}
	return p.ID
}

type GroupProperty struct {
	ID       string
	Name     string
	Children []Property
}

func (p *GroupProperty) GetID() string {
	if p.ID == "" {
		return strings.ToLower(p.Name)
	}
	return p.ID
}
