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
// target and draw, the reads of it, and the keys they address. places caches pathTo: for each node it
// was called from, every node under it with the fields leading there. fresh is each binding's
// freshFieldReads.
type resolvedNames struct {
	targets   map[*nameBinding]nameTarget
	draws     bindingDraws
	reads     map[*nameBinding][]*arm
	addressed map[*nameBinding]map[pickKey]string
	places    map[node]map[node][]string
	fresh     map[*nameBinding]map[pickKey][]fieldRead
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
		for _, linked := range t.rows.Children() {
			indexPaths(linked.Owner(), append(path[:len(path):len(path)], linked.Segment()), index)
		}
	}
}

// targetOf is what b binds, whether it is among the templates being resolved or a category's, resolved earlier.
func (names resolvedNames) targetOf(b *nameBinding) nameTarget {
	if target, resolving := names.targets[b]; resolving {
		return target
	}
	return b.target
}

func compileFormats(ts []templateSite, targets map[*nameBinding]nameTarget) error {
	for _, s := range ts {
		compiled, err := compileFormat(s.t, targets)
		if err != nil {
			return fmt.Errorf("%s: %w", s.label, err)
		}
		s.t.compiled = compiled
	}
	return nil
}

// bindingDraws is the draw of each name bound to a token with a where or a distinct clause, among
// the templates being resolved.
type bindingDraws map[*nameBinding]*valueDraw

// of is b's draw, whether b is among the templates being resolved or a category's, resolved earlier.
func (d bindingDraws) of(b *nameBinding) *valueDraw {
	if v, resolving := d[b]; resolving {
		return v
	}
	return b.values
}

// resolveTemplates resolves ts in three stages, each over every template before the next starts:
// compile, fence and commit. Each stage passes what it found to the next as arguments, and only
// the commit writes the bindings' targets, keys and draws.
func resolveTemplates(ts []templateSite, root map[string]node) error {
	names, err := compileTemplates(ts, root)
	if err != nil {
		return err
	}
	if err := fenceTemplates(ts, names); err != nil {
		return err
	}
	for b, target := range names.targets {
		b.target, b.addressed, b.values = target, names.addressed[b], names.draws[b]
	}
	return nil
}

// compileTemplates resolves each template's references and its names' targets, compiles its format
// and each op's call and draw, and finds what each name's reads address.
func compileTemplates(ts []templateSite, root map[string]node) (resolvedNames, error) {
	for _, s := range ts {
		refs, err := s.t.resolveRefs(s.folder, s.label, root)
		if err != nil {
			return resolvedNames{}, err
		}
		s.t.refs = refs
	}
	targets := nameTargets(ts)
	if err := compileFormats(ts, targets); err != nil {
		return resolvedNames{}, err
	}
	draws, err := buildOps(ts, targets)
	if err != nil {
		return resolvedNames{}, err
	}
	for _, s := range ts {
		s.t.readsColumn = columnReadOf(s.t, targets)
	}
	reads := readsOfNames(ts)
	names := resolvedNames{targets: targets, draws: draws, reads: reads, addressed: addressedKeys(reads, targets), places: map[node]map[node][]string{}, fresh: map[*nameBinding]map[pickKey][]fieldRead{}}
	for b, bound := range reads {
		names.fresh[b] = freshFieldReads(bound, names.addressed[b])
	}
	return names, nil
}

// buildOps completes each op of ts, once every template's format is compiled, and returns the draw
// of each name bound to a token with a where or a distinct clause. A proof of a draw reads the draws
// of the names that earlier ops bind.
func buildOps(ts []templateSite, targets map[*nameBinding]nameTarget) (bindingDraws, error) {
	draws, words := bindingDraws{}, map[wordsAt]*wordList{}
	for _, s := range ts {
		for i, parts := range s.t.compiled.ops {
			o, err := newOp(s.t, parts, targets, draws, words)
			if err != nil {
				return nil, fmt.Errorf("%s: token {%s}: %w", s.label, parts.Body, err)
			}
			s.t.compiled.ops[i] = o
			if v, ok := o.draw.(*valueDraw); ok && o.Kind == grammar.NameBind {
				draws[s.t.nameScope.bindings[o.Bound]] = v
			}
		}
	}
	return draws, nil
}

// fenceTemplates refuses a draw that would read itself or is read through a path, a read of a name
// that could disagree with another read of it, and a calc or adddays whose proof fails.
func fenceTemplates(ts []templateSite, names resolvedNames) error {
	if err := refuseSelfReads(ts, names.draws); err != nil {
		return err
	}
	if err := refuseFilteredTails(ts, names.draws); err != nil {
		return err
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
		if err := proveAddDays(s.label, s.t, names.draws); err != nil {
			return err
		}
	}
	return nil
}
