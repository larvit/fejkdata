package fejkdata

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/grammar"
)

type templateSite struct {
	t      *template
	folder []string
	label  string
}

// resolveCategoryTemplates resolves every template of the categories, once all data is merged, so a
// reference sees the override-resolved tree.
func resolveCategoryTemplates(sites []categorySite, root map[string]node, l *links) error {
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
	return resolveTemplates(ts, root, l)
}

func resolveInlineTemplates(nodes nodeSet, root map[string]node, l *links) error {
	var ts []templateSite
	if err := nodes(func(label string, n node) error {
		if t, isTemplate := n.(*template); isTemplate {
			ts = append(ts, templateSite{t: t, label: label})
		}
		return nil
	}); err != nil {
		return err
	}
	return resolveTemplates(ts, root, l)
}

// resolvedNames is what the fences of resolveTemplates read about the names of a set of
// templates. links holds each binding's target, draw and addressed keys, and reads each
// binding's reads. places caches pathTo: for each node it was called from, every node under
// it with the fields leading there. fresh is each binding's freshFieldReads.
type resolvedNames struct {
	links  *links
	reads  map[*nameBinding][]*arm
	places map[node]map[node][]string
	fresh  map[*nameBinding]map[pickKey][]fieldRead
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
	if t, isTable := n.(*table); isTable {
		for _, linked := range t.childTables() {
			indexPaths(linked, append(path[:len(path):len(path)], linked.rows.Segment()), index)
		}
	}
}

func compileFormats(ts []templateSite, l *links) error {
	for _, s := range ts {
		compiled, err := compileFormat(s.t, l)
		if err != nil {
			return fmt.Errorf("%s: %w", s.label, err)
		}
		l.templates[s.t].compiled = compiled
	}
	return nil
}

// resolveTemplates resolves ts into l in two stages, each over every template before the next
// starts: compile, then fence.
func resolveTemplates(ts []templateSite, root map[string]node, l *links) error {
	names, err := compileTemplates(ts, root, l)
	if err != nil {
		return err
	}
	return fenceTemplates(ts, names)
}

// compileTemplates resolves each template's references and its names' targets, compiles its format
// and each op's call and draw, and finds what each name's reads address.
func compileTemplates(ts []templateSite, root map[string]node, l *links) (resolvedNames, error) {
	for _, s := range ts {
		refs, err := s.t.resolveRefs(s.folder, s.label, root)
		if err != nil {
			return resolvedNames{}, err
		}
		l.template(s.t).refs = refs
	}
	for b, target := range nameTargets(ts, l) {
		l.bindings[b] = &bindingLinks{target: target}
	}
	if err := compileFormats(ts, l); err != nil {
		return resolvedNames{}, err
	}
	if err := buildOps(ts, l); err != nil {
		return resolvedNames{}, err
	}
	for _, s := range ts {
		l.templates[s.t].readsColumn = columnReadOf(s.t, l)
	}
	reads := readsOfNames(ts, l)
	for b, keys := range addressedKeys(reads, l) {
		l.bindings[b].addressed = keys
	}
	names := resolvedNames{links: l, reads: reads, places: map[node]map[node][]string{}, fresh: map[*nameBinding]map[pickKey][]fieldRead{}}
	for b, bound := range reads {
		names.fresh[b] = freshFieldReads(bound, l.bindings[b].addressed, l)
	}
	return names, nil
}

// buildOps completes each op of ts, once every template's format is compiled, and gives l the draw
// of each name bound to a token with a where or a distinct clause. A proof of a draw reads the draws
// of the names that earlier ops bind.
func buildOps(ts []templateSite, l *links) error {
	words := map[wordsAt]*wordList{}
	for _, s := range ts {
		tl := l.templates[s.t]
		for i, parts := range tl.compiled.ops {
			o, err := newOp(s.t, parts, l, words)
			if err != nil {
				return fmt.Errorf("%s: token {%s}: %w", s.label, parts.Body, err)
			}
			tl.compiled.ops[i] = o
			if v, ok := o.draw.(*valueDraw); ok && o.Kind == grammar.NameBind {
				l.bindings[tl.nameScope.bindings[o.Bound]].values = v
			}
		}
	}
	return nil
}

// fenceTemplates refuses a draw that would read itself or is read through a path, a read of a name
// that could disagree with another read of it, and a calc or adddays whose proof fails.
func fenceTemplates(ts []templateSite, names resolvedNames) error {
	if err := refuseSelfReads(ts, names.links); err != nil {
		return err
	}
	if err := refuseFilteredTails(ts, names.links); err != nil {
		return err
	}
	for _, s := range ts {
		if err := checkNameReads(s.label, s.t, names); err != nil {
			return err
		}
	}
	for _, s := range ts {
		if err := checkCalcNames(s.label, s.t, names.links); err != nil {
			return err
		}
		if err := proveAddDays(s.label, s.t, names.links); err != nil {
			return err
		}
	}
	return nil
}
