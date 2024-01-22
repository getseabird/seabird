package ui

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// TODO more sophisticated search - maybe using bleve? https://github.com/blevesearch/bleve
type SearchFilter struct {
	Name string
}

type SearchBar struct {
	*gtk.SearchEntry
}

func NewSearchBar(root *ClusterWindow) *SearchBar {
	entry := gtk.NewSearchEntry()
	bar := gtk.NewSearchBar()

	bar.ConnectEntry(entry)

	entry.ConnectSearchChanged(func() {
		root.listView.SetFilter(SearchFilter{Name: entry.Text()})
	})

	return &SearchBar{entry}
}
