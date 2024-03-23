package ui

import (
	"context"
	"log"
	"sort"
	"strconv"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/util"
	"github.com/imkira/go-observer/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ListView struct {
	*gtk.Box
	*common.ClusterState
	ctx         context.Context
	watchCancel context.CancelFunc
	objects     observer.Property[[]client.Object]
	model       *gtk.StringList
	columnView  *gtk.ColumnView
	columns     []*gtk.ColumnViewColumn
	columnType  *metav1.APIResource
	splitView   *adw.OverlaySplitView
}

func NewListView(ctx context.Context, state *common.ClusterState, header gtk.Widgetter, splitView *adw.OverlaySplitView) *ListView {
	l := ListView{
		ClusterState: state,
		ctx:          ctx,
		Box:          gtk.NewBox(gtk.OrientationVertical, 0),
		objects:      observer.NewProperty[[]client.Object](nil),
		splitView:    splitView,
	}
	l.AddCSSClass("view")
	l.Append(header)

	l.columnView = gtk.NewColumnView(nil)
	l.model = gtk.NewStringList([]string{})
	l.columnView.SetModel(gtk.NewNoSelection(gtk.NewSortListModel(l.model, l.columnView.Sorter())))
	l.columnView.SetSingleClickActivate(true)
	l.columnView.SetMarginStart(16)
	l.columnView.SetMarginEnd(16)
	l.columnView.SetMarginBottom(16)

	l.columnView.ConnectActivate(func(position uint) {
		i, _ := strconv.Atoi(l.model.Item(position).Cast().(*gtk.StringObject).String())
		obj := l.objects.Value()[i]
		l.SelectedObject.Update(obj)
		l.splitView.SetShowSidebar(true)
	})

	sw := gtk.NewScrolledWindow()
	sw.SetHExpand(true)
	sw.SetVExpand(true)
	sw.SetSizeRequest(400, 0)
	vp := gtk.NewViewport(nil, nil)
	vp.SetChild(l.columnView)
	sw.SetChild(vp)
	l.Append(sw)

	common.OnChange(ctx, l.SelectedResource, l.onSelectedResourceChange)
	common.OnChange(ctx, l.objects, l.onObjectsChange)
	common.OnChange(ctx, l.SearchFilter, l.onSearchFilterChange)

	return &l
}

func (l *ListView) onSelectedResourceChange(resource *metav1.APIResource) {
	if resource == nil {
		return
	}
	if l.watchCancel != nil {
		l.watchCancel()
	}
	var ctx context.Context
	ctx, l.watchCancel = context.WithCancel(l.ctx)
	api.ObjectWatcher(ctx, resource, l.objects)
	l.splitView.SetShowSidebar(false)
}

func (l *ListView) onObjectsChange(objects []client.Object) {
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

func (l *ListView) onSearchFilterChange(filter common.SearchFilter) {
	l.model.Splice(0, l.model.NItems(), nil)
	for i, object := range l.objects.Value() {
		if filter.Test(object) {
			l.model.Append(strconv.Itoa(i))
		}
	}
}

func (l *ListView) createColumns() []*gtk.ColumnViewColumn {
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
		gvk := util.ResourceGVK(l.SelectedResource.Value()).String()
		factory.ConnectBind(func(listitem *gtk.ListItem) {
			idx, _ := strconv.Atoi(listitem.Item().Cast().(*gtk.StringObject).String())
			object := l.objects.Value()[idx]

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
			col.Bind(listitem, object)
		})
		column := gtk.NewColumnViewColumn(col.Name, &factory.ListItemFactory)
		column.SetExpand(true)

		if col.Compare != nil {
			column.SetSorter(&gtk.NewCustomSorter(
				glib.NewObjectComparer[*gtk.StringObject](func(a, b *gtk.StringObject) int {
					ia, _ := strconv.Atoi(a.String())
					ib, _ := strconv.Atoi(b.String())
					return col.Compare(l.objects.Value()[ia], l.objects.Value()[ib])
				}),
			).Sorter)
		}

		gtkColumns = append(gtkColumns, column)
	}

	return gtkColumns
}
