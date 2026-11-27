package engine

import (
	"github.com/larvit/fejkdata/internal/datafiles"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// CategoryReads is a category Load compiles, and the path from the root of each reference its
// templates spell.
type CategoryReads struct {
	Path  string
	Reads [][]string
}

// References compiles sources as Load does, without resolving a reference, so it reads data Load
// would refuse for a reference to nothing. It returns what each category references, and the path of each
// category an index leaves unloaded.
func References(sources []datafiles.Source) (compiled []CategoryReads, unloaded []string, err error) {
	root, _, err := compile(sources)
	if err != nil {
		return nil, nil, loadError{err}
	}
	for _, s := range categorySites(root) {
		compiled = append(compiled, CategoryReads{Path: s.path, Reads: spelledReads(s)})
	}
	return compiled, unloadedPaths(nil, root), nil
}

// spelledReads is the path of each reference the templates of s spell, save one that no path can be read from.
func spelledReads(s categorySite) [][]string {
	var reads [][]string
	_ = load.EachNode(s.n, s.path, func(_ string, n tree.Node) error {
		t, isTemplate := n.(*tree.Template)
		if !isTemplate {
			return nil
		}
		for _, spelling := range load.RefSpellings(t.Tokens()) {
			if segs, err := load.RefSegments(spelling, s.dir); err == nil {
				reads = append(reads, segs)
			}
		}
		return nil
	})
	return reads
}

func unloadedPaths(dir []string, g *tree.Folder) []string {
	var paths []string
	for _, name := range tree.SortedNames(g.Unloaded()) {
		paths = append(paths, load.CategoryPath(dir, name))
	}
	for _, name := range tree.SortedNames(g.Children()) {
		if sub, isFolder := g.Children()[name].(*tree.Folder); isFolder {
			paths = append(paths, unloadedPaths(append(dir[:len(dir):len(dir)], name), sub)...)
		}
	}
	return paths
}
