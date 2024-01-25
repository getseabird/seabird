package behavior

import (
	"strings"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SearchFilter struct {
	Name      []string
	Namespace []string
}

func NewSearchFilter(text string) SearchFilter {
	filter := SearchFilter{}

	for _, term := range strings.Split(text, " ") {
		if strings.HasPrefix(term, "ns:") {
			filter.Namespace = append(filter.Namespace, strings.TrimPrefix(term, "ns:"))
		} else {
			filter.Name = append(filter.Name, term)
		}
	}

	return filter
}

func (f *SearchFilter) Test(object client.Object) bool {
	{
		var ok bool
		for _, n := range f.Namespace {
			if object.GetNamespace() == n {
				ok = true
				break
			}
		}
		if !ok && len(f.Namespace) > 0 {
			return false
		}
	}

	for _, term := range f.Name {
		var ok bool
		trimmed := strings.Trim(term, "\"")
		if strings.Contains(object.GetName(), trimmed) {
			ok = true
			continue
		}
		if term != trimmed {
			continue
		}
		for _, term := range strings.Split(term, "-") {
			for _, name := range strings.Split(object.GetName(), "-") {
				if strutil.Similarity(name, term, metrics.NewHamming()) > 0.5 {
					ok = true
				}
			}
		}
		if !ok {
			return false
		}
	}

	return true
}
