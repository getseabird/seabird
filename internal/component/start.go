package component

import (
	"context"
	"log"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/component/cluster"
	"github.com/getseabird/seabird/internal/icon"
	"github.com/getseabird/seabird/internal/pubsub"
	r "github.com/getseabird/seabird/internal/reactive"
	"github.com/getseabird/seabird/internal/style"
	"github.com/getseabird/seabird/internal/ui/common"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

type Start struct {
	*adw.Application
	*common.State
	*common.ClusterState
	toast *r.Ref[*adw.ToastOverlay]
}

type clusterConnected *common.ClusterState

func (c *Start) Init(ctx context.Context) {
	c.toast = &r.Ref[*adw.ToastOverlay]{}

	switch style.Get() {
	case style.Darwin:
		gtk.SettingsGetDefault().SetObjectProperty("gtk-decoration-layout", "close,minimize,maximize")
	}

	if err := icon.Register(); err != nil {
		klog.Infof("failed to load icons: %v", err)
	}

	var err error
	c.State, err = common.NewState()
	if err != nil {
		log.Fatal(err.Error())
	}

	c.State.Preferences.Sub(ctx, func(p api.Preferences) {
		adw.StyleManagerGetDefault().SetColorScheme(adw.ColorScheme(p.ColorScheme))
	})
	style.Load()
}

func (c *Start) Update(ctx context.Context, message any) bool {
	switch message := message.(type) {
	case clusterConnected:
		c.ClusterState = message
		return true
	default:
		return false
	}
}

func (c *Start) View(ctx context.Context) r.Model {
	if c.ClusterState != nil {
		return r.CreateComponent(&cluster.Window{
			Application:  c.Application,
			ClusterState: c.ClusterState,
		})
	}

	return &r.AdwApplicationWindow{
		ApplicationWindow: r.ApplicationWindow{
			Application: &c.Application.Application,
			Window: r.Window{
				Title:         "Seabird",
				IconName:      "seabird",
				DefaultHeight: 600,
				DefaultWidth:  650,
			},
		},
		Content: &r.AdwToastOverlay{
			Ref: c.toast,
			Child: &r.AdwNavigationView{
				Pages: []r.AdwNavigationPage{
					r.AdwNavigationPage{
						Title: "Seabird",
						Child: &r.Box{
							Orientation: gtk.OrientationVertical,
							Children: []r.Model{
								&r.AdwHeaderBar{},
								&r.AdwPreferencesPage{
									Groups: []r.AdwPreferencesGroup{
										r.AdwPreferencesGroup{
											Title: "Clusters",
											Children: r.Map(c.Preferences.Value().Clusters, func(p pubsub.Property[api.ClusterPreferences]) r.Model {
												return r.CreateComponent(&ClusterConnectActionRow{Prefs: p, State: c.State, ctx: ctx})
											}),
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

func (c *Start) On(hook r.Hook, widget gtk.Widgetter) {
	switch hook {
	case r.HookCreate:
		c.Application.ConnectActivate(func() {
			widget.(*adw.ApplicationWindow).Present()
		})
	}
}

type ClusterConnectActionRow struct {
	r.BaseComponent[*ClusterConnectActionRow]
	*common.State
	Prefs   pubsub.Property[api.ClusterPreferences]
	loading bool
	ctx     context.Context
}

func (c *ClusterConnectActionRow) Update(ctx context.Context, message any) bool {
	switch message.(type) {
	default:
		return false
	}
}

func (c *ClusterConnectActionRow) View(ctx context.Context) r.Model {
	cluster := c.Prefs.Value()

	return &r.AdwActionRow{
		Activated: func(actionRow *adw.ActionRow) {
			c.SetState(ctx, func(component *ClusterConnectActionRow) {
				component.loading = true
			})
			go func() {
				state, err := c.NewClusterState(c.ctx, c.Prefs)
				c.SetState(ctx, func(component *ClusterConnectActionRow) {
					component.loading = false
				})
				if err != nil {
					glib.IdleAdd(func() {
						r.AddToast(ctx, adw.NewToast(err.Error()))
					})
					return
				}
				c.Broadcast(ctx, clusterConnected(state))
			}()
		},
		AdwPreferencesRow: r.AdwPreferencesRow{
			Title: cluster.Name,
			ListBoxRow: r.ListBoxRow{
				Activatable: true,
			},
		},
		Suffixes: []r.Model{
			&r.Label{
				Label:  ptr.Deref(cluster.Kubeconfig, api.Kubeconfig{}).Path,
				Widget: r.Widget{CSSClasses: []string{"dim-label"}},
			},
			&r.Spinner{Spinning: c.loading},
			r.Static(gtk.NewImageFromIconName("go-next-symbolic")),
			// r.If[r.Model](c.loading, &r.Spinner{Spinning: true}, r.Static(gtk.NewImageFromIconName("go-next-symbolic"))),
		},
	}
}
