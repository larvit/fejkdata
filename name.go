package fejkdata

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
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
	name   string
	ref    string
	body   string
	where  string // the fields reaching the binding template, as a compile error spells them
	scope  *nameScope
	index  int
	binder *template
	uses   []nameUse
	head   node
	tail   []string
	// addressed is every key a read of the name lands on or passes, the spelling of the
	// first read reaching it beside it; a pick keeps the draws at these keys, and only these.
	addressed map[string]string
}

// nameUse is one read of a name: the path it reads into the name, "" for the name itself; whether
// a builtin reads it as an operand; the draw group its template renders in; and whether it sits in
// a repeat nested inside the name's scope.
type nameUse struct {
	tail, group     string
	operand, nested bool
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
// read by nothing.
func bindNames(root node) error {
	top := &nameScope{owner: root}
	scopes := []*nameScope{top}
	var gather func(n node, scope *nameScope, inChoice bool, where string) error
	gather = func(n node, scope *nameScope, inChoice bool, where string) error {
		switch t := n.(type) {
		case *template:
			if t.repeat > 1 && n != root {
				scope = &nameScope{up: scope, owner: t}
				scopes = append(scopes, scope)
				inChoice = false
			}
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
	if err := gather(root, top, false, ""); err != nil {
		return err
	}
	for _, sc := range scopes[1:] {
		for _, b := range sc.order {
			if outer := sc.up.lookup(b.name); outer != nil {
				return fmt.Errorf("%stoken {%s}: name %q is bound outside this repeat too, by {%s}; rename one", b.where, b.body, b.name, outer.body)
			}
		}
	}
	if err := resolveReads(root, "", "", scopes); err != nil {
		return err
	}
	for _, sc := range scopes {
		if err := sc.settle(); err != nil {
			return err
		}
	}
	return nil
}

// bindAll binds every name t's tokens bind.
func (sc *nameScope) bindAll(t *template, inChoice bool, where string) error {
	for _, tok := range t.tokens {
		if tok.kind != nameBind {
			continue
		}
		if inChoice {
			return fmt.Errorf("%stoken {%s}: a choice's item binds no name, since every other item would leave it unbound; bind it outside the choice, or move the item into a category of its own and reference that", where, tok.body)
		}
		if b, twice := sc.bindings[tok.bound]; twice {
			return fmt.Errorf("%stoken {%s}: name %q is bound twice %s, by {%s} too; rename one", where, tok.body, tok.bound, sc.spelled(), b.body)
		}
		b := &nameBinding{name: tok.bound, ref: tok.boundRef, body: tok.body, where: where, scope: sc, index: len(sc.order), binder: t}
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

// settle refuses a binding read by nothing, and hands a scope binding names to its owner, which
// renders a frame of it.
func (sc *nameScope) settle() error {
	for _, b := range sc.order {
		if len(b.uses) == 0 {
			return fmt.Errorf("%stoken {%s}: nothing reads name %q; drop the token", b.where, b.body, b.name)
		}
	}
	if t, isTemplate := sc.owner.(*template); isTemplate && len(sc.order) > 0 {
		t.ownNameScope = sc
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
// format as compile reports them, and refuses a field spelling a name. scopes are every scope of
// the category, which a refusal searches for a name bound where the read cannot see it.
func resolveReads(n node, where, group string, scopes []*nameScope) error {
	t, isTemplate := n.(*template)
	if isTemplate && t.drawGroup != "" {
		group = t.drawGroup
	}
	if err := eachContained(n, where, func(c node, where string) error { return resolveReads(c, where, group, scopes) }); err != nil {
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
		b.uses = append(b.uses, nameUse{tail: u.tail, group: group, operand: u.operand, nested: t.nameScope != b.scope})
	}
	return nil
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
	return !isRef(head) && t.fields[head] == nil && t.nameScope.lookup(head) != nil
}

// linkBindings gives each binding t's tokens make the head and tail its reference resolved to.
func linkBindings(t *template) {
	for _, tok := range t.tokens {
		if tok.kind != nameBind {
			continue
		}
		b := t.nameScope.bindings[tok.bound]
		ref := t.link.refs[b.ref]
		b.head, b.tail = t.link.refHeads[ref.head], ref.tail
	}
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

// linkNames compiles t's reads of a name as paths from the head its binding's reference names, once
// every template is linked, so each binder has resolved that reference.
func linkNames(path string, t *template) error {
	return namedReads(t, func(o *op, a *arm) error {
		if err := linkName(t, a); err != nil {
			return fmt.Errorf("%s: token {%s}: %w", t.site.label(path), o.body, err)
		}
		return nil
	})
}

func linkName(t *template, a *arm) error {
	b := t.nameScope.lookup(a.head)
	if hasSelector(a.tail) {
		return fmt.Errorf("a path through name %q may not select a row; read it directly, {%s.%s}, or bind the row to a name of its own", a.head, b.ref, joinSegments(a.tail))
	}
	full := append(b.tail[:len(b.tail):len(b.tail)], a.tail...)
	if err := checkPathResolves(b.head, full, a.head); err != nil {
		return err
	}
	w := compilePath(b.head, full)
	a.named, a.steps, a.leaves, a.cover = b, w.steps, w.leaves, w.cover
	a.levels = make([]string, len(full)+1)
	for i := range a.levels {
		a.levels[i] = joinSegments(full[:i])
	}
	a.path = joinSegments(full)
	if b.addressed == nil {
		b.addressed = map[string]string{}
	}
	for _, key := range append(a.levels[len(b.tail):], a.path) {
		if _, seen := b.addressed[key]; !seen {
			b.addressed[key] = a.spelling
		}
	}
	return nil
}

// checkNameReads refuses each binding of t that checkUses refuses, and a read such as {n} whose
// pick renders twice a field another read, {n.w}, reads once: the pick keeps one draw of it.
func checkNameReads(path string, t *template) error {
	for _, tok := range t.tokens {
		if tok.kind != nameBind {
			continue
		}
		if err := t.nameScope.bindings[tok.bound].checkUses(); err != nil {
			return fmt.Errorf("%s: token {%s}: %w", t.site.label(path), tok.body, err)
		}
	}
	return namedReads(t, func(o *op, a *arm) error {
		for _, leaf := range a.leaves {
			if err := a.named.checkOnce(a.spelling, leaf, a.path); err != nil {
				return fmt.Errorf("%s: token {%s}: %w", t.site.label(path), o.body, err)
			}
		}
		return nil
	})
}

// checkUses refuses a binding of a category read once whole, and reads of b in two draw groups,
// or inside a repeat, where what b names reads a reference path.
func (b *nameBinding) checkUses() error {
	if r := b.uses[0]; len(b.uses) == 1 && r.tail == "" && !r.nested && len(b.tail) == 0 {
		spelling := b.ref
		if !r.operand {
			spelling = "{" + spelling + "}"
		}
		return fmt.Errorf("name %q is read once, whole, which the bare reference draws the same way; write %s where it is read, and drop the token", b.name, spelling)
	}
	if !readsHeld(b.head, map[node]bool{}) {
		return nil
	}
	for _, u := range b.uses {
		if u.nested {
			return fmt.Errorf("name %q is read inside a repeat, and what it names reads a reference path, which each iteration draws apart; bind the name inside the repeat, or bind a second name there", b.name)
		}
		if u.group != b.uses[0].group {
			return fmt.Errorf("name %q is read in two draw groups, %s and %s, and what it names reads a reference path, which each draw group draws apart; read the name in one draw group, or bind a name in each", b.name, groupSpelling(b.uses[0].group), groupSpelling(u.group))
		}
	}
	return nil
}

func groupSpelling(group string) string {
	if group == "" {
		return "the unnamed one"
	}
	return fmt.Sprintf("%q", group)
}

// readsHeld reports whether rendering n, or any field under it, can read a reference path.
func readsHeld(n node, seen map[node]bool) bool {
	if seen[n] {
		return false
	}
	seen[n] = true
	if t, isTemplate := n.(*template); isTemplate {
		for _, o := range t.compiled.ops {
			for _, a := range append(o.arms[:len(o.arms):len(o.arms)], o.operands...) {
				if a.kind == refPathRead {
					return true
				}
			}
		}
	}
	for _, c := range contained(n) {
		if readsHeld(c.node, seen) {
			return true
		}
	}
	for _, e := range renderEdges(n) {
		if readsHeld(e.to, seen) {
			return true
		}
	}
	return false
}

// checkOnce walks n, rendering at key under b's pick, into each level a read of b addresses.
func (b *nameBinding) checkOnce(read string, n node, key string) error {
	switch n := n.(type) {
	case *choice:
		for _, item := range n.items {
			if err := b.checkOnce(read, item, key); err != nil {
				return err
			}
		}
	case *template:
		return b.checkTemplateOnce(read, n, key)
	}
	return nil
}

func (b *nameBinding) checkTemplateOnce(read string, t *template, key string) error {
	fresh := map[string]int{}
	var into []arm
	for _, o := range t.compiled.ops {
		for _, a := range append(o.arms[:len(o.arms):len(o.arms)], o.operands...) {
			if isRef(a.head) || a.kind == namedRead {
				continue
			}
			if _, kept := b.addressed[join(key, a.path)]; kept {
				into = append(into, a)
			}
			if a.kind == freshRead {
				fresh[a.head]++
			}
		}
	}
	heads := make([]string, 0, len(fresh))
	for head := range fresh {
		heads = append(heads, head)
	}
	sort.Strings(heads)
	for _, head := range heads {
		if by, kept := b.addressed[join(key, head)]; kept && fresh[head] > 1 {
			return fmt.Errorf("{%s} renders field %q twice, so {%s} cannot say which draw it reads; drop {%s} or {%s}", read, head, by, read, by)
		}
	}
	for _, a := range into {
		for _, leaf := range a.leaves {
			if err := b.checkOnce(read, leaf, join(key, a.path)); err != nil {
				return err
			}
		}
	}
	return nil
}

// namedPick is one draw of a name: the variant drawn at each level a read of it addresses, the
// value each read there produced, keyed by the path from the name, and the table rows they
// pinned. A reference read under it draws in the draw group of the read of the name, which the
// draw fences survey it in.
type namedPick struct {
	named *nameBinding
	memo  drawMemo
	pins  pinSet
}

// pickFrame is one render of a name scope: a pick per binding, each drawn on its first read.
type pickFrame struct {
	scope *nameScope
	picks []namedPick
}

func newPickFrame(scope *nameScope) *pickFrame {
	f := &pickFrame{scope: scope, picks: make([]namedPick, len(scope.order))}
	for i, b := range scope.order {
		f.picks[i].named = b
	}
	return f
}

// frameStack is the frames of the name scopes rendering, innermost last. A render's draws point
// at it, and a repeat iteration's share it: holding the frames in the draws, or in a renderScope,
// would move every render's draws to the heap.
type frameStack struct {
	frames []*pickFrame
}

// pushFrame opens f until popFrames closes it, returning the mark popFrames takes.
func (d *renderDraws) pushFrame(f *pickFrame) int {
	mark := d.depth()
	if d.frameStack == nil {
		d.frameStack = &frameStack{}
	}
	d.frameStack.frames = append(d.frameStack.frames, f)
	return mark
}

func (d *renderDraws) popFrames(mark int) {
	if d.frameStack != nil {
		d.frameStack.frames = d.frameStack.frames[:mark]
	}
}

func (d *renderDraws) depth() int {
	if d.frameStack == nil {
		return 0
	}
	return len(d.frameStack.frames)
}

// enter starts a read landing on n: the read sees no frame opened before it, and gets from memo a
// frame for each name scope around n, so reads sharing memo read one pick of each name. It
// returns the mark that closes those frames.
func (sc renderScope) enter(n node, memo *drawMemo) (renderScope, int) {
	sc.base, sc.pick = sc.draws.depth(), nil
	for scope := scopeAround(n); scope != nil; scope = scope.up {
		if len(scope.order) > 0 {
			sc.draws.pushFrame(memo.enteredFrame(scope))
		}
	}
	return sc, sc.base
}

// scopeAround is the innermost name scope a render of n reads names in, short of the frames n
// renders itself, one per iteration of a repeat.
func scopeAround(n node) *nameScope {
	switch n := n.(type) {
	case *template:
		if n.repeat > 1 && n.nameScope != nil && n.nameScope.owner == n {
			return n.nameScope.up
		}
		return n.nameScope
	case *choice:
		return n.nameScope
	}
	return nil
}

func (m *drawMemo) enteredFrame(scope *nameScope) *pickFrame {
	f, ok := m.enteredFrames[scope]
	if !ok {
		f = newPickFrame(scope)
		if m.enteredFrames == nil {
			m.enteredFrames = map[*nameScope]*pickFrame{}
		}
		m.enteredFrames[scope] = f
	}
	return f
}

// renderFrame opens a fresh frame of t's scope, where t binds names and no read entering the
// category opened one, returning the mark that closes it, or -1.
func (sc renderScope) renderFrame(t *template) int {
	if t.ownNameScope == nil || sc.frameOf(t.ownNameScope) != nil {
		return -1
	}
	return sc.draws.pushFrame(newPickFrame(t.ownNameScope))
}

// frameOf is the frame of scope rendering since the read entering the category, nil where none is.
func (sc renderScope) frameOf(scope *nameScope) *pickFrame {
	if stack := sc.draws.frameStack; stack != nil {
		for i := len(stack.frames) - 1; i >= sc.base; i-- {
			if f := stack.frames[i]; f.scope == scope {
				return f
			}
		}
	}
	return nil
}

func readName(s *generatorState, sc renderScope, a arm) readValue {
	f := sc.frameOf(a.named.scope)
	if f == nil {
		panic(internalError("name %q is read where no frame of its scope renders", a.named.name))
	}
	p := &f.picks[a.named.index]
	if r, done := p.memo.value[a.path]; done {
		return r
	}
	leaf := p.draw(s, a.named.head, a.steps, a.levels, a.path)
	sc, mark := sc.enter(leaf, &p.memo)
	r := p.renderAt(s, leaf, a.path, sc)
	sc.draws.popFrames(mark)
	return r
}

// drawRowOf draws the row a render of t reads: inside the pick's rows where t renders as part of one.
func (sc renderScope) drawRowOf(s *generatorState, t *table) int {
	if sc.pick != nil {
		return t.drawIn(s, &sc.pick.pins)
	}
	return t.drawRow(s)
}

// keeps reports whether a read of sc's name addresses the level a starts at.
func (sc renderScope) keeps(a arm) bool {
	_, kept := sc.pick.named.addressed[underKey(sc.pickKey, a.head)]
	return kept
}

// readUnder reads a of t, which renders as part of the pick sc.pick at sc.pickKey, where a read of
// the name addresses it: once per pick, by its path from the name.
func readUnder(s *generatorState, t *template, sc renderScope, a arm) readValue {
	p, key := sc.pick, underKey(sc.pickKey, a.path)
	if r, done := p.memo.value[key]; done {
		return r
	}
	levels := make([]string, len(a.levels))
	for i, l := range a.levels {
		levels[i] = underKey(sc.pickKey, l)
	}
	return p.renderAt(s, p.draw(s, t.head(a.head), a.steps, levels, key), key, sc)
}

// underKey is the key of path under the pick key prefix. It never returns prefix itself, which a
// memo would keep: anything a renderScope holds reaching the heap moves every render's draws there.
func underKey(prefix, path string) string {
	if prefix == "" {
		return path
	}
	return prefix + "." + path
}

// draw draws the path key names under p, its variant at key kept too.
func (p *namedPick) draw(s *generatorState, head node, steps []pathStep, levels []string, key string) node {
	leaf := drawSteps(s, head, steps, &p.pins, &p.memo, levels)
	if c, isChoice := leaf.(*choice); isChoice {
		leaf = p.memo.variantOf(s, c, key)
	}
	return leaf
}

// renderAt renders leaf, drawn at key, as part of p, and keeps what it rendered.
func (p *namedPick) renderAt(s *generatorState, leaf node, key string, sc renderScope) readValue {
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
