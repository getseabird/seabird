package component

import (
	"context"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	r "github.com/getseabird/seabird/internal/reactive"
	"github.com/getseabird/seabird/internal/ui/common"
)

type Cluster struct {
	r.BaseComponent
	*adw.Application
	*common.ClusterState
}

func (c *Cluster) Init(ctx context.Context, ch chan<- any) {

}

func (c *Cluster) Update(ctx context.Context, message any, ch chan<- any) bool {
	switch message.(type) {
	default:
		return false
	}
}

func (c *Cluster) View(ctx context.Context, ch chan<- any) r.Model {
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
								&r.AdwHeaderBar{},
								&r.Box{
									Children: []r.Model{
										r.CreateComponent(&Navigation{resources: c.Resources}),
									},
								},
								&r.Label{Label: "foo"},
							},
						},
					},
				},
			},
		},
	}
}

func (c *Cluster) On(hook r.Hook, widget gtk.Widgetter) {
	switch hook {
	case r.HookCreate:
		c.Application.ConnectActivate(func() {
			widget.(*adw.ApplicationWindow).Present()
		})
	}
}
