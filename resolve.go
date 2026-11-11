package fejkdata

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

type templateSite struct {
	t      *tree.Template
	folder []string
	label  string
}

// resolveCategoryTemplates resolves every template of the categories, once all data is merged, so a
// reference sees the override-resolved tree.
func resolveCategoryTemplates(sites []categorySite, root map[string]tree.Node, l *tree.LoadFinds) ([]templateSite, error) {
	var ts []templateSite
	for _, s := range sites {
		if err := eachNode(s.n, s.path, func(label string, n tree.Node) error {
			if t, isTemplate := n.(*tree.Template); isTemplate {
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

// resolveInlineTemplates resolves every template in n, labelled from label, against the categories
// in root.
func resolveInlineTemplates(n tree.Node, label string, root map[string]tree.Node, l *tree.LoadFinds) ([]templateSite, error) {
	var ts []templateSite
	if err := eachNode(n, label, func(label string, n tree.Node) error {
		if t, isTemplate := n.(*tree.Template); isTemplate {
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
	reads  map[*tree.NameBinding][]*tree.Arm
	places map[tree.Node]map[tree.Node][]string
	fresh  map[*tree.NameBinding]map[tree.PickKey][]fieldRead
}

// pathTo is the fields leading from from down to want, if want lies under from.
func (names resolvedNames) pathTo(from, want tree.Node) ([]string, bool) {
	index, built := names.places[from]
	if !built {
		index = map[tree.Node][]string{}
		indexPaths(from, nil, index)
		names.places[from] = index
	}
	path, within := index[want]
	return path, within
}

func indexPaths(n tree.Node, path []string, index map[tree.Node][]string) {
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
	if t, isTable := n.(*tree.Table); isTable {
		for _, linked := range t.ChildTables() {
			indexPaths(linked, append(path[:len(path):len(path)], linked.Rows().Segment()), index)
		}
	}
}

func compileFormats(ts []templateSite, l *tree.LoadFinds) error {
	for _, s := range ts {
		compiled, err := compileFormat(s.t, l)
		if err != nil {
			return fmt.Errorf("%s: %w", s.label, err)
		}
		l.SetCompiled(s.t, compiled)
	}
	return nil
}

// compileTemplates resolves each template's references and its names' targets, compiles its format
// and each op's call and draw, and finds what each name's reads address.
func compileTemplates(ts []templateSite, root map[string]tree.Node, l *tree.LoadFinds) error {
	for _, s := range ts {
		refs, err := resolveRefs(s.t, s.folder, s.label, root)
		if err != nil {
			return err
		}
		l.SetRefs(s.t, refs)
	}
	for b, target := range nameTargets(ts, l) {
		l.SetTarget(b, target)
	}
	if err := compileFormats(ts, l); err != nil {
		return err
	}
	if err := buildOps(ts, l); err != nil {
		return err
	}
	for _, s := range ts {
		l.SetReadsColumn(s.t, columnReadOf(s.t, l))
	}
	for b, keys := range tree.AddressedKeys(readsOfNames(ts, l), l) {
		l.SetAddressed(b, keys)
	}
	return nil
}

// buildOps completes each op of ts, once every template's format is compiled, and gives l the draw
// of each name bound to a token with a where or a distinct clause.
func buildOps(ts []templateSite, l *tree.LoadFinds) error {
	words := map[tree.WordsAt]*tree.WordList{}
	for _, s := range ts {
		tl := l.Of(s.t)
		for i, parts := range tl.Compiled().Ops() {
			o, err := newOp(s.t, parts, l, words)
			if err != nil {
				return fmt.Errorf("%s: token {%s}: %w", s.label, parts.Token().Body, err)
			}
			l.SetOp(s.t, i, o)
			if v, ok := o.Draw().(*tree.ValueDraw); ok && o.Token().Kind == grammar.NameBind {
				l.SetValues(tl.NameScope().Bindings()[o.Token().Bound], v)
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
	names := resolvedNames{reads: reads, places: map[tree.Node]map[tree.Node][]string{}, fresh: map[*tree.NameBinding]map[tree.PickKey][]fieldRead{}}
	for b, bound := range reads {
		names.fresh[b] = freshFieldReads(bound, b.Addressed())
	}
	return names
}
