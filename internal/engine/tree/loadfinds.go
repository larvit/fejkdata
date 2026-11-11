package tree

import (
	"github.com/larvit/fejkdata/internal/invariant"
)

// LoadFinds holds what one load finds for its templates, choices and name bindings. bindNames
// records each template's and choice's name scope; compileTemplates what each template reads
// and draws, and what each binding resolves to; settleRecords each record's columns. The
// stages before finish read them here, and finish writes them into the nodes.
type LoadFinds struct {
	templates map[*Template]*TemplateFinds
	choices   map[*Choice]*NameScope
	bindings  map[*NameBinding]*BindingFinds
}

func NewLoadFinds() *LoadFinds {
	return &LoadFinds{templates: map[*Template]*TemplateFinds{}, choices: map[*Choice]*NameScope{}, bindings: map[*NameBinding]*BindingFinds{}}
}

// TemplateFinds is what a load found for a template.
type TemplateFinds struct {
	nameScope   *NameScope // where its tokens look a name up
	refs        TemplateRefs
	compiled    FormatOps
	readsColumn *ColumnRead    // the column its format only reads, nil where it reads more
	columns     []RecordColumn // a record's, where it has fields, in name order
}

// BindingFinds is what a load resolved a name binding to.
type BindingFinds struct {
	target NameTarget
	// addressed maps each key a read of the name lands on or passes to the spelling of the
	// first read reaching that key. A pick keeps the draws at these keys only.
	addressed map[PickKey]string
	values    *ValueDraw // what a {ref where … as name} or {ref distinct N as name} draws from; nil without either clause
}

// Of is what l holds for t, else what an earlier load wrote into t. A nil l reads only
// what was written.
func (l *LoadFinds) Of(t *Template) *TemplateFinds {
	if l != nil {
		if tl, loading := l.templates[t]; loading {
			return tl
		}
	}
	return t.finds
}

// template is what l holds for t, which must belong to l's load; it makes the entry on first use.
func (l *LoadFinds) template(t *Template) *TemplateFinds {
	tl, held := l.templates[t]
	if !held {
		tl = &TemplateFinds{}
		l.templates[t] = tl
	}
	return tl
}

// Binding is what l holds for b, else what an earlier load wrote into b. A nil l reads
// only what was written.
func (l *LoadFinds) Binding(b *NameBinding) *BindingFinds {
	if l != nil {
		if bl, loading := l.bindings[b]; loading {
			return bl
		}
	}
	return b.finds
}

// SetNameScope records sc as the scope t's tokens look a name up in.
func (l *LoadFinds) SetNameScope(t *Template, sc *NameScope) { l.template(t).nameScope = sc }

// SetChoiceScope records sc as the scope c's items look a name up in.
func (l *LoadFinds) SetChoiceScope(c *Choice, sc *NameScope) { l.choices[c] = sc }

// SetRefs records what t's references resolve to.
func (l *LoadFinds) SetRefs(t *Template, refs TemplateRefs) { l.template(t).refs = refs }

// SetCompiled records t's compiled format.
func (l *LoadFinds) SetCompiled(t *Template, ops FormatOps) { l.template(t).compiled = ops }

// SetOp replaces op i of t's compiled format with o, its call and draw built.
func (l *LoadFinds) SetOp(t *Template, i int, o Op) { l.template(t).compiled.ops[i] = o }

// SetReadsColumn records the column t's format only reads; nil where it reads more.
func (l *LoadFinds) SetReadsColumn(t *Template, c *ColumnRead) { l.template(t).readsColumn = c }

// SetColumns records a record's columns, in name order.
func (l *LoadFinds) SetColumns(t *Template, columns []RecordColumn) { l.template(t).columns = columns }

// SetTarget records what b resolves to, the first find of a binding.
func (l *LoadFinds) SetTarget(b *NameBinding, target NameTarget) {
	l.bindings[b] = &BindingFinds{target: target}
}

// SetAddressed records the keys a read of b lands on or passes, each with the spelling of the
// first read reaching it.
func (l *LoadFinds) SetAddressed(b *NameBinding, keys map[PickKey]string) {
	l.bindings[b].addressed = keys
}

// SetValues records what b's where or distinct clause draws from.
func (l *LoadFinds) SetValues(b *NameBinding, v *ValueDraw) { l.bindings[b].values = v }

// Finish writes what l holds into its templates, choices and bindings. Each is written
// once, by the load that found it, so finish panics on a template written before; every
// choice and binding a load finds sits in a template of that load.
func (l *LoadFinds) Finish() {
	for t, tl := range l.templates {
		if t.finds != nil {
			panic(invariant.Broken("template %q is finished twice", t.syntax.Format))
		}
		t.finds = tl
	}
	for c, sc := range l.choices {
		c.nameScope = sc
	}
	for b, bl := range l.bindings {
		b.finds = bl
	}
}

func (r *BindingFinds) Target() NameTarget { return r.target }

func (r *TemplateFinds) Compiled() FormatOps { return r.compiled }

func (r *TemplateFinds) NameScope() *NameScope { return r.nameScope }

func (r *TemplateFinds) ReadsColumn() *ColumnRead { return r.readsColumn }

func (r *TemplateFinds) Refs() TemplateRefs { return r.refs }
