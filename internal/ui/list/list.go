package list

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/core/gioutil"
	coreglib "github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/ui/editor"
	"github.com/getseabird/seabird/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type List struct {
	*adw.ToolbarView
	*common.ClusterState
	ctx         context.Context
	watchCancel context.CancelFunc
	model       *gioutil.ListModel[client.Object]
	sortModel   *gtk.SortListModel
	columnView  *gtk.ColumnView
	columns     []*gtk.ColumnViewColumn
	columnType  *metav1.APIResource
	dialog      *adw.Dialog
}

func NewList(ctx context.Context, state *common.ClusterState, dialog *adw.Dialog, editor *editor.EditorWindow) *List {
	l := List{
		ToolbarView:  adw.NewToolbarView(),
		ClusterState: state,
		ctx:          ctx,
		dialog:       dialog,
	}

	l.AddCSSClass("view")
	l.AddTopBar(newListHeader(ctx, state, editor))

	l.columnView = gtk.NewColumnView(nil)
	l.model = gioutil.NewListModel[client.Object]()
	l.sortModel = gtk.NewSortListModel(l.model, l.columnView.Sorter())
	l.columnView.SetModel(gtk.NewNoSelection(l.sortModel))
	l.columnView.SetSingleClickActivate(true)
	l.columnView.SetMarginStart(16)
	l.columnView.SetMarginEnd(16)
	l.columnView.SetMarginBottom(16)

	l.columnView.ConnectActivate(func(position uint) {
		obj := gioutil.ObjectValue[client.Object](l.columnView.Model().Item(position))
		l.SelectedObject.Pub(obj)
		l.dialog.Present(l)
	})

	sw := gtk.NewScrolledWindow()
	sw.SetHExpand(true)
	sw.SetVExpand(true)
	sw.SetSizeRequest(400, 0)
	vp := gtk.NewViewport(nil, nil)
	vp.SetChild(l.columnView)
	sw.SetChild(vp)
	l.SetContent(sw)

	l.SelectedResource.Sub(ctx, l.onSelectedResourceChange)
	l.Objects.Sub(ctx, l.onObjectsChange)
	l.SearchFilter.Sub(ctx, l.onSearchFilterChange)

	filterNamespace := gio.NewSimpleAction("filterNamespace", glib.NewVariantType("s"))
	filterNamespace.ConnectActivate(func(parameter *glib.Variant) {
		text := strings.Trim(fmt.Sprintf("%s ns:%s", l.SearchText.Value(), parameter.String()), " ")
		l.SearchText.Pub(text)
	})
	actionGroup := gio.NewSimpleActionGroup()
	actionGroup.AddAction(filterNamespace)
	l.InsertActionGroup("list", actionGroup)

	return &l
}

func (l *List) onSelectedResourceChange(resource *metav1.APIResource) {
	if resource == nil {
		return
	}
	if l.watchCancel != nil {
		l.watchCancel()
	}
	var ctx context.Context
	ctx, l.watchCancel = context.WithCancel(l.ctx)
	api.InformerConnectProperty(ctx, l.Cluster, util.GVRForResource(resource), l.Objects)
}

func (l *List) onObjectsChange(objects []client.Object) {
	resource := l.SelectedResource.Value()
	if resource == nil {
		return
	}
	l.model.Splice(0, int(l.model.NItems()))

	if l.columnType == nil || !util.ResourceEquals(l.columnType, resource) {
		l.columnType = resource

		for _, column := range l.columns {
			l.columnView.RemoveColumn(column)
		}
		l.columns = l.createColumns()
		for _, column := range l.columns {
			l.columnView.AppendColumn(column)
		}
	}

	filter := l.SearchFilter.Value()
	for _, o := range objects {
		if !filter.Test(o) {
			continue
		}
		l.model.Append(o)
	}

}

func (l *List) onSearchFilterChange(filter common.SearchFilter) {
	l.model.Splice(0, int(l.model.NItems()))
	for _, object := range l.Objects.Value() {
		if filter.Test(object) {
			l.model.Append(object)
		}
	}
}

func (l *List) createColumns() []*gtk.ColumnViewColumn {
	var columns []api.Column

	for _, e := range l.Extensions {
		columns = e.CreateColumns(l.ctx, l.SelectedResource.Value(), columns)
	}
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Priority > columns[j].Priority
	})

	var gtkColumns []*gtk.ColumnViewColumn
	for _, col := range columns {
		factory := gtk.NewSignalListItemFactory()
		gvk := util.GVKForResource(l.SelectedResource.Value()).String()
		factory.ConnectBind(func(c *coreglib.Object) {
			cell := c.Cast().(*gtk.ColumnViewCell)
			object := gioutil.ObjectValue[client.Object](cell.Item())

			// Very fast resource switches (e.g. holding tab in the ui) can cause panics
			// This is a safeguard to make sure we don't send the wrong type
			// We should use the object as the model instead of the index once gotk supports subtyping
			gvks, _, _ := l.Cluster.Scheme.ObjectKinds(object)
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

		if col.Compare != nil {
			column.SetSorter(&gtk.NewCustomSorter(
				glib.NewObjectComparer(func(a, b *coreglib.Object) int {
					return col.Compare(gioutil.ObjectValue[client.Object](a), gioutil.ObjectValue[client.Object](b))
				}),
			).Sorter)
		}

		gtkColumns = append(gtkColumns, column)
	}

	return gtkColumns
}
