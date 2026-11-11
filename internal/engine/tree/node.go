// Package tree holds the nodes of a compiled template tree. Compile builds each node with its
// syntax, a load puts each folder's children, linkTables gives each table its siblings, and
// LoadFinds.Finish writes in once the rest of what the load found. A node's fields are
// unexported, so another package changes a node only through those methods.
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

// Choice draws one of its items. cum holds cumulative weights for a weighted
// draw; when nil the choice is uniform and selection is O(1). shared is the set of
// relative dot paths every item carries, counting paths inside a repeated template.
type Choice struct {
	items     []Node
	cum       []float64
	shared    map[string]bool
	nameScope *NameScope // where its items look a name up, written by finish
}

func (*Choice) isNode() {}

func NewChoice(items []Node, cum []float64, shared map[string]bool) *Choice {
	return &Choice{items: items, cum: cum, shared: shared}
}

// NullItem is a column's missing value, rendered ""; sized so two nulls are two map keys.
type NullItem struct{ _ byte }

func (*NullItem) isNode() {}

// Template renders a format string, substituting {tokens} from fields. A bare
// JSON string is a template with no fields. repeat (default 1) renders that format
// that many times and joins the results with separator (default ""), each render
// an independent draw. Every format compiles in `compileTemplates`.
type Template struct {
	syntax TemplateSyntax
	finds  *TemplateFinds // written by finish, from what its load found
}

// TemplateSyntax is what `compileString`, `compileTemplate` and `compileRowFormat` build a
// template from.
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

// Finished says whether finish has written what t's load found.
func (t *Template) Finished() bool { return t.finds != nil }

// TemplateRefs is what a template resolves to in the assembled tree.
type TemplateRefs struct {
	bySpelling map[string]ResolvedRef // each reference the format reads -> what it resolves to
	categories map[string]Node        // each ResolvedRef.head -> the category it names
}

func NewTemplateRefs(bySpelling map[string]ResolvedRef, categories map[string]Node) TemplateRefs {
	return TemplateRefs{bySpelling, categories}
}

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

func (r *Choice) Cum() []float64 { return r.cum }

func (r *Choice) Items() []Node { return r.items }

func (r *Choice) NameScope() *NameScope { return r.nameScope }

func (r *Choice) Shared() map[string]bool { return r.shared }

func (r *Folder) Children() map[string]Node { return r.children }

func (r *Folder) Unloaded() map[string]*Indexed { return r.unloaded }

func (r *Template) Columns() []RecordColumn { return r.finds.columns }

func (r *Template) Compiled() FormatOps { return r.finds.compiled }

func (r *Template) DataType() *datatype.DataType { return r.syntax.DataType }

func (r *Template) Fields() map[string]Node { return r.syntax.Fields }

func (r *Template) Format() string { return r.syntax.Format }

func (r *Template) FromString() bool { return r.syntax.FromString }

func (r *Template) IsRecord() bool { return r.syntax.IsRecord }

func (r *Template) Lets() map[string]bool { return r.syntax.Lets }

func (r *Template) NameScope() *NameScope { return r.finds.nameScope }

func (r *Template) ReadsColumn() *ColumnRead { return r.finds.readsColumn }

func (r *Template) Refs() TemplateRefs { return r.finds.refs }

func (r *Template) Repeat() int { return r.syntax.Repeat }

func (r *Template) Separator() string { return r.syntax.Separator }

func (r *Template) Tokens() []grammar.Token { return r.syntax.Tokens }

func (r *Template) Unbound() []UnboundRead { return r.syntax.Unbound }

func (r TemplateRefs) BySpelling() map[string]ResolvedRef { return r.bySpelling }

func (r TemplateRefs) Categories() map[string]Node { return r.categories }
