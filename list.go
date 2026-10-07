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

// sharedPaths is the sub-paths every item carries — the only ones a path may step
// through a choice to reach. It intersects, bailing as soon as the set
// is empty, which is immediate for a choice of plain strings.
func sharedPaths(items []node, intoRepeats bool) map[string]bool {
	shared := subPaths(items[0], intoRepeats)
	for _, it := range items[1:] {
		if len(shared) == 0 {
			return nil
		}
		next := subPaths(it, intoRepeats)
		for p := range shared {
			if !next[p] {
				delete(shared, p)
			}
		}
	}
	return shared
}

// subPaths is paths(n, intoRepeats) as a set, without the empty path that means n itself.
func subPaths(n node, intoRepeats bool) map[string]bool {
	out := map[string]bool{}
	for _, p := range paths(n, intoRepeats) {
		if p != "" {
			out[p] = true
		}
	}
	return out
}

func join(prefix, name string) string {
	switch {
	case prefix == "":
		return name
	case name == "":
		return prefix
	}
	return prefix + "." + name
}
