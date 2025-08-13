package cluster

import (
	"context"
	"fmt"
	"sort"

	"github.com/diamondburned/gotk4/pkg/core/gioutil"
	coreglib "github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	r "github.com/getseabird/seabird/internal/reactive"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resources struct {
	r.BaseComponent[*Resources]
	*common.ClusterState
	model *gioutil.ListModel[client.Object]
}

func (c *Resources) Init(ctx context.Context) {
	c.model = gioutil.NewListModel[client.Object]()

	var cancel context.CancelFunc
	c.SelectedResource.Sub(ctx, func(a *metav1.APIResource) {
		var subctx context.Context
		if cancel != nil {
			cancel()
		}
		subctx, cancel = context.WithCancel(ctx)
		if err := api.InformerConnectProperty(subctx, c.Cluster, util.GVRForResource(c.SelectedResource.Value()), c.Objects); err != nil {
			klog.Error(err.Error())
		}
	})

	c.Objects.Sub(ctx, func(objects []client.Object) {
		c.model.Splice(0, int(c.model.NItems()))
		// filter := c.SearchFilter.Value()
		for _, o := range objects {
			// if !filter.Test(o) {
			// 	continue
			// }
			c.model.Append(o)
		}
		// c.SetState(ctx, func(component *Resources) {})
	})

}

func (c *Resources) Update(ctx context.Context, message any) bool {
	switch message.(type) {
	default:
		return false
	}
}

func (c *Resources) View(ctx context.Context) r.Model {
	var columns []api.Column

	for _, e := range c.Extensions {
		columns = e.CreateColumns(ctx, c.SelectedResource.Value(), columns)
	}
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Priority > columns[j].Priority
	})

	var gtkColumns []*gtk.ColumnViewColumn
	for _, col := range columns {
		factory := gtk.NewSignalListItemFactory()
		gvk := util.GVKForResource(c.SelectedResource.Value()).String()
		factory.ConnectBind(func(ce *coreglib.Object) {
			cell := ce.Cast().(*gtk.ColumnViewCell)
			object := gioutil.ObjectValue[client.Object](cell.Item())

			// Very fast resource switches (e.g. holding tab in the ui) can cause panics
			// This is a safeguard to make sure we don't send the wrong type
			// We should use the object as the model instead of the index once gotk supports subtyping
			gvks, _, _ := c.Cluster.Scheme.ObjectKinds(object)
			if len(gvks) == 1 {
				if gvks[0].String() != gvk {
					klog.Infof("list bind error: expected '%s', got '%s'", gvk, gvks[0].String())
					return
				}
			}
			col.Bind(api.Cell{ColumnViewCell: cell}, object)
		})

		column := gtk.NewColumnViewColumn(col.Name, &factory.ListItemFactory)
		column.SetExpand(true)
		column.SetResizable(true)
		column.SetID(fmt.Sprintf("%v %v", c.SelectedResource.Value().String(), col.Name))

		if col.Compare != nil {
			column.SetSorter(&gtk.NewCustomSorter(
				glib.NewObjectComparer(func(a, b *coreglib.Object) int {
					return col.Compare(gioutil.ObjectValue[client.Object](a), gioutil.ObjectValue[client.Object](b))
				}),
			).Sorter)
		}

		gtkColumns = append(gtkColumns, column)
	}

	return &r.Box{
		Widget: r.Widget{
			CSSClasses: []string{"view"},
		},
		Orientation: gtk.OrientationVertical,
		Children: []r.Model{
			&r.AdwHeaderBar{
				Widget: r.Widget{
					CSSClasses: []string{"flat"},
				},
				ShowStartTitleButtons: ptr.To(false),
				Start: []r.Model{
					&r.Button{
						Widget: r.Widget{
							TooltipText: "New Resource",
						},
						IconName: "document-new-symbolic",
					},
				},
				TitleWidget: &r.Box{
					Widget: r.Widget{
						CSSClasses: []string{"linked"},
						Margin:     [4]int{0, 32, 0, 32},
					},
					Children: []r.Model{
						&r.SearchEntry{
							Widget: r.Widget{
								HExpand: true,
							},
							PlaceholderText: "Search",
						},
						&r.Button{
							Widget: r.Widget{
								TooltipText: "Filter",
							},
							IconName: "funnel-symbolic",
						},
					},
				},
			},
			&r.ScrolledWindow{
				Widget: r.Widget{
					VExpand: true,
				},
				Child: &r.ColumnView{
					Widget: r.Widget{
						Margin: [4]int{0, 8, 0, 8},
					},
					Columns: gtkColumns,
				},
			},
		},
	}

}

func (c *Resources) On(hook r.Hook, widget gtk.Widgetter) {
	switch hook {
	case r.HookCreate:
		cv := widget.(*gtk.Box).LastChild().(*gtk.ScrolledWindow).Child().(*gtk.ColumnView)
		cv.SetModel(gtk.NewNoSelection(gtk.NewSortListModel(c.model, cv.Sorter())))
	}
}
