package main

import (
	"context"
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	r "github.com/getseabird/seabird/internal/reactive"
)

type Increment struct{}

type SampleComponent struct {
	r.BaseComponent[*SampleComponent]
	counter int
}

func (c *SampleComponent) Update(ctx context.Context, message any) bool {
	switch message.(type) {
	case Increment:
		c.counter++
		return true
	default:
		return false
	}
}

func (c *SampleComponent) View(ctx context.Context) r.Model {
	return &r.Box{
		Orientation: gtk.OrientationVertical,
		Spacing:     5,
		Children: []r.Model{
			&r.AdwHeaderBar{},
			&r.AdwBin{
				Child: &r.Label{
					Label: fmt.Sprintf("Clicked %d times", c.counter),
				},
			},
			&r.Button{
				Label: "Click me",
				Clicked: func(button *gtk.Button) {
					c.Broadcast(ctx, Increment{})
				},
			},
		},
	}
}

func (c *SampleComponent) On(hook r.Hook, widget gtk.Widgetter) {

}
