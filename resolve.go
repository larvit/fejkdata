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
func resolveCategoryTemplates(sites []categorySite, root map[string]node, l *loadFinds) ([]templateSite, error) {
	var ts []templateSite
	for _, s := range sites {
		if err := eachNode(s.n, s.path, func(label string, n node) error {
			if t, isTemplate := n.(*template); isTemplate {
				ts = append(ts, templateSite{t: t, folder: s.dir, label: label})
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	if err := compileTemplates(ts, root, l); err != nil {
		return nil, err
	}
	return ts, nil
}

func resolveInlineTemplates(nodes nodeSet, root map[string]node, l *loadFinds) ([]templateSite, error) {
	var ts []templateSite
	if err := nodes(func(label string, n node) error {
		if t, isTemplate := n.(*template); isTemplate {
			ts = append(ts, templateSite{t: t, label: label})
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := compileTemplates(ts, root, l); err != nil {
		return nil, err
	}
	return ts, nil
}

// resolvedNames is what the name fences read about the names of a set of finished templates.
// places caches pathTo: for each node it was called from,
// every node under it with the fields leading there. fresh is each binding's freshFieldReads.
type resolvedNames struct {
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

func compileFormats(ts []templateSite, l *loadFinds) error {
	for _, s := range ts {
		compiled, err := compileFormat(s.t, l)
		if err != nil {
			return fmt.Errorf("%s: %w", s.label, err)
		}
		l.templates[s.t].compiled = compiled
	}
	return nil
}

// compileTemplates resolves each template's references and its names' targets, compiles its format
// and each op's call and draw, and finds what each name's reads address.
func compileTemplates(ts []templateSite, root map[string]node, l *loadFinds) error {
	for _, s := range ts {
		refs, err := resolveRefs(s.t, s.folder, s.label, root)
		if err != nil {
			return err
		}
		l.template(s.t).refs = refs
	}
	for b, target := range nameTargets(ts, l) {
		l.bindings[b] = &bindingFinds{target: target}
	}
	if err := compileFormats(ts, l); err != nil {
		return err
	}
	if err := buildOps(ts, l); err != nil {
		return err
	}
	for _, s := range ts {
		l.templates[s.t].readsColumn = columnReadOf(s.t, l)
	}
	for b, keys := range addressedKeys(readsOfNames(ts, l), l) {
		l.bindings[b].addressed = keys
	}
	return nil
}

// buildOps completes each op of ts, once every template's format is compiled, and gives l the draw
// of each name bound to a token with a where or a distinct clause.
func buildOps(ts []templateSite, l *loadFinds) error {
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

// fenceTemplates runs the template fences over the finished templates of ts.
func fenceTemplates(ts []templateSite) error {
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
		if err := checkNameReads(s.label, s.t, names); err != nil {
			return err
		}
	}
	for _, s := range ts {
		if err := checkCalcNames(s.label, s.t); err != nil {
			return err
		}
		if err := proveAddDays(s.label, s.t); err != nil {
			return err
		}
	}
	return nil
}

func namesOf(ts []templateSite) resolvedNames {
	reads := readsOfNames(ts, nil)
	names := resolvedNames{reads: reads, places: map[node]map[node][]string{}, fresh: map[*nameBinding]map[pickKey][]fieldRead{}}
	for b, bound := range reads {
		names.fresh[b] = freshFieldReads(bound, b.addressed)
	}
	return names
}
