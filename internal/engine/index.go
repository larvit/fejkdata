package engine

import (
	"fmt"
	"slices"
	"sort"

	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
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
		line := IndexLine{Path: s.path, Entry: datafiles.IndexEntry{Parent: parentOf(s.n), Paths: load.Paths(s.n)}}
		sort.Strings(line.Entry.Paths)
		for _, c := range needs(root, s) {
			line.Needs = append(line.Needs, load.CategoryPath(c.dir, c.name))
		}
		out = append(out, line)
	}
	return out
}

// DrawnFrom is the dot path of the category a caller's path draws its value from: the last table
// the path steps through, a linked table being a category of its own, else the category it enters.
func DrawnFrom(root *tree.Folder, path string) (string, error) {
	path, segments, err := callerPath(root, path)
	if err != nil {
		return "", err
	}
	entered, found := categoryOn(root, segments)
	if !found {
		return "", fmt.Errorf("%s names no category", path)
	}
	table, err := load.LastTableOn(root, segments)
	if err != nil {
		return "", fmt.Errorf("%s: %w", path, err)
	}
	if table == nil {
		return load.CategoryPath(entered.dir, entered.name), nil
	}
	sites := categorySites(root)
	at := slices.IndexFunc(sites, func(s categorySite) bool { return s.n == tree.Node(table) })
	if at < 0 {
		panic(invariant.Broken("%s steps through a table that is no category", path))
	}
	return sites[at].path, nil
}
