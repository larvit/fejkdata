package engine

import (
	"sort"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// IndexLine is a loaded category's dot path, what a manifest's index says of it, and the dot
// path of each category it needs.
type IndexLine struct {
	Path  string
	Entry datafiles.IndexEntry
	Needs []string
}

// Index is an IndexLine for every category root has loaded, in path order.
func Index(root *tree.Folder) []IndexLine {
	var out []IndexLine
	for _, s := range categorySites(root) {
		line := IndexLine{Path: s.path, Entry: datafiles.IndexEntry{Parent: parentOf(s.n), Paths: load.Paths(s.n, false)}}
		sort.Strings(line.Entry.Paths)
		for _, c := range needs(root, s) {
			line.Needs = append(line.Needs, load.CategoryPath(c.dir, c.name))
		}
		out = append(out, line)
	}
	return out
}
