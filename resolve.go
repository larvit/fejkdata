package fejkdata

import (
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// resolveCategoryTemplates resolves every template of the categories, once all data is merged, so a
// reference sees the override-resolved tree.
func resolveCategoryTemplates(sites []categorySite, root map[string]tree.Node, l *tree.LoadFinds) ([]load.TemplateSite, error) {
	var ts []load.TemplateSite
	for _, s := range sites {
		if err := load.EachNode(s.n, s.path, func(label string, n tree.Node) error {
			if t, isTemplate := n.(*tree.Template); isTemplate {
				ts = append(ts, load.TemplateSite{Template: t, Folder: s.dir, Label: label})
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	if err := load.CompileTemplates(ts, root, l); err != nil {
		return nil, err
	}
	return ts, nil
}

// resolvedNames is what the name fences read about the names of a set of finished templates.
// places caches pathTo: for each node it was called from,
// every node under it with the fields leading there. fresh is each binding's freshFieldReads.
type resolvedNames struct {
	reads  map[*tree.NameBinding][]*tree.Arm
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

// fenceTemplates runs the template fences over the finished templates of ts.
func fenceTemplates(ts []load.TemplateSite) error {
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
	reads := load.ReadsOfNames(ts, nil)
	names := resolvedNames{reads: reads, places: map[tree.Node]map[tree.Node][]string{}, fresh: map[*tree.NameBinding]map[tree.PickKey][]fieldRead{}}
	for b, bound := range reads {
		names.fresh[b] = freshFieldReads(bound, b.Addressed())
	}
	return names
}
