package fejkdata

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/grammar"
)

// linkSite is a template to link, with its folder and category, both empty for an inline
// template, and eachNode's label for it.
type linkSite struct {
	t               *template
	folder          []string
	category, label string
}

// linkCategories links every template of the categories, once all data is merged, so a reference
// sees the override-resolved tree.
func linkCategories(sites []categorySite, root map[string]node) error {
	var ts []linkSite
	for _, s := range sites {
		if err := eachNode(s.n, s.path, func(label string, n node) error {
			if t, isTemplate := n.(*template); isTemplate {
				ts = append(ts, linkSite{t: t, folder: s.dir, category: s.path, label: label})
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return linkTemplates(ts, root)
}

// linkInline links the templates of an inline node against the loaded tree.
func linkInline(scope nodeScope, root map[string]node) error {
	var ts []linkSite
	if err := scope(func(label string, n node) error {
		if t, isTemplate := n.(*template); isTemplate {
			ts = append(ts, linkSite{t: t, label: label})
		}
		return nil
	}); err != nil {
		return err
	}
	return linkTemplates(ts, root)
}

// nameTarget is what a binding resolves to in the assembled tree: the node its head names and the
// path it reads into that node.
type nameTarget struct {
	head node
	tail []string
}

// linkedNames is what the link resolved about the names of a set of templates: each binding's
// target, the keys its reads address, and its reads.
type linkedNames struct {
	targets   map[*nameBinding]nameTarget
	addressed map[*nameBinding]map[pickKey]string
	uses      map[*nameBinding][]nameUse
}

// linkTemplates links ts in steps, each over every template before the next starts. Each step
// returns what it builds. The references, the compiled format and the column read are kept on each
// template. The names' targets, keys and reads pass to later steps as arguments, and the bindings
// take their targets and keys only once every check has passed.
func linkTemplates(ts []linkSite, root map[string]node) error {
	for _, s := range ts {
		link, err := s.t.resolveLink(s.folder, s.label, s.category, root)
		if err != nil {
			return err
		}
		s.t.link = link
	}
	targets := nameTargets(ts)
	for _, s := range ts {
		compiled, err := compileFormat(s.t, targets)
		if err != nil {
			return fmt.Errorf("%s: %w", s.label, err)
		}
		s.t.compiled = compiled
	}
	names := linkedNames{targets: targets, addressed: addressedKeys(ts, targets), uses: nameUses(ts)}
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

// nameTargets resolves what each binding of ts binds, from its binder's link.
func nameTargets(ts []linkSite) map[*nameBinding]nameTarget {
	targets := map[*nameBinding]nameTarget{}
	for _, s := range ts {
		for _, tok := range s.t.tokens {
			if tok.Kind == grammar.NameBind {
				a := splitArm(tok.BoundRef, s.t.link.refs)
				targets[s.t.nameScope.bindings[tok.Bound]] = nameTarget{head: s.t.head(a.head), tail: a.tail}
			}
		}
	}
	return targets
}

// addressedKeys is every key the reads of each name in ts land on or pass, from the name's own
// level, with the spelling of the first read reaching it.
func addressedKeys(ts []linkSite, targets map[*nameBinding]nameTarget) map[*nameBinding]map[pickKey]string {
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

// nameUses is every read of each name in ts.
func nameUses(ts []linkSite) map[*nameBinding][]nameUse {
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
	head, category, tail := t.head(a.head), a.head[min(1, len(a.head)):], a.tail
	switch b := a.named; {
	case a.kind == namedRead && b.bindsField():
		head, category = b.binder, ""
		tail = append(append([]string{splitArm(b.ref, nil).head}, targets[b].tail...), a.tail...)
	case a.kind == namedRead:
		head, category = targets[b].head, b.binder.link.refs[b.ref].head[1:]
		tail = append(targets[b].tail[:len(targets[b].tail):len(targets[b].tail)], a.tail...)
	case !grammar.IsRef(a.head):
		return nil
	}
	if target, isTemplate := head.(*template); isTemplate && target.isRecord && len(tail) == 1 {
		return &columnRead{a: a, category: category, field: tail[0], column: target.fields[tail[0]]}
	}
	return nil
}
