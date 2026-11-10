package fejkdata

import "github.com/larvit/fejkdata/internal/invariant"

// links is what one load finds for the templates, choices and names it loads: bindNames
// gives each its scope, resolveTemplates what it reads and draws, and settleRecords each
// record's columns. Every stage reads them here, and finish writes them into the nodes
// once the load passed every check before it.
type links struct {
	templates map[*template]*templateLinks
	choices   map[*choice]*nameScope
	bindings  map[*nameBinding]*bindingLinks
}

func newLinks() *links {
	return &links{templates: map[*template]*templateLinks{}, choices: map[*choice]*nameScope{}, bindings: map[*nameBinding]*bindingLinks{}}
}

// templateLinks is what a load found for a template.
type templateLinks struct {
	nameScope   *nameScope // where its tokens look a name up
	refs        templateRefs
	compiled    formatOps
	readsColumn *columnRead    // the column its format only reads, nil where it reads more; read only in columnread.go
	columns     []recordColumn // a record's, where it has fields, in name order
}

// bindingLinks is what a load resolved a name binding to.
type bindingLinks struct {
	target nameTarget
	// addressed maps each key a read of the name lands on or passes to the spelling of the
	// first read reaching that key. A pick keeps the draws at these keys only.
	addressed map[pickKey]string
	values    *valueDraw // what a {ref where … as name} or {ref distinct N as name} draws from; nil without either clause
}

// of is what l holds for t, else what an earlier load wrote into t. A nil l reads only
// what was written.
func (l *links) of(t *template) *templateLinks {
	if l != nil {
		if tl, loading := l.templates[t]; loading {
			return tl
		}
	}
	return t.templateLinks
}

// template is what l holds for t, a template of its own load, made on first use.
func (l *links) template(t *template) *templateLinks {
	tl, held := l.templates[t]
	if !held {
		tl = &templateLinks{}
		l.templates[t] = tl
	}
	return tl
}

// binding is what l holds for b, else what an earlier load wrote into b. A nil l reads
// only what was written.
func (l *links) binding(b *nameBinding) *bindingLinks {
	if l != nil {
		if bl, loading := l.bindings[b]; loading {
			return bl
		}
	}
	return b.bindingLinks
}

// finish writes what l holds into its templates, choices and bindings. Each is written
// once, by the load that found it, so finish panics on a template written before; every
// choice and binding a load finds sits in a template of that load.
func (l *links) finish() {
	for t, tl := range l.templates {
		if t.templateLinks != nil {
			panic(invariant.Broken("template %q is finished twice", t.format))
		}
		t.templateLinks = tl
	}
	for c, sc := range l.choices {
		c.nameScope = sc
	}
	for b, bl := range l.bindings {
		b.bindingLinks = bl
	}
}
