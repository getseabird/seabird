package cluster

import (
	"context"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/getseabird/seabird/internal/icon"
	r "github.com/getseabird/seabird/internal/reactive"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Navigation struct {
	r.BaseComponent[*Navigation]
	resources []metav1.APIResource
	*common.ClusterState
}

func (c *Navigation) Update(ctx context.Context, message any) bool {
	switch message.(type) {
	default:
		return false
	}
}

func (c *Navigation) View(ctx context.Context) r.Model {
	return &r.AdwToolbarView{
		Widget: r.Widget{
			CSSClasses: []string{"flat", "view"},
		},
		TopBars: []r.Model{
			&r.AdwHeaderBar{
				ShowEndTitleButtons: false,
			},
		},
		Content: &r.Box{
			Orientation: gtk.OrientationVertical,
			Children: []r.Model{
				&r.Box{
					Orientation: gtk.OrientationHorizontal,
					Children: []r.Model{
						&r.ToggleButton{
							Button: r.Button{
								IconName: "view-list-symbolic",
							},
						},
						&r.ToggleButton{
							Button: r.Button{
								IconName: "pin-symbolic",
							},
						},
					},
				},
				&r.ScrolledWindow{
					Widget: r.Widget{
						VExpand: true,
						HExpand: true,
					},
					Child: &r.ListBox{
						RowSelected: func(listBox *gtk.ListBox, listBoxRow *gtk.ListBoxRow) {
							res := c.resources[listBoxRow.Index()]
							c.SelectedResource.Pub(&res)
						},
						Widget: r.Widget{
							CSSClasses: []string{"navigation-sidebar"},
						},
						Children: r.Map(c.resources, func(res metav1.APIResource) r.Model {
							return &r.ListBoxRow{
								Child: &r.Box{
									Spacing: 8,
									Widget: r.Widget{
										Margin: [4]int{4, 4, 4, 4},
									},
									Children: []r.Model{
										r.Static(icon.Kind(util.GVKForResource(&res))),
										&r.Box{
											Orientation: gtk.OrientationVertical,
											Spacing:     2,
											Widget: r.Widget{
												HExpand: true,
											},
											Children: []r.Model{
												&r.Label{
													Widget: r.Widget{
														HAlign: gtk.AlignStart,
													},
													Label:     res.Kind,
													Ellipsize: pango.EllipsizeEnd,
												},
												&r.Label{
													Widget: r.Widget{
														HAlign:     gtk.AlignStart,
														CSSClasses: []string{"caption", "dim-label"},
													},
													Label:     r.If(res.Group == "", "k8s.io", res.Group),
													Ellipsize: pango.EllipsizeEnd,
												},
											},
										},
										&r.Box{
											Widget: r.Widget{
												HAlign: gtk.AlignEnd,
												VAlign: gtk.AlignCenter,
											},
											Children: []r.Model{
												&r.Label{
													Widget: r.Widget{
														CSSClasses: []string{"success", "pill"},
													},
													Label: "4",
												},
											},
										},
									},
								},
							}
						}),
					},
				},
			},
		},
	}
}

// func (n *Navigation) createResourceRow(resource *metav1.APIResource, idx int, fav bool) *gtk.ListBoxRow {
// 	gvr := util.GVRForResource(resource)

// 	row := gtk.NewListBoxRow()
// 	json, err := json.Marshal(gvr)
// 	if err != nil {
// 		panic(err)
// 	}
// 	row.SetName(string(json))
// 	box := gtk.NewBox(gtk.OrientationHorizontal, 8)
// 	box.SetMarginTop(4)
// 	box.SetMarginBottom(4)
// 	box.Append(icon.Kind(util.GVKForResource(resource)))
// 	vbox := gtk.NewBox(gtk.OrientationVertical, 2)
// 	vbox.SetVAlign(gtk.AlignCenter)
// 	box.Append(vbox)
// 	label := gtk.NewLabel(resource.Kind)
// 	label.SetHAlign(gtk.AlignStart)
// 	label.SetEllipsize(pango.EllipsizeEnd)
// 	vbox.Append(label)
// 	label = gtk.NewLabel(resource.Group)
// 	if resource.Group == "" {
// 		label.SetText("k8s.io")
// 	}
// 	label.SetHAlign(gtk.AlignStart)
// 	label.AddCSSClass("caption")
// 	label.AddCSSClass("dim-label")
// 	label.SetEllipsize(pango.EllipsizeEnd)
// 	vbox.Append(label)
// 	row.SetChild(box)

// 	statusBox := gtk.NewBox(gtk.OrientationHorizontal, 4)
// 	statusBox.SetHExpand(true)
// 	statusBox.SetHAlign(gtk.AlignEnd)
// 	statusBox.SetVAlign(gtk.AlignCenter)
// 	box.Append(statusBox)

// 	errorLabel := gtk.NewLabel("")
// 	errorLabel.AddCSSClass("warning")
// 	errorLabel.AddCSSClass("pill")
// 	errorLabel.SetVisible(false)
// 	statusBox.Append(errorLabel)
// 	readyLabel := gtk.NewLabel("")
// 	readyLabel.AddCSSClass("success")
// 	readyLabel.AddCSSClass("pill")
// 	readyLabel.SetVisible(false)
// 	statusBox.Append(readyLabel)

// 	if fav && n.Scheme.IsGroupRegistered(resource.Group) && slices.Contains(resource.Verbs, "watch") {
// 		go func() {
// 			informer := n.Cluster.GetInformer(util.GVRForResource(resource))
// 			h := bindStatusCount(n.ctx, informer, func(m map[api.StatusType]int) {
// 				glib.IdleAdd(func() {
// 					readys := m[api.StatusSuccess]
// 					readyLabel.SetVisible(readys > 0)
// 					readyLabel.SetText(fmt.Sprintf("%d", readys))
// 					errors := m[api.StatusError] + m[api.StatusWarning]
// 					errorLabel.SetVisible(errors > 0)
// 					errorLabel.SetText(fmt.Sprintf("%d", errors))
// 				})
// 			})
// 			if h != nil {
// 				glib.IdleAdd(func() {
// 					n.informerRegistrations[informer] = h
// 				})
// 			}
// 		}()
// 	}

// 	gesture := gtk.NewGestureClick()
// 	gesture.SetButton(gdk.BUTTON_SECONDARY)
// 	gesture.ConnectPressed(func(nPress int, x, y float64) {
// 		menu := gio.NewMenu()
// 		if fav {
// 			menu.Append("Move to Resources", fmt.Sprintf("navigation.unpin('%d')", idx))
// 		} else {
// 			menu.Append("Move to Favourites", fmt.Sprintf("navigation.pin('%d')", idx))
// 		}
// 		popover := gtk.NewPopoverMenuFromModel(menu)
// 		popover.SetHasArrow(false)
// 		row.FirstChild().(*gtk.Box).Append(popover)
// 		popover.SetVisible(true)
// 	})
// 	row.AddController(gesture)

// 	return row
// }
