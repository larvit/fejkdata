package fejkdata

import (
	"fmt"
	"strings"
)

// nameScope is where a name is bound: a category, or one iteration of a repeat inside it.
// Every token of the category outside a nested repeat reads the category's names, so one
// field may bind a name and another read it.
type nameScope struct {
	up       *nameScope
	owner    node // the category's root, or the repeat
	bindings map[string]*nameBinding
	order    []*nameBinding
}

// nameBinding is one {ref as name} token: the scope it binds name in, its index among
// that scope's bindings, and, once linked, the head node and tail its reference reads.
type nameBinding struct {
	name   string
	ref    string
	body   string
	where  string // the fields reaching the binding template, as a compile error spells them
	scope  *nameScope
	index  int
	binder *template
	read   bool
	head   node
	tail   []string
}

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
// nothing reads.
func bindNames(root node) error {
	top := &nameScope{owner: root}
	scopes := []*nameScope{top}
	var gather func(n node, scope *nameScope, inChoice bool, where string) error
	gather = func(n node, scope *nameScope, inChoice bool, where string) error {
		if t, isTemplate := n.(*template); isTemplate {
			if t.repeat > 1 && n != root {
				scope = &nameScope{up: scope, owner: t}
				scopes = append(scopes, scope)
				inChoice = false
			}
			t.scope = scope
			if err := scope.bindAll(t, inChoice, where); err != nil {
				return err
			}
		}
		_, isChoice := n.(*choice)
		return eachContained(n, where, func(c node, where string) error { return gather(c, scope, inChoice || isChoice, where) })
	}
	if err := gather(root, top, false, ""); err != nil {
		return err
	}
	for _, sc := range scopes[1:] {
		for _, b := range sc.order {
			if outer := sc.up.lookup(b.name); outer != nil {
				return fmt.Errorf("%stoken {%s}: name %q is bound outside this repeat already, by {%s}; rename one", b.where, b.body, b.name, outer.body)
			}
		}
	}
	if err := resolveReads(root, ""); err != nil {
		return err
	}
	for _, sc := range scopes {
		if err := sc.close(); err != nil {
			return err
		}
	}
	return nil
}

// bindAll binds every name t's tokens bind. A choice's item binds none: a read outside the item
// would find the name bound by an item that may never be drawn.
func (sc *nameScope) bindAll(t *template, inChoice bool, where string) error {
	for _, tok := range t.tokens {
		if tok.kind != nameBind {
			continue
		}
		if inChoice {
			return fmt.Errorf("%stoken {%s}: a choice's item binds no name, since every other item would leave it unbound; bind it outside the choice", where, tok.body)
		}
		if b, twice := sc.bindings[tok.bound]; twice {
			return fmt.Errorf("%stoken {%s}: name %q is bound twice %s, by {%s} too; rename one", where, tok.body, tok.bound, sc.spelled(), b.body)
		}
		b := &nameBinding{name: tok.bound, ref: tok.names[0], body: tok.body, where: where, scope: sc, index: len(sc.order), binder: t}
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
	if sc.up == nil {
		return "outside any repeat"
	}
	return "in one repeat"
}

// close refuses a binding nothing reads, and gives the owner of a scope binding names the scope.
func (sc *nameScope) close() error {
	for _, b := range sc.order {
		if !b.read {
			return fmt.Errorf("%stoken {%s}: nothing reads name %q; drop the token", b.where, b.body, b.name)
		}
	}
	if t, isTemplate := sc.owner.(*template); isTemplate && len(sc.order) > 0 {
		t.ownScope = sc
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

// resolveReads answers each unbound read under n with a name its template sees, fields before the
// format as compile reports them, and refuses a field spelling a name.
func resolveReads(n node, where string) error {
	if err := eachContained(n, where, resolveReads); err != nil {
		return err
	}
	t, isTemplate := n.(*template)
	if !isTemplate {
		return nil
	}
	for _, name := range sortedNames(t.fields) {
		if b := t.scope.lookup(name); b != nil {
			holder := "the root template"
			if where != "" {
				holder = strings.TrimSuffix(where, ": ")
			}
			return fmt.Errorf("%stoken {%s}: name %q is a field of %s too; options, fields and names share one namespace, so rename one", b.where, b.body, name, holder)
		}
	}
	for _, u := range t.unbound {
		b := t.scope.lookup(u.head)
		if b == nil {
			return fmt.Errorf("%s%w", where, u.err)
		}
		b.read = true
	}
	return nil
}

// isName reports whether a head t reads is a name: neither a reference nor a field.
func (t *template) isName(head string) bool {
	return !isRef(head) && t.fields[head] == nil && t.scope.lookup(head) != nil
}

// linkNames compiles t's reads of a name as paths from the head its binding's reference names, once
// every template is linked, so the binder has resolved that reference.
func linkNames(path string, t *template) error {
	for i := range t.compiled.ops {
		o := &t.compiled.ops[i]
		for _, reads := range [][]arm{o.arms, o.operands} {
			for j := range reads {
				if reads[j].kind != namedRead {
					continue
				}
				if err := linkName(t, &reads[j]); err != nil {
					return fmt.Errorf("%s: token {%s}: %w", t.site.label(path), o.body, err)
				}
			}
		}
	}
	return nil
}

func linkName(t *template, a *arm) error {
	b := t.scope.lookup(a.head)
	if b.head == nil {
		ref := b.binder.link.refs[b.ref]
		b.head, b.tail = b.binder.link.refHeads[ref.head], ref.tail
	}
	if hasSelector(a.tail) {
		return fmt.Errorf("a path through name %q selects no row; select it in the reference %q binds", a.head, a.head)
	}
	full := append(b.tail[:len(b.tail):len(b.tail)], a.tail...)
	if err := checkPathResolves(b.head, full, a.head); err != nil {
		return err
	}
	w := compilePath(b.head, full)
	a.bind, a.steps, a.leaves, a.cover = b, w.steps, w.leaves, w.cover
	a.levels = make([]string, len(full)+1)
	for i := range a.levels {
		a.levels[i] = joinSegments(full[:i])
	}
	a.path = joinSegments(full)
	return nil
}

// namedPick is one draw of a name: the variant drawn at each level under it, the value each
// path through it read, keyed by the path from the name, and the table rows they pinned.
type namedPick struct {
	memo drawMemo
	pins pinSet
}

// pickFrame is one render of a name scope: a pick per binding, each drawn on its first read.
type pickFrame struct {
	scope *nameScope
	picks []namedPick
}

func newPickFrame(scope *nameScope) *pickFrame {
	return &pickFrame{scope: scope, picks: make([]namedPick, len(scope.order))}
}

// frameStack is the frames of the name scopes rendering, innermost last. A render's draws point
// at it, and a repeat iteration's share it: holding the frames in the draws, or in a renderScope,
// would move every render's draws to the heap.
type frameStack struct {
	frames []*pickFrame
}

// pushFrame opens f until popFrame closes it, returning the mark popFrame takes.
func (d *renderDraws) pushFrame(f *pickFrame) int {
	mark := d.depth()
	if d.frames == nil {
		d.frames = &frameStack{}
	}
	d.frames.frames = append(d.frames.frames, f)
	return mark
}

func (d *renderDraws) popFrame(mark int) { d.frames.frames = d.frames.frames[:mark] }

func (d *renderDraws) depth() int {
	if d.frames == nil {
		return 0
	}
	return len(d.frames.frames)
}

// entering is sc as a read entering another category sees it: none of the frames rendering, and
// memo, which keeps the frame of each scope the read enters past its owner.
func (sc renderScope) entering(memo *drawMemo) renderScope {
	sc.base, sc.entry, sc.pick = sc.draws.depth(), memo, nil
	return sc
}

// frameOf is the frame a render of scope's owner opens: the one the read entering the category
// keeps, so {n} and {n.path} read one frame, else a fresh one.
func (sc renderScope) frameOf(scope *nameScope) *pickFrame {
	if sc.entry == nil {
		return newPickFrame(scope)
	}
	return sc.entry.frameOf(scope)
}

func (m *drawMemo) frameOf(scope *nameScope) *pickFrame {
	f, ok := m.frames[scope]
	if !ok {
		f = newPickFrame(scope)
		if m.frames == nil {
			m.frames = map[*nameScope]*pickFrame{}
		}
		m.frames[scope] = f
	}
	return f
}

// pickOf is b's pick: in the frame rendering its scope since the read entering the category, else
// in the frame that read keeps, as for Fake("cat.field") or {/cat.field}.
func (sc renderScope) pickOf(b *nameBinding) *namedPick {
	if stack := sc.draws.frames; stack != nil {
		for i := len(stack.frames) - 1; i >= sc.base; i-- {
			if f := stack.frames[i]; f.scope == b.scope {
				return &f.picks[b.index]
			}
		}
	}
	memo := sc.entry
	if memo == nil {
		memo = &sc.groupDraws().memo
	}
	return &memo.frameOf(b.scope).picks[b.index]
}

func readName(s *generatorState, sc renderScope, a arm) readValue {
	p := sc.pickOf(a.bind)
	return readPicked(s, p, a.bind.head, a.steps, a.levels, a.path, sc.entering(&p.memo))
}

// drawRowOf draws the row a render of t reads: inside the pick's rows where t renders as part of one.
func (sc renderScope) drawRowOf(s *generatorState, t *table) int {
	if sc.pick != nil {
		return t.drawIn(s, &sc.pick.pins)
	}
	return t.drawRow(s)
}

// readUnder reads a of t, which renders as part of the pick sc.pick at sc.pickKey: once per pick,
// by its path from the name.
func readUnder(s *generatorState, t *template, sc renderScope, a arm) readValue {
	levels := make([]string, len(a.levels))
	for i, l := range a.levels {
		levels[i] = underKey(sc.pickKey, l)
	}
	return readPicked(s, sc.pick, t.head(a.head), a.steps, levels, underKey(sc.pickKey, a.path), sc)
}

// underKey is the key of path under the pick key prefix. It never returns prefix itself, which a
// memo would keep: anything a renderScope holds reaching the heap moves every render's draws there.
func underKey(prefix, path string) string {
	if prefix == "" {
		return path
	}
	return prefix + "." + path
}

// readPicked draws the path key names under pick p, once, and renders what it lands on as part of p.
func readPicked(s *generatorState, p *namedPick, head node, steps []pathStep, levels []string, key string, sc renderScope) readValue {
	if r, done := p.memo.value[key]; done {
		return r
	}
	leaf := drawSteps(s, head, steps, &p.pins, &p.memo, levels)
	if c, isChoice := leaf.(*choice); isChoice {
		leaf = p.memo.variantOf(s, c, key)
	}
	sc = sc.at(leaf, &p.pins)
	sc.pick, sc.pickKey = p, key
	r := renderLeaf(s, leaf, sc)
	if p.memo.value == nil {
		p.memo.value = map[string]readValue{}
	}
	p.memo.value[key] = r
	return r
}

// variantOf is the variant of c drawn at level, drawn now where none was.
func (m *drawMemo) variantOf(s *generatorState, c *choice, level string) node {
	n, drew := m.variant[level]
	if !drew {
		n = resolveChoice(s, c)
		if m.variant == nil {
			m.variant = map[string]node{}
		}
		m.variant[strings.Clone(level)] = n
	}
	return n
}
