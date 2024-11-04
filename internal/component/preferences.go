package component

import (
	"context"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/pubsub"
	r "github.com/getseabird/seabird/internal/reactive"
	"github.com/getseabird/seabird/internal/ui/common"
	"k8s.io/utils/ptr"
)

type colorScheme adw.ColorScheme

type Preferences struct {
	*common.State
}

func (c *Preferences) Update(ctx context.Context, message any, ch chan<- any) bool {
	switch message := message.(type) {
	case colorScheme:
		prefs := c.Preferences.Value()
		prefs.ColorScheme = adw.ColorScheme(message)
		if prefs.ColorScheme == adw.ColorSchemePreferLight {
			prefs.ColorScheme = adw.ColorSchemeForceDark
		}
		c.Preferences.Pub(prefs)
		return false
	default:
		return false
	}
}

func (c *Preferences) View(ctx context.Context, ch chan<- any) r.Model {
	var stack *adw.ViewStack

	return &r.AdwPreferencesWindow{
		AdwWindow: r.AdwWindow{
			Content: &r.AdwNavigationView{
				Pages: []r.AdwNavigationPage{
					r.AdwNavigationPage{
						Title: "Preferences",
						Child: &r.Box{
							Children: []r.Model{
								&r.AdwHeaderBar{
									TitleWidget: &r.AdwViewSwitcher{
										ViewStack: stack,
										Policy:    adw.ViewSwitcherPolicyWide,
									},
								},
								&r.AdwViewStack{
									Ref: r.Ref[*adw.ViewStack]{Ref: stack},
									Pages: []r.AdwViewStackPage{
										r.AdwViewStackPage{
											Name:     "general",
											Title:    "General",
											IconName: "settings-symbolic",
											Child: &r.AdwPreferencesPage{
												Groups: []r.AdwPreferencesGroup{
													r.AdwPreferencesGroup{
														Children: []r.Model{
															&r.AdwComboRow{
																Widget: r.Widget{
																	Signals: map[string]any{
																		"notify::selected-item": func(row *adw.ComboRow) {
																			ch <- colorScheme(row.Selected())
																		},
																	},
																},
																AdwPreferencesRow: r.AdwPreferencesRow{
																	Title: "Color Scheme",
																},
																Model: gtk.NewStringList([]string{"Default", "Light", "Dark"}),
															},
														},
													},
													r.AdwPreferencesGroup{
														Title: "Clusters",
														HeaderSuffix: &r.Button{
															Widget: r.Widget{
																CSSClasses: []string{"flat"},
															},
															IconName: "plus-symbolic",
														},
														Children: r.Map(c.Preferences.Value().Clusters, func(c pubsub.Property[api.ClusterPreferences]) r.Model {
															cluster := c.Value()
															return &r.AdwActionRow{
																// Activatable: true,

																Suffixes: []r.Model{
																	&r.Label{
																		Label:  ptr.Deref(cluster.Kubeconfig, api.Kubeconfig{}).Path,
																		Widget: r.Widget{CSSClasses: []string{"dim-label"}},
																	},
																	r.Static(gtk.NewImageFromIconName("go-next-symbolic")),
																},
															}
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
				},
			},
		},
	}
}
