package ui

import (
	"sort"
	"strconv"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/behavior"
	"github.com/getseabird/seabird/internal/util"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const GtkInvalidListPosition uint = 4294967295

type ListView struct {
	*gtk.Box
	behavior   *behavior.ListBehavior
	parent     *gtk.Window
	selection  *gtk.SingleSelection
	columnView *gtk.ColumnView
	columns    []*gtk.ColumnViewColumn
	columnType *v1.APIResource
	objects    []client.Object
	selected   types.UID
}

func NewListView(parent *gtk.Window, behavior *behavior.ListBehavior) *ListView {
	l := ListView{
		Box:      gtk.NewBox(gtk.OrientationVertical, 0),
		parent:   parent,
		behavior: behavior,
	}
	l.AddCSSClass("view")

	header := adw.NewHeaderBar()
	header.AddCSSClass("flat")
	header.SetShowEndTitleButtons(false)
	header.SetShowStartTitleButtons(false)
	header.SetTitleWidget(NewListHeader(behavior))
	l.Append(header)

	l.selection = l.createModel()
	l.columnView = gtk.NewColumnView(l.selection)
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

	onChange(l.behavior.Objects, l.onObjectsChange)
	onChange(l.behavior.SearchFilter, l.onSearchFilterChange)

	return &l
}

func (l *ListView) onObjectsChange(objects []client.Object) {
	l.objects = objects

	if l.columnType == nil || !util.ResourceEquals(l.columnType, l.behavior.SelectedResource.Value()) {
		l.columnType = l.behavior.SelectedResource.Value()

		l.selection = l.createModel()
		l.columnView.SetModel(l.selection)

		for _, column := range l.columns {
			l.columnView.RemoveColumn(column)
		}
		l.columns = l.createColumns()
		for _, column := range l.columns {
			l.columnView.AppendColumn(column)
		}
	} else {
		list := l.selection.Model().Cast().(*gtk.StringList)
		list.Splice(0, list.NItems(), nil)
	}

	filter := l.behavior.SearchFilter.Value()
	for i, o := range objects {
		if !filter.Test(o) {
			continue
		}
		l.selection.Model().Cast().(*gtk.StringList).Append(strconv.Itoa(i))
		if o.GetUID() == l.selected {
			l.selection.SetSelected(uint(i))
		}
	}

	if len(l.objects) > 0 {
		if selected := l.selection.Selected(); selected == GtkInvalidListPosition {
			l.selection.SetSelected(0)
			l.behavior.RootDetailBehavior.SelectedObject.Update(l.objects[0])
		} else {
			i, _ := strconv.Atoi(l.selection.ListModel.Item(selected).Cast().(*gtk.StringObject).String())
			l.behavior.RootDetailBehavior.SelectedObject.Update(l.objects[i])
		}
	} else {
		l.behavior.RootDetailBehavior.SelectedObject.Update(nil)
	}
}

func (l *ListView) onSearchFilterChange(filter behavior.SearchFilter) {
	list := l.selection.Model().Cast().(*gtk.StringList)
	list.Splice(0, list.NItems(), nil)
	l.selection.SetSelected(GtkInvalidListPosition)
	for i, object := range l.behavior.Objects.Value() {
		if filter.Test(object) {
			list.Append(strconv.Itoa(i))
		}
		if object.GetUID() == l.selected {
			l.selection.SetSelected(uint(i))
		}
	}
	if list.NItems() > 0 && l.selection.Selected() == GtkInvalidListPosition {
		l.selection.SetSelected(0)
	}
	if l.selection.Selected() != GtkInvalidListPosition {
		// SelectionChanged isn't triggered when calling SetSelected
		l.selection.SelectionChanged(uint(l.selection.Selected()), 1)
	} else {
		l.behavior.RootDetailBehavior.SelectedObject.Update(nil)
	}
}

func (l *ListView) createColumns() []*gtk.ColumnViewColumn {
	var columns []api.Column

	for _, e := range l.behavior.Extensions {
		columns = e.CreateColumns(l.behavior.SelectedResource.Value(), columns)
	}
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Priority > columns[j].Priority
	})

	var gtkColumns []*gtk.ColumnViewColumn
	for _, col := range columns {
		factory := gtk.NewSignalListItemFactory()
		factory.ConnectBind(func(listitem *gtk.ListItem) {
			idx, _ := strconv.Atoi(listitem.Item().Cast().(*gtk.StringObject).String())
			object := l.objects[idx]
			col.Bind(listitem, object)
		})
		column := gtk.NewColumnViewColumn(col.Name, &factory.ListItemFactory)
		column.SetResizable(true)
		column.SetExpand(true)
		gtkColumns = append(gtkColumns, column)
	}

	return gtkColumns
}

func (l *ListView) createModel() *gtk.SingleSelection {
	selection := gtk.NewSingleSelection(gtk.NewStringList([]string{}))
	selection.ConnectSelectionChanged(func(_, _ uint) {
		selected := l.selection.Selected()
		if selected == GtkInvalidListPosition {
			return
		}
		i, _ := strconv.Atoi(l.selection.ListModel.Item(selected).Cast().(*gtk.StringObject).String())
		obj := l.objects[i]
		l.selected = obj.GetUID()
		l.behavior.RootDetailBehavior.SelectedObject.Update(obj)
	})
	return selection
}
