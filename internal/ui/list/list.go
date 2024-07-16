package list

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	coreglib "github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/ui/editor"
	"github.com/getseabird/seabird/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type List struct {
	*adw.ToolbarView
	*common.ClusterState
	ctx         context.Context
	watchCancel context.CancelFunc
	model       *gtk.StringList
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
	l.model = gtk.NewStringList([]string{})
	l.columnView.SetModel(gtk.NewNoSelection(gtk.NewSortListModel(l.model, l.columnView.Sorter())))
	l.columnView.SetSingleClickActivate(true)
	l.columnView.SetMarginStart(16)
	l.columnView.SetMarginEnd(16)
	l.columnView.SetMarginBottom(16)

	l.columnView.ConnectActivate(func(position uint) {
		i, _ := strconv.Atoi(l.columnView.Model().Item(position).Cast().(*gtk.StringObject).String())
		obj := l.Objects.Value()[i]
		l.SelectedObject.Update(obj)
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

	common.OnChange(ctx, l.SelectedResource, l.onSelectedResourceChange)
	common.OnChange(ctx, l.Objects, l.onObjectsChange)
	common.OnChange(ctx, l.SearchFilter, l.onSearchFilterChange)

	filterNamespace := gio.NewSimpleAction("filterNamespace", glib.NewVariantType("s"))
	filterNamespace.ConnectActivate(func(parameter *glib.Variant) {
		text := strings.Trim(fmt.Sprintf("%s ns:%s", l.SearchText.Value(), parameter.String()), " ")
		l.SearchText.Update(text)
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
	l.model.Splice(0, l.model.NItems(), nil)

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
	for i, o := range objects {
		if !filter.Test(o) {
			continue
		}
		l.model.Append(strconv.Itoa(i))
	}

}

func (l *List) onSearchFilterChange(filter common.SearchFilter) {
	l.model.Splice(0, l.model.NItems(), nil)
	for i, object := range l.Objects.Value() {
		if filter.Test(object) {
			l.model.Append(strconv.Itoa(i))
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
		factory.ConnectBind(func(o *coreglib.Object) {
			cell := o.Cast().(*gtk.ColumnViewCell)
			idx, _ := strconv.Atoi(cell.Item().Cast().(*gtk.StringObject).String())
			object := l.Objects.Value()[idx]

			// Very fast resource switches (e.g. holding tab in the ui) can cause panics
			// This is a safeguard to make sure we don't send the wrong type
			// We should use the object as the model instead of the index once gotk supports subtyping
			gvks, _, _ := l.Cluster.Scheme.ObjectKinds(object)
			if len(gvks) == 1 {
				if gvks[0].String() != gvk {
					log.Printf("list bind error: expected '%s', got '%s'", gvk, gvks[0].String())
					return
				}
			}
			col.Bind(cell, object)
		})
		column := gtk.NewColumnViewColumn(col.Name, &factory.ListItemFactory)
		column.SetExpand(true)

		if col.Compare != nil {
			column.SetSorter(&gtk.NewCustomSorter(
				glib.NewObjectComparer[*gtk.StringObject](func(a, b *gtk.StringObject) int {
					ia, _ := strconv.Atoi(a.String())
					ib, _ := strconv.Atoi(b.String())
					return col.Compare(l.Objects.Value()[ia], l.Objects.Value()[ib])
				}),
			).Sorter)
		}

		gtkColumns = append(gtkColumns, column)
	}

	return gtkColumns
}
