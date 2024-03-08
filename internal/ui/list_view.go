package ui

import (
	"context"
	"log"
	"sort"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/behavior"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/util"
	"github.com/imkira/go-observer/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ListView struct {
	*gtk.Box
	ctx         context.Context
	watchCancel context.CancelFunc
	objects     observer.Property[[]client.Object]
	behavior    *behavior.ClusterBehavior
	selection   *gtk.SingleSelection
	columnView  *gtk.ColumnView
	columns     []*gtk.ColumnViewColumn
	columnType  *metav1.APIResource
	selected    types.UID
}

func NewListView(ctx context.Context, behavior *behavior.ClusterBehavior, header gtk.Widgetter) *ListView {
	l := ListView{
		ctx:      ctx,
		behavior: behavior,
		Box:      gtk.NewBox(gtk.OrientationVertical, 0),
		objects:  observer.NewProperty[[]client.Object](nil),
	}
	l.AddCSSClass("view")
	l.Append(header)

	l.columnView = gtk.NewColumnView(nil)
	l.selection = l.createModel()
	l.columnView.SetModel(l.selection)
	l.columnView.SetMarginStart(16)
	l.columnView.SetMarginEnd(16)
	l.columnView.SetMarginBottom(16)

	sw := gtk.NewScrolledWindow()
	sw.SetHExpand(true)
	sw.SetVExpand(true)
	sw.SetSizeRequest(400, 0)
	vp := gtk.NewViewport(nil, nil)
	vp.SetChild(l.columnView)
	sw.SetChild(vp)
	l.Append(sw)

	common.OnChange(ctx, behavior.SelectedResource, l.onSelectedResourceChange)
	common.OnChange(ctx, l.objects, l.onObjectsChange)
	common.OnChange(ctx, behavior.SearchFilter, l.onSearchFilterChange)

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
}

func (l *ListView) onObjectsChange(objects []client.Object) {
	resource := l.behavior.SelectedResource.Value()
	if resource == nil {
		return
	}
	list := l.getModel()
	list.Splice(0, list.NItems(), nil)

	if l.columnType == nil || !util.ResourceEquals(l.columnType, resource) {
		l.columnType = resource

		l.selection = l.createModel()
		l.columnView.SetModel(l.selection)

		for _, column := range l.columns {
			l.columnView.RemoveColumn(column)
		}
		l.columns = l.createColumns()
		for _, column := range l.columns {
			l.columnView.AppendColumn(column)
		}
	}

	filter := l.behavior.SearchFilter.Value()
	for i, o := range objects {
		if !filter.Test(o) {
			continue
		}
		l.getModel().Append(strconv.Itoa(i))
		if o.GetUID() == l.selected {
			l.selection.SetSelected(uint(i))
		}
	}

	if len(objects) > 0 {
		if selected := l.selection.Selected(); selected == gtk.InvalidListPosition {
			l.selection.SetSelected(0)
			l.behavior.RootDetailBehavior.SelectedObject.Update(objects[0])
		} else {
			i, _ := strconv.Atoi(l.selection.ListModel.Item(selected).Cast().(*gtk.StringObject).String())
			l.behavior.RootDetailBehavior.SelectedObject.Update(objects[i])
		}
	} else {
		l.behavior.RootDetailBehavior.SelectedObject.Update(nil)
	}
}

func (l *ListView) onSearchFilterChange(filter behavior.SearchFilter) {
	list := l.getModel()
	list.Splice(0, list.NItems(), nil)
	l.selection.SetSelected(gtk.InvalidListPosition)
	for i, object := range l.objects.Value() {
		if filter.Test(object) {
			list.Append(strconv.Itoa(i))
		}
		if object.GetUID() == l.selected {
			l.selection.SetSelected(uint(i))
		}
	}
	if list.NItems() > 0 && l.selection.Selected() == gtk.InvalidListPosition {
		l.selection.SetSelected(0)
	}
	if l.selection.Selected() != gtk.InvalidListPosition {
		// SelectionChanged isn't triggered when calling SetSelected
		l.selection.SelectionChanged(uint(l.selection.Selected()), 1)
	} else {
		l.behavior.RootDetailBehavior.SelectedObject.Update(nil)
	}
}

func (l *ListView) createColumns() []*gtk.ColumnViewColumn {
	var columns []api.Column

	for _, e := range l.behavior.Extensions {
		columns = e.CreateColumns(l.ctx, l.behavior.SelectedResource.Value(), columns)
	}
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Priority > columns[j].Priority
	})

	var gtkColumns []*gtk.ColumnViewColumn
	for _, col := range columns {
		factory := gtk.NewSignalListItemFactory()
		gvk := util.ResourceGVK(l.behavior.SelectedResource.Value()).String()
		factory.ConnectBind(func(listitem *gtk.ListItem) {
			idx, _ := strconv.Atoi(listitem.Item().Cast().(*gtk.StringObject).String())
			object := l.objects.Value()[idx]

			// Very fast resource switches (e.g. holding tab in the ui) can cause panics
			// This is a safeguard to make sure we don't send the wrong type
			// We should use the object as the model instead of the index once gotk supports subtyping
			gvks, _, _ := l.behavior.Cluster.Scheme.ObjectKinds(object)
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

func (l *ListView) createModel() *gtk.SingleSelection {
	model := gtk.NewSortListModel(gtk.NewStringList([]string{}), l.columnView.Sorter())
	selection := gtk.NewSingleSelection(model)
	selection.ConnectSelectionChanged(func(_, _ uint) {
		selected := l.selection.Selected()
		if selected == gtk.InvalidListPosition {
			return
		}
		i, _ := strconv.Atoi(l.selection.ListModel.Item(selected).Cast().(*gtk.StringObject).String())
		obj := l.objects.Value()[i]
		l.selected = obj.GetUID()
		l.behavior.RootDetailBehavior.SelectedObject.Update(obj)
	})

	return selection
}

func (l *ListView) getModel() *gtk.StringList {
	return l.selection.Model().Cast().(*gtk.SortListModel).Model().Cast().(*gtk.StringList)
}
