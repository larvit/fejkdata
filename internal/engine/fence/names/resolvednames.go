package names

import (
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// resolvedNames is what the name fences read about the names of a set of finished templates.
// paths caches pathTo: per starting node, the fields leading down from it to each node below. fresh is each binding's freshFieldReads.
type resolvedNames struct {
	reads map[*tree.NameBinding][]*tree.Read
	paths map[tree.Node]map[tree.Node][]string
	fresh map[*tree.NameBinding]map[tree.PickKey][]fieldRead
}

// pathTo is the fields leading from from down to want, if want lies under from.
func (names resolvedNames) pathTo(from, want tree.Node) ([]string, bool) {
	index, built := names.paths[from]
	if !built {
		index = map[tree.Node][]string{}
		load.IndexPaths(from, nil, index)
		names.paths[from] = index
	}
	path, within := index[want]
	return path, within
}

func namesOf(ts []load.TemplateSite) resolvedNames {
	reads := load.ReadsOfNames(ts, tree.AfterFinish)
	names := resolvedNames{reads: reads, paths: map[tree.Node]map[tree.Node][]string{}, fresh: map[*tree.NameBinding]map[tree.PickKey][]fieldRead{}}
	for b, bound := range reads {
		names.fresh[b] = freshFieldReads(bound, b.Addressed())
	}
	return names
}
