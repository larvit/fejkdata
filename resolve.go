package fejkdata

import (
	"fmt"
)

type templateSite struct {
	t      *template
	folder []string
	label  string
}

// resolveCategoryTemplates resolves every template of the categories, once all data is merged, so a
// reference sees the override-resolved tree.
func resolveCategoryTemplates(sites []categorySite, root map[string]node) error {
	var ts []templateSite
	for _, s := range sites {
		if err := eachNode(s.n, s.path, func(label string, n node) error {
			if t, isTemplate := n.(*template); isTemplate {
				ts = append(ts, templateSite{t: t, folder: s.dir, label: label})
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return resolveTemplates(ts, root)
}

func resolveInlineTemplates(nodes nodeSet, root map[string]node) error {
	var ts []templateSite
	if err := nodes(func(label string, n node) error {
		if t, isTemplate := n.(*template); isTemplate {
			ts = append(ts, templateSite{t: t, label: label})
		}
		return nil
	}); err != nil {
		return err
	}
	return resolveTemplates(ts, root)
}

// resolvedNames is what resolveTemplates resolved about the names of a set of templates: each binding's
// target, the reads of it, and the keys they address. places caches pathTo: for each node it was
// called from, every node under it with the fields leading there.
type resolvedNames struct {
	targets   map[*nameBinding]nameTarget
	reads     map[*nameBinding][]*arm
	addressed map[*nameBinding]map[pickKey]string
	places    map[node]map[node][]string
}

// pathTo is the fields leading from from down to want, if want lies under from.
func (names resolvedNames) pathTo(from, want node) ([]string, bool) {
	index, built := names.places[from]
	if !built {
		index = map[node][]string{}
		indexPaths(from, nil, index)
		names.places[from] = index
	}
	path, within := index[want]
	return path, within
}

func indexPaths(n node, path []string, index map[node][]string) {
	if _, seen := index[n]; seen {
		return
	}
	index[n] = path
	for _, c := range contained(n) {
		below := path
		if c.name != "" {
			below = append(path[:len(path):len(path)], c.name)
		}
		indexPaths(c.node, below, index)
	}
}

// targetOf is what b binds, whether it is among the templates being resolved or a category's, resolved earlier.
func (names resolvedNames) targetOf(b *nameBinding) nameTarget {
	if target, resolving := names.targets[b]; resolving {
		return target
	}
	return b.target
}

// resolveTemplates resolves ts in steps, each over every template before the next starts. The
// names' targets, keys and reads pass to later steps as arguments, and the bindings take their
// targets and keys only once every check has passed.
func resolveTemplates(ts []templateSite, root map[string]node) error {
	for _, s := range ts {
		refs, err := s.t.resolveRefs(s.folder, s.label, root)
		if err != nil {
			return err
		}
		s.t.refs = refs
	}
	targets := nameTargets(ts)
	for _, s := range ts {
		compiled, err := compileFormat(s.t, targets)
		if err != nil {
			return fmt.Errorf("%s: %w", s.label, err)
		}
		s.t.compiled = compiled
	}
	reads := readsOfNames(ts)
	names := resolvedNames{targets: targets, reads: reads, addressed: addressedKeys(reads, targets), places: map[node]map[node][]string{}}
	for _, s := range ts {
		s.t.readsColumn = columnReadOf(s.t, targets)
	}
	for _, s := range ts {
		if err := checkNameReads(s.label, s.t, names); err != nil {
			return err
		}
	}
	for _, s := range ts {
		if err := checkCalcNames(s.label, s.t); err != nil {
			return err
		}
	}
	for b, target := range targets {
		b.target, b.addressed = target, names.addressed[b]
	}
	return nil
}
