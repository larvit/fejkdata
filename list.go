package fejkdata

import (
	"sort"

	"github.com/larvit/fejkdata/internal/invariant"
)

// paths lists the dot paths addressable from n, relative to it, where "" is n
// itself. A folder has no value of its own, so it contributes only its children's.
// With intoRepeats, it also lists the paths below a level carrying a repeat, so a path
// there meets the repeat's refusal.
func paths(n node, intoRepeats bool) []string {
	switch n := n.(type) {
	case *folder:
		var out []string
		for _, name := range sortedNames(n.children) {
			out = appendUnder(out, name, paths(n.children[name], intoRepeats))
		}
		for _, name := range sortedNames(n.unloaded) {
			out = appendUnder(out, name, n.unloaded[name].paths)
		}
		return out
	case *template:
		out := []string{""}
		if n.repeat > 1 && !intoRepeats {
			return out
		}
		for _, name := range sortedNames(n.fields) {
			out = appendUnder(out, name, paths(n.fields[name], intoRepeats))
		}
		return out
	case *table:
		return tablePaths(n, intoRepeats)
	case *tableColumn, *tableRow, *nullItem:
		return []string{""}
	case *choice:
		shared := n.shared
		if !intoRepeats {
			shared = sharedPaths(n.items, false)
		}
		return append([]string{""}, sortedNames(shared)...)
	default:
		panic(invariant.Broken("paths has no case for node %T", n))
	}
}

func appendUnder(out []string, name string, ps []string) []string {
	for _, p := range ps {
		out = append(out, join(name, p))
	}
	return out
}

// tablePaths is a table's columns, then each table linked to it under its name: the
// direct descents, a step at a time.
func tablePaths(t *table, intoRepeats bool) []string {
	out := append([]string{""}, t.rows.Header()...)
	sort.Strings(out[1:])
	for _, c := range t.rows.Children() {
		out = appendUnder(out, c.Segment(), paths(c.Owner(), intoRepeats))
	}
	return out
}
