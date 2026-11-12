// Package fence proves a finished node tree at load. It refuses a reference cycle, a repeat past
// the cap, a template that could render a value its datatype or its calc forbids, and one that two
// readers could read differently.
package fence

import (
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// NodeSet is the set of nodes one fence covers, walked by calling it with fn.
type NodeSet = func(fn func(label string, n tree.Node) error) error

// resolvedNames is what the name fences read about the names of a set of finished templates.
// places caches pathTo: per starting node, the fields leading down from it to each node below. fresh is each binding's freshFieldReads.
type resolvedNames struct {
	reads  map[*tree.NameBinding][]*tree.Read
	places map[tree.Node]map[tree.Node][]string
	fresh  map[*tree.NameBinding]map[tree.PickKey][]fieldRead
}

// pathTo is the fields leading from from down to want, if want lies under from.
func (names resolvedNames) pathTo(from, want tree.Node) ([]string, bool) {
	index, built := names.places[from]
	if !built {
		index = map[tree.Node][]string{}
		load.IndexPaths(from, nil, index)
		names.places[from] = index
	}
	path, within := index[want]
	return path, within
}

// Templates runs the template fences over the finished templates of ts.
func Templates(ts []load.TemplateSite) error {
	if err := proveDrawLengths(ts); err != nil {
		return err
	}
	if err := refuseSelfReads(ts); err != nil {
		return err
	}
	if err := refuseFilteredTails(ts); err != nil {
		return err
	}
	names := namesOf(ts)
	for _, s := range ts {
		if err := checkNameReads(s.Label, s.Template, names); err != nil {
			return err
		}
	}
	proof := &valueProof{}
	for _, s := range ts {
		if err := proof.checkCalcs(s.Label, s.Template); err != nil {
			return err
		}
		if err := proveAddDays(s.Label, s.Template); err != nil {
			return err
		}
	}
	return nil
}

func namesOf(ts []load.TemplateSite) resolvedNames {
	reads := load.ReadsOfNames(ts, tree.AfterFinish)
	names := resolvedNames{reads: reads, places: map[tree.Node]map[tree.Node][]string{}, fresh: map[*tree.NameBinding]map[tree.PickKey][]fieldRead{}}
	for b, bound := range reads {
		names.fresh[b] = freshFieldReads(bound, b.Addressed())
	}
	return names
}

// CheckNodes runs the per-node fences over s, each over every node before the next,
// so which of several broken nodes is reported does not depend on the walk.
func CheckNodes(s NodeSet) error {
	mem := RenderCounts{}
	if err := s(func(label string, n tree.Node) error { return RepeatCheck(label, n, mem) }); err != nil {
		return err
	}
	return s((&valueProof{}).checkDatatype)
}
