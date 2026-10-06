package fejkdata

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/grammar"
)

// templateSite is a template to resolve, with its folder and category, both empty for an inline
// template, and eachNode's label for it.
type templateSite struct {
	t               *template
	folder          []string
	category, label string
}

// resolveCategoryTemplates resolves every template of the categories, once all data is merged, so a
// reference sees the override-resolved tree.
func resolveCategoryTemplates(sites []categorySite, root map[string]node) error {
	var ts []templateSite
	for _, s := range sites {
		if err := eachNode(s.n, s.path, func(label string, n node) error {
			if t, isTemplate := n.(*template); isTemplate {
				ts = append(ts, templateSite{t: t, folder: s.dir, category: s.path, label: label})
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

// nameTarget is what a binding resolves to in the assembled tree: start, the node its head names, and
// the path it reads into that node.
type nameTarget struct {
	start node
	tail  []string
}

// resolvedNames is what resolveTemplates resolved about the names of a set of templates: each binding's
// target, the keys its reads address, and its reads.
type resolvedNames struct {
	targets   map[*nameBinding]nameTarget
	addressed map[*nameBinding]map[pickKey]string
	uses      map[*nameBinding][]nameUse
}

// resolveTemplates resolves ts in steps, each over every template before the next starts. The
// names' targets, keys and reads pass to later steps as arguments, and the bindings take their
// targets and keys only once every check has passed.
func resolveTemplates(ts []templateSite, root map[string]node) error {
	for _, s := range ts {
		refs, err := s.t.resolveRefs(s.folder, s.label, s.category, root)
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
	names := resolvedNames{targets: targets, addressed: addressedKeys(ts, targets), uses: nameUses(ts)}
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
		b.nameTarget, b.addressed = target, names.addressed[b]
	}
	return nil
}

// nameTargets resolves what each binding of ts binds, from its binder's references.
func nameTargets(ts []templateSite) map[*nameBinding]nameTarget {
	targets := map[*nameBinding]nameTarget{}
	for _, s := range ts {
		for _, tok := range s.t.tokens {
			if tok.Kind == grammar.NameBind {
				a := splitArm(tok.BoundRef, s.t.refs.byName)
				targets[s.t.nameScope.bindings[tok.Bound]] = nameTarget{start: s.t.startOf(a.head), tail: a.tail}
			}
		}
	}
	return targets
}

// addressedKeys is every key the reads of each name in ts land on or pass, from the name's own
// level, with the spelling of the first read reaching it.
func addressedKeys(ts []templateSite, targets map[*nameBinding]nameTarget) map[*nameBinding]map[pickKey]string {
	keys := map[*nameBinding]map[pickKey]string{}
	for _, s := range ts {
		for _, r := range namedReads(s.t) {
			b := r.a.named
			if keys[b] == nil {
				keys[b] = map[pickKey]string{}
			}
			for _, key := range r.a.levels[len(targets[b].tail):] {
				if _, seen := keys[b][key]; !seen {
					keys[b][key] = r.a.spelling
				}
			}
		}
	}
	return keys
}

func nameUses(ts []templateSite) map[*nameBinding][]nameUse {
	uses := map[*nameBinding][]nameUse{}
	for _, s := range ts {
		for _, r := range namedReads(s.t) {
			b := r.a.named
			uses[b] = append(uses[b], nameUse{
				tail:    grammar.JoinSegments(r.a.tail),
				in:      s.t,
				operand: r.operand,
				noRef:   r.operand && builtinfunc.NoRefOperands(r.o.Fn),
				nested:  s.t.nameScope != b.scope,
			})
		}
	}
	return uses
}

// columnRead is a record's column read by a format that only reads one reference or name,
// which is the column: it takes the column's datatype and null. category and field name it;
// category is "" for a column of the reading template's own record, which checkColumns reaches
// on its own.
type columnRead struct {
	a               arm
	category, field string
	column          node
}

// columnReadOf is the record's column t's format reads, where the format only reads one reference
// or name, and nil where it does not.
func columnReadOf(t *template, targets map[*nameBinding]nameTarget) *columnRead {
	ops := t.compiled.ops
	if t.repeat != 1 || len(ops) != 1 || ops[0].Kind != grammar.NameRead || len(ops[0].arms) != 1 {
		return nil
	}
	a := ops[0].arms[0]
	var start node
	var category string
	var tail []string
	switch b := a.named; {
	case a.kind == namedRead && b.bindsField():
		// {f as n}{n}: column f of the record binding n, in t's own category.
		start, tail = b.binder, append(append([]string{splitArm(b.ref, nil).head}, targets[b].tail...), a.tail...)
	case a.kind == namedRead:
		// {/c as n}{n.x} or {/c.x as n}{n}: column x of category c.
		start, category = targets[b].start, categoryOf(b.binder.refs.byName[b.ref].head)
		tail = append(targets[b].tail[:len(targets[b].tail):len(targets[b].tail)], a.tail...)
	case grammar.IsRef(a.head):
		// {/c.x}: column x of category c.
		start, category, tail = t.startOf(a.head), categoryOf(a.head), a.tail
	default:
		return nil
	}
	if target, isTemplate := start.(*template); isTemplate && target.isRecord && len(tail) == 1 {
		return &columnRead{a: a, category: category, field: tail[0], column: target.fields[tail[0]]}
	}
	return nil
}
