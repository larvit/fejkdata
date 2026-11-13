package load

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

// bindNames records in l, for every template and choice of a compiled category or inline
// template, the scope its names live in. It refuses a name bound twice along one chain of
// scopes, a name a field spells too, and a read no field or name answers.
func bindNames(root tree.Node, l *tree.LoadFinds) error {
	var scopes []*tree.NameScope
	var gather func(n tree.Node, scope *tree.NameScope, item bool, where string) error
	gather = func(n tree.Node, scope *tree.NameScope, item bool, where string) error {
		if t, isTemplate := n.(*tree.Template); n == root || isTemplate && (t.Repeat() > 1 || item) {
			scope = tree.NewNameScope(scope, n)
			scopes = append(scopes, scope)
		}
		switch t := n.(type) {
		case *tree.Template:
			l.SetNameScope(t, scope)
			if err := bindAll(scope, t, where); err != nil {
				return err
			}
		case *tree.Choice:
			l.SetChoiceScope(t, scope)
		}
		_, isChoice := n.(*tree.Choice)
		return eachContained(n, where, func(c ContainedNode, where string) error { return gather(c.Node, scope, isChoice, where) })
	}
	if err := gather(root, nil, false, ""); err != nil {
		return err
	}
	if err := refuseShadowed(scopes[1:]); err != nil {
		return err
	}
	return answerReads(root, scopes, l)
}

// refuseShadowed refuses a name that one of scopes binds and a scope around it binds too.
func refuseShadowed(scopes []*tree.NameScope) error {
	for _, sc := range scopes {
		for _, b := range sc.Order() {
			if outer := sc.Up().Lookup(b.Name()); outer != nil {
				return fmt.Errorf("%stoken {%s}: name %q is bound outside this %s too, by {%s}; rename one", b.Where(), b.Body(), b.Name(), sc.Kind(), outer.Body())
			}
		}
	}
	return nil
}

func bindAll(sc *tree.NameScope, t *tree.Template, where string) error {
	for _, tok := range t.Tokens() {
		if tok.Kind != grammar.NameBind {
			continue
		}
		if b, twice := sc.Bindings()[tok.Bound]; twice {
			return fmt.Errorf("%stoken {%s}: name %q is bound twice %s, by {%s} too; rename one", where, tok.Body, tok.Bound, sc.Spelled(), b.Body())
		}
		sc.Bind(tok, where, t)
	}
	return nil
}

// eachContained calls fn for each node n contains, with where naming the fields reaching it as
// compile errors do.
func eachContained(n tree.Node, where string, fn func(c ContainedNode, where string) error) error {
	for _, c := range Contained(n) {
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

// answerReads proves each unbound read under root is answered by a name its template sees. It
// answers a template's fields before its format, the order Compile reports in, and refuses a
// field spelling a name. A refusal searches scopes, every scope of the category, for a name bound
// where the read cannot see it.
func answerReads(root tree.Node, scopes []*tree.NameScope, l *tree.LoadFinds) error {
	var answer func(n tree.Node, where string, up besideField) error
	answer = func(n tree.Node, where string, up besideField) error {
		t, isTemplate := n.(*tree.Template)
		err := eachContained(n, where, func(c ContainedNode, where string) error {
			inner := up
			switch n.(type) {
			case *tree.Template:
				inner = besideField{t: t, scope: l.Of(t).NameScope(), field: c.name}
			case *tree.Choice:
				inner.inItem = true
			}
			return answer(c.Node, where, inner)
		})
		if err != nil || !isTemplate {
			return err
		}
		sc := l.Of(t).NameScope()
		for _, name := range tree.SortedNames(t.Fields()) {
			if b := sc.Lookup(name); b != nil {
				holder := "the root template"
				if where != "" {
					holder = strings.TrimSuffix(where, ": ")
				}
				return fmt.Errorf("%stoken {%s}: name %q is a field of %s too; options, fields and names share one namespace, so rename one", b.Where(), b.Body(), name, holder)
			}
		}
		for _, u := range t.Unbound() {
			if sc.Lookup(u.Head()) == nil {
				return unanswered(where, u, sc, scopes, up)
			}
		}
		return nil
	}
	return answer(root, "", besideField{})
}

// besideField locates a template from above: t is the nearest template containing it, scope
// is t's, field is the field of t it lies under, and inItem says it is an item of a choice in
// that field.
type besideField struct {
	t      *tree.Template
	scope  *tree.NameScope
	field  string
	inItem bool
}

// refusal is the refusal of u where up's template holds u's head.
func (up besideField) refusal(where string, u tree.UnboundRead) error {
	holder := fmt.Sprintf("field %q", up.field)
	if up.inItem {
		holder = fmt.Sprintf("the item of field %q holding it", up.field)
	}
	fix := fmt.Sprintf("give %s a format and a %q of its own", holder, u.Head())
	if up.scope.Kind() != tree.ScopeChoiceItem {
		fix += fmt.Sprintf(", or bind {%s as n} in the format beside field %q and read n in place of %s", u.Head(), up.field, u.Head())
	}
	return fmt.Errorf("%stoken {%s}: %q is a field beside field %q, and a token reads only the fields of the template whose format holds it: %s", where, u.Body(), u.Head(), up.field, fix)
}

// unanswered is the refusal of u, a read no field or name answers: naming the repeat binding the
// name where the read cannot see it, else the names the read sees.
func unanswered(where string, u tree.UnboundRead, seen *tree.NameScope, scopes []*tree.NameScope, up besideField) error {
	for _, sc := range scopes {
		if b, ok := sc.Bindings()[u.Head()]; ok {
			return outOfSight(where, u, sc, b)
		}
	}
	if up.t != nil && up.t.Fields()[u.Head()] != nil {
		return up.refusal(where, u)
	}
	if visible := visibleNames(seen); len(visible) > 0 {
		return fmt.Errorf("%stoken {%s}: no field or name %q; the names bound here are %s", where, u.Body(), u.Head(), strings.Join(visible, ", "))
	}
	return fmt.Errorf("%s%w", where, u.Err())
}

// outOfSight is the refusal of u, a read of the name b binds inside sc, a repeat or a choice's
// item the read is outside.
func outOfSight(where string, u tree.UnboundRead, sc *tree.NameScope, b *tree.NameBinding) error {
	at := strings.TrimSuffix(b.Where(), ": ")
	if at == "" {
		at = "an item of the root choice"
	}
	if sc.Kind() == tree.ScopeChoiceItem {
		return fmt.Errorf("%s%w; name %q is bound at %s, inside an item of a choice, which a read outside that item cannot see; bind it outside the choice", where, u.Err(), u.Head(), at)
	}
	return fmt.Errorf("%s%w; name %q is bound at %s, inside a repeat, which a read outside the repeat cannot see; bind it outside the repeat to read one pick on every line", where, u.Err(), u.Head(), at)
}

// visibleNames is every name seen and the scopes around it bind, sorted and quoted.
func visibleNames(seen *tree.NameScope) []string {
	var visible []string
	for sc := seen; sc != nil; sc = sc.Up() {
		for _, b := range sc.Order() {
			visible = append(visible, strconv.Quote(b.Name()))
		}
	}
	sort.Strings(visible)
	return visible
}

func isName(t *tree.Template, head string, l *tree.LoadFinds) bool {
	return !grammar.IsRef(head) && t.Fields()[head] == nil && l.Of(t).NameScope().Lookup(head) != nil
}

// namedReadAt is one read of a name in a format: the op holding it, and the compiled read.
type namedReadAt struct {
	Op   *tree.Op
	Read *tree.Read
}

// NamedReads is every read of a name t's format makes, in format order.
func NamedReads(t *tree.Template, l *tree.LoadFinds) []namedReadAt {
	ops := l.Of(t).Compiled().Ops()
	var out []namedReadAt
	for i := range ops {
		o := &ops[i]
		for _, a := range o.AllReads() {
			if a.Kind() == tree.NamedRead {
				out = append(out, namedReadAt{Op: o, Read: a})
			}
		}
	}
	return out
}

// ReadsOfNames is every read of each name in ts, in the order ts lists them.
func ReadsOfNames(ts []TemplateSite, l *tree.LoadFinds) map[*tree.NameBinding][]*tree.Read {
	reads := map[*tree.NameBinding][]*tree.Read{}
	for _, s := range ts {
		for _, r := range NamedReads(s.Template, l) {
			reads[r.Read.Named()] = append(reads[r.Read.Named()], r.Read)
		}
	}
	return reads
}

// nameTargets resolves what each binding of ts binds, from its binder's references.
func nameTargets(ts []TemplateSite, l *tree.LoadFinds) map[*tree.NameBinding]tree.NameTarget {
	targets := map[*tree.NameBinding]tree.NameTarget{}
	for _, s := range ts {
		tl := l.Of(s.Template)
		for _, tok := range s.Template.Tokens() {
			if tok.Kind == grammar.NameBind {
				a := splitRead(tok.BoundRef, tl.Refs().BySpelling())
				targets[tl.NameScope().Bindings()[tok.Bound]] = tree.NewNameTarget(s.Template.StartOf(a.head, tl.Refs()), a.tail)
			}
		}
	}
	return targets
}
