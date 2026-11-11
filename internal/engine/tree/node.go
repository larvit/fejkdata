// Package tree holds the nodes of a compiled template tree. A load builds each node with its
// syntax, puts each folder's children, gives each table its siblings, and then writes in, through
// LoadFinds.Finish, the rest of what it found. A node's fields are unexported, so another package
// changes a node only through its methods. A map or slice a method returns is the node's own,
// not a copy, so a caller only reads it.
package tree

import (
	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/grammar"
)

// Node is a compiled element of the namespace tree: a folder, choice, null,
// template, table, column or row.
type Node interface{ isNode() }

// Folder is a namespace of named children, built from a directory of JSON files
// and subdirectories. It has no value of its own: descend into a named child by
// dot path; rendering one is an error.
// A name stands in children or in unloaded, never both; Put and PutUnloaded keep it so.
type Folder struct {
	children map[string]Node
	unloaded map[string]*Indexed
}

func NewFolderOf(children map[string]Node) *Folder { return &Folder{children: children} }

func NewFolder() *Folder { return NewFolderOf(map[string]Node{}) }

func (*Folder) isNode() {}

func (g *Folder) Put(name string, n Node) {
	g.children[name] = n
	delete(g.unloaded, name)
}

func (g *Folder) PutUnloaded(name string, x *Indexed) {
	delete(g.children, name)
	if g.unloaded == nil {
		g.unloaded = map[string]*Indexed{}
	}
	g.unloaded[name] = x
}

func (g *Folder) Children() map[string]Node { return g.children }

func (g *Folder) Unloaded() map[string]*Indexed { return g.unloaded }

// Choice draws one of its items. cum holds cumulative weights for a weighted
// draw; when nil the choice is uniform and selection is O(1). shared is the set of
// relative dot paths every item carries, counting paths inside a repeated template.
type Choice struct {
	items     []Node
	cum       []float64
	shared    map[string]bool
	nameScope *NameScope // where its items look a name up, written by LoadFinds.Finish
}

func (*Choice) isNode() {}

func NewChoice(items []Node, cum []float64, shared map[string]bool) *Choice {
	return &Choice{items: items, cum: cum, shared: shared}
}

func (c *Choice) Cum() []float64 { return c.cum }

func (c *Choice) Items() []Node { return c.items }

func (c *Choice) NameScope() *NameScope { return c.nameScope }

func (c *Choice) Shared() map[string]bool { return c.shared }

// NullItem is a column's missing value, rendered ""; sized so two nulls are two map keys.
type NullItem struct{ _ byte }

func (*NullItem) isNode() {}

// Template renders a format string, substituting {tokens} from fields. A bare
// JSON string is a template with no fields. repeat (default 1) renders that format
// that many times and joins the results with separator (default ""), each render
// an independent draw.
type Template struct {
	syntax TemplateSyntax
	finds  *TemplateFinds // written by LoadFinds.Finish, from what its load found
}

func (t *Template) Columns() []RecordColumn { return t.finds.columns }

func (t *Template) Compiled() FormatOps { return t.finds.compiled }

func (t *Template) DataType() *datatype.DataType { return t.syntax.DataType }

func (t *Template) Fields() map[string]Node { return t.syntax.Fields }

func (t *Template) Format() string { return t.syntax.Format }

func (t *Template) FromString() bool { return t.syntax.FromString }

func (t *Template) IsRecord() bool { return t.syntax.IsRecord }

func (t *Template) Lets() map[string]bool { return t.syntax.Lets }

func (t *Template) NameScope() *NameScope { return t.finds.nameScope }

func (t *Template) ReadsColumn() *ColumnRead { return t.finds.readsColumn }

func (t *Template) Refs() TemplateRefs { return t.finds.refs }

func (t *Template) Repeat() int { return t.syntax.Repeat }

func (t *Template) Separator() string { return t.syntax.Separator }

func (t *Template) Tokens() []grammar.Token { return t.syntax.Tokens }

func (t *Template) Unbound() []UnboundRead { return t.syntax.Unbound }

// TemplateSyntax is what a template compiles from: a JSON string, a JSON object or a table's
// row format.
type TemplateSyntax struct {
	Format     string
	Tokens     []grammar.Token
	Fields     map[string]Node
	Lets       map[string]bool // the fields its "let" holds, which write no column
	Repeat     int
	Separator  string
	DataType   *datatype.DataType // the "datatype" it carries, which wins over that of a column it reads; nil where none
	FromString bool               // written as a JSON string rather than an object
	IsRecord   bool               // compiled at the top or as a column without a repeat, so its fields are columns
	Unbound    []UnboundRead      // the heads its tokens read that no field holds
}

func NewTemplate(s TemplateSyntax) *Template { return &Template{syntax: s} }

// Finished says whether LoadFinds.Finish has written what t's load found.
func (t *Template) Finished() bool { return t.finds != nil }

// TemplateRefs is what a template resolves to in the assembled tree.
type TemplateRefs struct {
	bySpelling map[string]ResolvedRef // each reference the format reads -> what it resolves to
	categories map[string]Node        // each ResolvedRef.head -> the category it names
}

func NewTemplateRefs(bySpelling map[string]ResolvedRef, categories map[string]Node) TemplateRefs {
	return TemplateRefs{bySpelling, categories}
}

func (t TemplateRefs) BySpelling() map[string]ResolvedRef { return t.bySpelling }

func (t TemplateRefs) Categories() map[string]Node { return t.categories }

func (*Template) isNode() {}

// StartOf is the node an arm's head names: a sibling field, or a reference's category.
func (t *Template) StartOf(name string, refs TemplateRefs) Node {
	if grammar.IsRef(name) {
		return refs.categories[name]
	}
	return t.syntax.Fields[name]
}

// FixedText is one render's output when the format holds no token; repeat is the caller's.
func (t *Template) FixedText() (string, bool) {
	switch {
	case len(t.syntax.Tokens) == 0:
		return "", true
	case len(t.syntax.Tokens) == 1 && t.syntax.Tokens[0].Kind == grammar.LiteralRun:
		return t.syntax.Tokens[0].Lit, true
	}
	return "", false
}

// OwnScope is the scope t renders a frame of: a category's, once per render of its root, a
// repeat's, once per iteration, or a choice item's, once per draw of the item. It is nil where t owns no scope, or its scope binds no name.
func (t *Template) OwnScope() *NameScope {
	if sc := t.finds.nameScope; sc != nil && sc.owner == Node(t) && len(sc.order) > 0 {
		return sc
	}
	return nil
}
