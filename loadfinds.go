package fejkdata

import "github.com/larvit/fejkdata/internal/invariant"

// loadFinds holds what one load finds for its templates, choices and name bindings. bindNames
// records each template's and choice's name scope; compileTemplates what each template reads
// and draws, and what each binding resolves to; settleRecords each record's columns. The
// stages before finish read them here, and finish writes them into the nodes.
type loadFinds struct {
	templates map[*template]*templateFinds
	choices   map[*choice]*nameScope
	bindings  map[*nameBinding]*bindingFinds
}

func newLoadFinds() *loadFinds {
	return &loadFinds{templates: map[*template]*templateFinds{}, choices: map[*choice]*nameScope{}, bindings: map[*nameBinding]*bindingFinds{}}
}

// templateFinds is what a load found for a template.
type templateFinds struct {
	nameScope   *nameScope // where its tokens look a name up
	refs        templateRefs
	compiled    formatOps
	readsColumn *columnRead    // the column its format only reads, nil where it reads more; read only in columnread.go
	columns     []recordColumn // a record's, where it has fields, in name order
}

// bindingFinds is what a load resolved a name binding to.
type bindingFinds struct {
	target nameTarget
	// addressed maps each key a read of the name lands on or passes to the spelling of the
	// first read reaching that key. A pick keeps the draws at these keys only.
	addressed map[pickKey]string
	values    *valueDraw // what a {ref where … as name} or {ref distinct N as name} draws from; nil without either clause
}

// of is what l holds for t, else what an earlier load wrote into t. A nil l reads only
// what was written.
func (l *loadFinds) of(t *template) *templateFinds {
	if l != nil {
		if tl, loading := l.templates[t]; loading {
			return tl
		}
	}
	return t.templateFinds
}

// template is what l holds for t, which must belong to l's load; it makes the entry on first use.
func (l *loadFinds) template(t *template) *templateFinds {
	tl, held := l.templates[t]
	if !held {
		tl = &templateFinds{}
		l.templates[t] = tl
	}
	return tl
}

// binding is what l holds for b, else what an earlier load wrote into b. A nil l reads
// only what was written.
func (l *loadFinds) binding(b *nameBinding) *bindingFinds {
	if l != nil {
		if bl, loading := l.bindings[b]; loading {
			return bl
		}
	}
	return b.bindingFinds
}

// finish writes what l holds into its templates, choices and bindings. Each is written
// once, by the load that found it, so finish panics on a template written before; every
// choice and binding a load finds sits in a template of that load.
func (l *loadFinds) finish() {
	for t, tl := range l.templates {
		if t.templateFinds != nil {
			panic(invariant.Broken("template %q is finished twice", t.format))
		}
		t.templateFinds = tl
	}
	for c, sc := range l.choices {
		c.nameScope = sc
	}
	for b, bl := range l.bindings {
		b.bindingFinds = bl
	}
}
