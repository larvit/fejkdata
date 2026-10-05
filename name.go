package fejkdata

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/larvit/fejkdata/internal/grammar"
)

// nameScope is where a name is bound: a category, or one iteration of a repeat inside it. A
// token sees its scope's names and every enclosing scope's, so one field may bind a name and
// another read it.
type nameScope struct {
	up       *nameScope
	owner    node // the category's root, or the repeat
	bindings map[string]*nameBinding
	order    []*nameBinding
}

type nameBinding struct {
	// Filled by `bindNames`, from the compiled category:
	name   string
	ref    string // what it binds, as written: a reference, or a path into a field of binder
	body   string
	where  string // the fields reaching the binding template, as a compile error spells them
	scope  *nameScope
	index  int
	binder *template

	// Filled by `linkTemplates`, from the assembled tree:
	head node
	tail []string
	// addressed is every key a read of the name lands on or passes, the spelling of the
	// first read reaching it beside it; a pick keeps the draws at these keys, and only these.
	addressed map[string]string
}

// nameUse is one read of a name: the path it reads into the name, "" for the name itself; the
// template reading it; whether a builtin reads it as an operand, and whether no reference can
// stand there; and whether it sits in a repeat nested inside the name's scope.
type nameUse struct {
	tail                   string
	in                     *template
	operand, noRef, nested bool
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
// names live in, and refuses a name bound twice along one chain of scopes, a binding in a
// choice's item, a name a field spells too, a read no field or name answers, and a binding
// read by nothing.
func bindNames(root node) error {
	var scopes []*nameScope
	var gather func(n node, scope *nameScope, inChoice bool, where string) error
	gather = func(n node, scope *nameScope, inChoice bool, where string) error {
		if t, isTemplate := n.(*template); n == root || isTemplate && t.repeat > 1 {
			scope = &nameScope{up: scope, owner: n}
			scopes = append(scopes, scope)
			inChoice = false
		}
		switch t := n.(type) {
		case *template:
			t.nameScope = scope
			if err := scope.bindAll(t, inChoice, where); err != nil {
				return err
			}
		case *choice:
			t.nameScope = scope
		}
		_, isChoice := n.(*choice)
		return eachContained(n, where, func(c node, where string) error { return gather(c, scope, inChoice || isChoice, where) })
	}
	if err := gather(root, nil, false, ""); err != nil {
		return err
	}
	for _, sc := range scopes[1:] {
		for _, b := range sc.order {
			if outer := sc.up.lookup(b.name); outer != nil {
				return fmt.Errorf("%stoken {%s}: name %q is bound outside this repeat too, by {%s}; rename one", b.where, b.body, b.name, outer.body)
			}
		}
	}
	reads, err := resolveReads(root, scopes)
	if err != nil {
		return err
	}
	return checkNameUses(scopes, reads)
}

// bindAll binds every name t's tokens bind.
func (sc *nameScope) bindAll(t *template, inChoice bool, where string) error {
	for _, tok := range t.tokens {
		if tok.Kind != grammar.NameBind {
			continue
		}
		if inChoice {
			return fmt.Errorf("%stoken {%s}: a choice's item binds no name, since every other item would leave it unbound; bind it outside the choice, or move the item into a category of its own and reference that", where, tok.Body)
		}
		if b, twice := sc.bindings[tok.Bound]; twice {
			return fmt.Errorf("%stoken {%s}: name %q is bound twice %s, by {%s} too; rename one", where, tok.Body, tok.Bound, sc.spelled(), b.body)
		}
		b := &nameBinding{name: tok.Bound, ref: tok.BoundRef, body: tok.Body, where: where, scope: sc, index: len(sc.order), binder: t}
		if sc.bindings == nil {
			sc.bindings = map[string]*nameBinding{}
		}
		sc.bindings[b.name] = b
		sc.order = append(sc.order, b)
	}
	return nil
}

// spelled names the scope in an error.
func (sc *nameScope) spelled() string {
	if t, isTemplate := sc.owner.(*template); sc.up == nil && !(isTemplate && t.repeat > 1) {
		return "outside any repeat"
	}
	return "in one repeat"
}

// checkNameUses refuses a binding read by nothing, then each binding checkUses refuses.
func checkNameUses(scopes []*nameScope, reads map[*nameBinding][]nameUse) error {
	for _, sc := range scopes {
		for _, b := range sc.order {
			if len(reads[b]) == 0 {
				return fmt.Errorf("%stoken {%s}: nothing reads name %q; drop the token", b.where, b.body, b.name)
			}
		}
	}
	for _, sc := range scopes {
		for _, b := range sc.order {
			if err := b.checkUses(reads[b]); err != nil {
				return fmt.Errorf("%stoken {%s}: %w", b.where, b.body, err)
			}
		}
	}
	return nil
}

// ownScope is the scope t renders a frame of, nil where that scope binds no name: a category's,
// on its root, or a repeat's, per iteration.
func (t *template) ownScope() *nameScope {
	if sc := t.nameScope; sc != nil && sc.owner == node(t) && len(sc.order) > 0 {
		return sc
	}
	return nil
}

// eachContained calls fn for each node n contains, with where naming the fields reaching it as
// compile errors do.
func eachContained(n node, where string, fn func(c node, where string) error) error {
	for _, c := range contained(n) {
		next := where
		if c.name != "" {
			next += fmt.Sprintf("field %q: ", c.name)
		}
		if err := fn(c.node, next); err != nil {
			return err
		}
	}
	return nil
}

// resolveReads answers each unbound read under root with a name its template sees, fields before
// the format as compile reports them, returning every binding's reads, and refuses a field spelling
// a name. scopes are every scope of the category, which a refusal searches for a name bound where
// the read cannot see it.
func resolveReads(root node, scopes []*nameScope) (map[*nameBinding][]nameUse, error) {
	reads := map[*nameBinding][]nameUse{}
	var resolve func(n node, where string) error
	resolve = func(n node, where string) error {
		t, isTemplate := n.(*template)
		if err := eachContained(n, where, resolve); err != nil {
			return err
		}
		if !isTemplate {
			return nil
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
			b := t.nameScope.lookup(u.head)
			if b == nil {
				return unresolved(where, u, t.nameScope, scopes)
			}
			reads[b] = append(reads[b], nameUse{tail: u.tail, in: t, operand: u.operand, noRef: u.noRef, nested: t.nameScope != b.scope})
		}
		return nil
	}
	return reads, resolve(root, "")
}

// unresolved is the refusal of u, a read no field or name answers: naming the repeat binding the
// name where the read cannot see it, else the names the read sees.
func unresolved(where string, u unboundRead, seen *nameScope, scopes []*nameScope) error {
	for _, sc := range scopes {
		if b, ok := sc.bindings[u.head]; ok {
			at := strings.TrimSuffix(b.where, ": ")
			if at == "" {
				at = "an item of the root choice"
			}
			return fmt.Errorf("%s%w; name %q is bound at %s, inside a repeat, which a read outside the repeat cannot see; bind it outside the repeat to read one pick on every line", where, u.err, u.head, at)
		}
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

// namedReads calls fn with every read of a name t's format makes, and the token holding it.
func namedReads(t *template, fn func(o *op, a *arm) error) error {
	for i := range t.compiled.ops {
		o := &t.compiled.ops[i]
		for _, reads := range [][]arm{o.arms, o.operands} {
			for j := range reads {
				if reads[j].kind == namedRead {
					if err := fn(o, &reads[j]); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
