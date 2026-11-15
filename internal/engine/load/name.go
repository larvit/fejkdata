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
	g := scopeGatherer{root: root, l: l}
	if err := g.gather(root, nil, false, ""); err != nil {
		return err
	}
	if err := refuseShadowed(g.scopes[1:]); err != nil {
		return err
	}
	return answerReads(root, g.scopes, l)
}

// scopeGatherer collects, in scopes, every scope it opens under root.
type scopeGatherer struct {
	root   tree.Node
	l      *tree.LoadFinds
	scopes []*tree.NameScope
}

// gather binds the names of n and of every node below it. scope is the scope around n, and item
// says n is an item of a choice.
func (g *scopeGatherer) gather(n tree.Node, scope *tree.NameScope, item bool, where string) error {
	if g.opensScope(n, item) {
		scope = tree.NewNameScope(scope, n)
		g.scopes = append(g.scopes, scope)
	}
	switch t := n.(type) {
	case *tree.Template:
		g.l.SetNameScope(t, scope)
		if err := bindAll(scope, t, where); err != nil {
			return err
		}
	case *tree.Choice:
		g.l.SetChoiceScope(t, scope)
	}
	_, isChoice := n.(*tree.Choice)
	return eachContained(n, where, func(c ContainedNode, where string) error { return g.gather(c.Node, scope, isChoice, where) })
}

func (g *scopeGatherer) opensScope(n tree.Node, item bool) bool {
	t, isTemplate := n.(*tree.Template)
	return n == g.root || isTemplate && (t.Repeat() > 1 || item)
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
	return readAnswerer{scopes: scopes, l: l}.answer(root, "", besideField{})
}

type readAnswerer struct {
	scopes []*tree.NameScope
	l      *tree.LoadFinds
}

// answer answers the reads below n, then, if n is a template, n's own. up locates n from above.
func (a readAnswerer) answer(n tree.Node, where string, up besideField) error {
	err := eachContained(n, where, func(c ContainedNode, where string) error {
		return a.answer(c.Node, where, a.below(n, c.name, up))
	})
	t, isTemplate := n.(*tree.Template)
	if err != nil || !isTemplate {
		return err
	}
	sc := a.l.Of(t).NameScope()
	if err := refuseFieldNamed(t, sc, where); err != nil {
		return err
	}
	for _, u := range t.Unbound() {
		if sc.Lookup(u.Head()) == nil {
			return unanswered(where, u, sc, a.scopes, up)
		}
	}
	return nil
}

// below locates a node in the field named field of n, given up, which locates n.
func (a readAnswerer) below(n tree.Node, field string, up besideField) besideField {
	switch n := n.(type) {
	case *tree.Template:
		return besideField{t: n, scope: a.l.Of(n).NameScope(), field: field}
	case *tree.Choice:
		up.inItem = true
	}
	return up
}

// refuseFieldNamed refuses a field of t spelling a name that sc sees.
func refuseFieldNamed(t *tree.Template, sc *tree.NameScope, where string) error {
	for _, name := range tree.SortedNames(t.Fields()) {
		b := sc.Lookup(name)
		if b == nil {
			continue
		}
		holder := "the root template"
		if where != "" {
			holder = strings.TrimSuffix(where, ": ")
		}
		return fmt.Errorf("%stoken {%s}: name %q is a field of %s too; options, fields and names share one namespace, so rename one", b.Where(), b.Body(), name, holder)
	}
	return nil
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
	return fmt.Errorf("%s%w", where, u.Refusal())
}

// outOfSight refuses u, which reads the name b binds inside sc. sc is a repeat or a choice's item,
// and u sits outside it.
func outOfSight(where string, u tree.UnboundRead, sc *tree.NameScope, b *tree.NameBinding) error {
	at := strings.TrimSuffix(b.Where(), ": ")
	if at == "" {
		at = "an item of the root choice"
	}
	if sc.Kind() == tree.ScopeChoiceItem {
		return fmt.Errorf("%s%w; name %q is bound at %s, inside an item of a choice, which a read outside that item cannot see; bind it outside the choice", where, u.Refusal(), u.Head(), at)
	}
	return fmt.Errorf("%s%w; name %q is bound at %s, inside a repeat, which a read outside the repeat cannot see; bind it outside the repeat to read one pick on every line", where, u.Refusal(), u.Head(), at)
}

// visibleNames is every name that seen or a scope around it binds, sorted and quoted.
func visibleNames(seen *tree.NameScope) []string {
	var visible []string
	for sc := seen; sc != nil; sc = sc.Up() {
		for _, b := range sc.Order() {
			visible = append(visible, b.Name())
		}
	}
	sort.Strings(visible)
	for i, name := range visible {
		visible[i] = strconv.Quote(name)
	}
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

// NamedReads is every read of a name a compiled format makes, in format order.
func NamedReads(compiled tree.FormatOps) []namedReadAt {
	ops := compiled.Ops()
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

// ReadsOfNames is every read of each name in ts, in the order ts lists them. compiled is
// LoadFinds.Compiled while a load runs, Template.Compiled once it has finished.
func ReadsOfNames(ts []TemplateSite, compiled func(*tree.Template) tree.FormatOps) map[*tree.NameBinding][]*tree.Read {
	reads := map[*tree.NameBinding][]*tree.Read{}
	for _, s := range ts {
		for _, r := range NamedReads(compiled(s.Template)) {
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
