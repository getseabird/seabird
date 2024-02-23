package api

import (
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Property interface {
	GetID() string
	GetPriority() int8
}

type TextProperty struct {
	ID     string
	Name   string
	Value  string
	Source client.Object
	Widget func(gtk.Widgetter, *adw.NavigationView)
}

func (p *TextProperty) GetID() string {
	if p.ID == "" {
		return strings.ToLower(p.Name)
	}
	return p.ID
}

func (p *TextProperty) GetPriority() int8 {
	return 0
}

type GroupProperty struct {
	ID       string
	Priority int8
	Name     string
	Children []Property
	Widget   func(gtk.Widgetter, *adw.NavigationView)
}

func (p *GroupProperty) GetID() string {
	if p.ID == "" {
		return strings.ToLower(p.Name)
	}
	return p.ID
}

func (p *GroupProperty) GetPriority() int8 {
	return p.Priority
}
