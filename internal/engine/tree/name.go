package tree

import (
	"github.com/larvit/fejkdata/internal/grammar"
)

// NameScope is where a name is bound: a category, one iteration of a repeat inside it, or one
// render of a choice's item. A token sees its scope's names and every enclosing scope's, so one
// field may bind a name and another read it.
type NameScope struct {
	up       *NameScope
	owner    Node // the category's root, the repeat, or the choice's item
	bindings map[string]*NameBinding
	order    []*NameBinding
}

func NewNameScope(up *NameScope, owner Node) *NameScope { return &NameScope{up: up, owner: owner} }

// Bind records in sc the name tok binds, in binder's format; where spells the fields reaching
// binder, as a compile error spells them.
func (sc *NameScope) Bind(tok grammar.Token, where string, binder *Template) {
	b := &NameBinding{name: tok.Bound, ref: tok.BoundRef, body: tok.Body, distinct: tok.Distinct, where: where, scope: sc, index: len(sc.order), binder: binder}
	if sc.bindings == nil {
		sc.bindings = map[string]*NameBinding{}
	}
	sc.bindings[b.name] = b
	sc.order = append(sc.order, b)
}

func (sc *NameScope) Bindings() map[string]*NameBinding { return sc.bindings }

func (sc *NameScope) Order() []*NameBinding { return sc.order }

func (sc *NameScope) Owner() Node { return sc.owner }

func (sc *NameScope) Up() *NameScope { return sc.up }

type NameBinding struct {
	// Set as the load binds the names of its compiled category:
	name     string
	ref      string // what it binds, as written: a reference, or a path into a field of binder
	body     string
	distinct *grammar.Distinct // what a {ref distinct N as name} draws; nil for a single value
	where    string            // the fields reaching the binding template, as a compile error spells them
	scope    *NameScope
	index    int
	binder   *Template

	finds *BindingFinds // written by LoadFinds.Finish, from what its load found
}

// Finished says whether LoadFinds.Finish has written what b's load found.
func (b *NameBinding) Finished() bool { return b.finds != nil }

func (b *NameBinding) Addressed() map[PickKey]string { return b.finds.addressed }

func (b *NameBinding) Binder() *Template { return b.binder }

func (b *NameBinding) Body() string { return b.body }

func (b *NameBinding) Distinct() *grammar.Distinct { return b.distinct }

func (b *NameBinding) Index() int { return b.index }

func (b *NameBinding) Name() string { return b.name }

func (b *NameBinding) Ref() string { return b.ref }

func (b *NameBinding) Scope() *NameScope { return b.scope }

func (b *NameBinding) Target() NameTarget { return b.finds.target }

func (b *NameBinding) Values() *ValueDraw { return b.finds.values }

func (b *NameBinding) Where() string { return b.where }

// NameTarget is what a binding resolves to in the assembled tree: start, the node its head names, and
// the path it reads into that node.
type NameTarget struct {
	start Node
	tail  []string
}

func NewNameTarget(start Node, tail []string) NameTarget { return NameTarget{start, tail} }

func (n NameTarget) Start() Node { return n.start }

func (n NameTarget) Tail() []string { return n.tail }

// BindsField reports whether b binds a path into a field, so a read of it stays in the category.
func (b *NameBinding) BindsField() bool { return !grammar.IsRef(b.ref) }

func (sc *NameScope) Lookup(name string) *NameBinding {
	for ; sc != nil; sc = sc.up {
		if b, ok := sc.bindings[name]; ok {
			return b
		}
	}
	return nil
}

// Within reports whether sc is other or lies inside it, so other's frame is open wherever sc's is.
func (sc *NameScope) Within(other *NameScope) bool {
	for ; sc != nil; sc = sc.up {
		if sc == other {
			return true
		}
	}
	return false
}

func (sc *NameScope) Spelled() string {
	if k := sc.Kind(); k != ScopeCategory {
		return "in one " + k.String()
	}
	return "outside any repeat"
}

// ScopeKind is what a name scope is: the category's own, a repeat's, or a choice's item's.
type ScopeKind uint8

const (
	ScopeCategory ScopeKind = iota
	ScopeRepeat
	ScopeChoiceItem
)

func (k ScopeKind) String() string {
	return [...]string{ScopeCategory: "category", ScopeRepeat: "repeat", ScopeChoiceItem: "choice item"}[k]
}

func (sc *NameScope) Kind() ScopeKind {
	switch t, isTemplate := sc.owner.(*Template); {
	case isTemplate && t.syntax.Repeat > 1:
		return ScopeRepeat
	case sc.up == nil:
		return ScopeCategory
	}
	return ScopeChoiceItem
}
