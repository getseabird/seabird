package cluster

import (
	"context"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	r "github.com/getseabird/seabird/internal/reactive"
	"github.com/getseabird/seabird/internal/ui/common"
)

type Window struct {
	r.BaseComponent[*Window]
	*adw.Application
	*common.ClusterState
}

func (c *Window) Update(ctx context.Context, message any) bool {
	switch message.(type) {
	default:
		return false
	}
}

func (c *Window) View(ctx context.Context) r.Model {
	return &r.AdwApplicationWindow{
		ApplicationWindow: r.ApplicationWindow{
			Application: &c.Application.Application,
			Window: r.Window{
				Title:         "Seabird",
				IconName:      "seabird",
				DefaultHeight: 700,
				DefaultWidth:  1000,
			},
		},
		Content: &r.AdwToastOverlay{
			Child: &r.AdwNavigationView{
				Pages: []r.AdwNavigationPage{
					r.AdwNavigationPage{
						Title: "Seabird",
						Child: &r.Box{
							Orientation: gtk.OrientationVertical,
							Children: []r.Model{
								&r.Box{
									Children: []r.Model{
										&r.Paned{
											StartChild: r.CreateComponent(&Navigation{resources: c.Resources, ClusterState: c.ClusterState}),
											EndChild:   r.CreateComponent(&Resources{ClusterState: c.ClusterState}),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (c *Window) On(hook r.Hook, widget gtk.Widgetter) {
	switch hook {
	case r.HookCreate:
		c.Application.ConnectActivate(func() {
			widget.(*adw.ApplicationWindow).Present()
		})
	}
}
