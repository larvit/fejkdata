package tree

import (
	"sort"
)

// SortedNames is the keys of m, sorted: the order a walk over a folder's children or a template's
// fields takes, so a load and what it reports run the same way every time.
func SortedNames[V any](m map[string]V) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
