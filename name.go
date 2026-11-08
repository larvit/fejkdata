package fejkdata

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/larvit/fejkdata/internal/grammar"
)

// nameScope is where a name is bound: a category, one iteration of a repeat inside it, or one
// render of a choice's item. A token sees its scope's names and every enclosing scope's, so one
// field may bind a name and another read it.
type nameScope struct {
	up       *nameScope
	owner    node // the category's root, the repeat, or the choice's item
	bindings map[string]*nameBinding
	order    []*nameBinding
}

type nameBinding struct {
	// Filled by `bindNames`, from the compiled category:
	name     string
	ref      string // what it binds, as written: a reference, or a path into a field of binder
	body     string
	distinct *grammar.Distinct // what a {ref distinct N as name} draws; nil for one pick
	where    string            // the fields reaching the binding template, as a compile error spells them
	scope    *nameScope
	index    int
	binder   *template

	// Filled by `resolveTemplates`, once every check it runs passed:
	target nameTarget
	// addressed is every key a read of the name lands on or passes, the spelling of the
	// first read reaching it beside it; a pick keeps the draws at these keys, and only these.
	addressed map[pickKey]string
	values    *valueDraw // what a {ref where … as name} or {ref distinct N as name} draws from; nil without either clause
}

// nameTarget is what a binding resolves to in the assembled tree: start, the node its head names, and
// the path it reads into that node.
type nameTarget struct {
	start node
	tail  []string
}

// bindsField reports whether b binds a path into a field, so a read of it stays in the category.
func (b *nameBinding) bindsField() bool { return !grammar.IsRef(b.ref) }

func (sc *nameScope) lookup(name string) *nameBinding {
	for ; sc != nil; sc = sc.up {
		if b, ok := sc.bindings[name]; ok {
			return b
		}
	}
	return nil
}

// bindNames gives every template of a compiled category or inline template the scope its
// names live in, and refuses a name bound twice along one chain of scopes, a name a field spells
// too, and a read no field or name answers.
func bindNames(root node) error {
	var scopes []*nameScope
	var gather func(n node, scope *nameScope, item bool, where string) error
	gather = func(n node, scope *nameScope, item bool, where string) error {
		if t, isTemplate := n.(*template); n == root || isTemplate && (t.repeat > 1 || item) {
			scope = &nameScope{up: scope, owner: n}
			scopes = append(scopes, scope)
		}
		switch t := n.(type) {
		case *template:
			t.nameScope = scope
			if err := scope.bindAll(t, where); err != nil {
				return err
			}
		case *choice:
			t.nameScope = scope
		}
		_, isChoice := n.(*choice)
		return eachContained(n, where, func(c containedNode, where string) error { return gather(c.node, scope, isChoice, where) })
	}
	if err := gather(root, nil, false, ""); err != nil {
		return err
	}
	for _, sc := range scopes[1:] {
		for _, b := range sc.order {
			if outer := sc.up.lookup(b.name); outer != nil {
				return fmt.Errorf("%stoken {%s}: name %q is bound outside this %s too, by {%s}; rename one", b.where, b.body, b.name, sc.kind(), outer.body)
			}
		}
	}
	return answerReads(root, scopes)
}

func (sc *nameScope) bindAll(t *template, where string) error {
	for _, tok := range t.tokens {
		if tok.Kind != grammar.NameBind {
			continue
		}
		if b, twice := sc.bindings[tok.Bound]; twice {
			return fmt.Errorf("%stoken {%s}: name %q is bound twice %s, by {%s} too; rename one", where, tok.Body, tok.Bound, sc.spelled(), b.body)
		}
		b := &nameBinding{name: tok.Bound, ref: tok.BoundRef, body: tok.Body, distinct: tok.Distinct, where: where, scope: sc, index: len(sc.order), binder: t}
		if sc.bindings == nil {
			sc.bindings = map[string]*nameBinding{}
		}
		sc.bindings[b.name] = b
		sc.order = append(sc.order, b)
	}
	return nil
}

// within reports whether sc is other or lies inside it, so other's frame is open wherever sc's is.
func (sc *nameScope) within(other *nameScope) bool {
	for ; sc != nil; sc = sc.up {
		if sc == other {
			return true
		}
	}
	return false
}

func (sc *nameScope) spelled() string {
	if k := sc.kind(); k != scopeCategory {
		return "in one " + k.String()
	}
	return "outside any repeat"
}

// scopeKind is what a name scope is: the category's own, a repeat's, or a choice's item's.
type scopeKind uint8

const (
	scopeCategory scopeKind = iota
	scopeRepeat
	scopeChoiceItem
)

func (k scopeKind) String() string {
	return [...]string{scopeCategory: "category", scopeRepeat: "repeat", scopeChoiceItem: "choice item"}[k]
}

func (sc *nameScope) kind() scopeKind {
	switch t, isTemplate := sc.owner.(*template); {
	case isTemplate && t.repeat > 1:
		return scopeRepeat
	case sc.up == nil:
		return scopeCategory
	}
	return scopeChoiceItem
}

// ownScope is the scope t renders a frame of: a category's, once per render of its root, a
// repeat's, once per iteration, or a choice item's, once per draw of the item. It is nil where t owns no scope, or its scope binds no name.
func (t *template) ownScope() *nameScope {
	if sc := t.nameScope; sc != nil && sc.owner == node(t) && len(sc.order) > 0 {
		return sc
	}
	return nil
}

// eachContained calls fn for each node n contains, with where naming the fields reaching it as
// compile errors do.
func eachContained(n node, where string, fn func(c containedNode, where string) error) error {
	for _, c := range contained(n) {
		next := where
		if c.name != "" {
			next += fmt.Sprintf("field %q: ", c.name)
		}
		if err := fn(c, next); err != nil {
			return err
		}
	}
	return nil
}

// answerReads proves each unbound read under root is answered by a name its template sees. It answers a template's fields before its format, the
// order compile reports in, and refuses a field spelling a name. A refusal searches scopes, every
// scope of the category, for a name bound where the read cannot see it.
func answerReads(root node, scopes []*nameScope) error {
	var answer func(n node, where string, up besideField) error
	answer = func(n node, where string, up besideField) error {
		t, isTemplate := n.(*template)
		err := eachContained(n, where, func(c containedNode, where string) error {
			inner := up
			switch n.(type) {
			case *template:
				inner = besideField{t: t, field: c.name}
			case *choice:
				inner.inItem = true
			}
			return answer(c.node, where, inner)
		})
		if err != nil || !isTemplate {
			return err
		}
		for _, name := range sortedNames(t.fields) {
			if b := t.nameScope.lookup(name); b != nil {
				holder := "the root template"
				if where != "" {
					holder = strings.TrimSuffix(where, ": ")
				}
				return fmt.Errorf("%stoken {%s}: name %q is a field of %s too; options, fields and names share one namespace, so rename one", b.where, b.body, name, holder)
			}
		}
		for _, u := range t.unbound {
			if t.nameScope.lookup(u.head) == nil {
				return unanswered(where, u, t.nameScope, scopes, up)
			}
		}
		return nil
	}
	return answer(root, "", besideField{})
}

// besideField locates a template from above: t is the nearest template containing it, field
// is the field of t it lies under, and inItem says it is an item of a choice in that field.
type besideField struct {
	t      *template
	field  string
	inItem bool
}

// refusal is the refusal of u where up's template holds u's head.
func (up besideField) refusal(where string, u unboundRead) error {
	holder := fmt.Sprintf("field %q", up.field)
	if up.inItem {
		holder = fmt.Sprintf("the item of field %q holding it", up.field)
	}
	fix := fmt.Sprintf("give %s a format and a %q of its own", holder, u.head)
	if up.t.nameScope.kind() != scopeChoiceItem {
		fix += fmt.Sprintf(", or bind {%s as n} in the format beside field %q and read n in place of %s", u.head, up.field, u.head)
	}
	return fmt.Errorf("%stoken {%s}: %q is a field beside field %q, and a token reads only the fields of the template whose format holds it: %s", where, u.body, u.head, up.field, fix)
}

// unanswered is the refusal of u, a read no field or name answers: naming the repeat binding the
// name where the read cannot see it, else the names the read sees.
func unanswered(where string, u unboundRead, seen *nameScope, scopes []*nameScope, up besideField) error {
	for _, sc := range scopes {
		if b, ok := sc.bindings[u.head]; ok {
			at := strings.TrimSuffix(b.where, ": ")
			if at == "" {
				at = "an item of the root choice"
			}
			if sc.kind() == scopeChoiceItem {
				return fmt.Errorf("%s%w; name %q is bound at %s, inside an item of a choice, which a read outside that item cannot see; bind it outside the choice", where, u.err, u.head, at)
			}
			return fmt.Errorf("%s%w; name %q is bound at %s, inside a repeat, which a read outside the repeat cannot see; bind it outside the repeat to read one pick on every line", where, u.err, u.head, at)
		}
	}
	if up.t != nil && up.t.fields[u.head] != nil {
		return up.refusal(where, u)
	}
	var visible []string
	for sc := seen; sc != nil; sc = sc.up {
		for _, b := range sc.order {
			visible = append(visible, b.name)
		}
	}
	if len(visible) > 0 {
		sort.Strings(visible)
		for i, name := range visible {
			visible[i] = strconv.Quote(name)
		}
		return fmt.Errorf("%stoken {%s}: no field or name %q; the names bound here are %s", where, u.body, u.head, strings.Join(visible, ", "))
	}
	return fmt.Errorf("%s%w", where, u.err)
}

func (t *template) isName(head string) bool {
	return !grammar.IsRef(head) && t.fields[head] == nil && t.nameScope.lookup(head) != nil
}

// namedReadAt is one read of a name in a format: the op holding it, and the compiled arm.
type namedReadAt struct {
	o *op
	a *arm
}

// namedReads is every read of a name t's format makes, in format order.
func namedReads(t *template) []namedReadAt {
	var out []namedReadAt
	for i := range t.compiled.ops {
		o := &t.compiled.ops[i]
		for j := range o.arms {
			if o.arms[j].kind == namedRead {
				out = append(out, namedReadAt{o: o, a: &o.arms[j]})
			}
		}
		for j := range o.operands {
			if o.operands[j].kind == namedRead {
				out = append(out, namedReadAt{o: o, a: &o.operands[j]})
			}
		}
	}
	return out
}

// readsOfNames is every read of each name in ts, in the order ts lists them.
func readsOfNames(ts []templateSite) map[*nameBinding][]*arm {
	reads := map[*nameBinding][]*arm{}
	for _, s := range ts {
		for _, r := range namedReads(s.t) {
			reads[r.a.named] = append(reads[r.a.named], r.a)
		}
	}
	return reads
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
