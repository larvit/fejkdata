package fejkdata

import (
	"fmt"
	"strings"
)

// rng is the randomness a builtin sample draws from; *rand.Rand satisfies it. The
// render path takes the concrete *generatorState instead, which keeps the draws of the walk
// drawing through it off the heap.
type rng interface {
	IntN(n int) int
	Float64() float64
}

// Fake generates a value for a dot path. Each segment descends one level: folder
// names and the category (JSON file) come first, then named fields within it,
// e.g. "sv_SE.address" or "sv_SE.address.street". Choices along the way are
// resolved at random. A path naming a folder (no value of its own) is an error.
func (f *Generator) Fake(path string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	segments, err := splitPath(path)
	if err != nil {
		return "", fmt.Errorf("fejkdata: %w", err)
	}
	var draws renderDraws
	sc := renderScope{draws: &draws}
	n, err := descend(f.rand, &f.root, segments, sc)
	if err != nil {
		return "", fmt.Errorf("fejkdata: %s: %w", path, err)
	}
	if _, ok := n.(*folder); ok {
		return "", fmt.Errorf("fejkdata: %s names a folder, not a value", path)
	}
	group := sc.groupDraws()
	sc, _ = sc.at(n, &group.pins).enter(n, &group.memo)
	return render(f.rand, n, sc), nil
}

// descend walks a caller's path to the node it names, pinning in sc the rows it
// selects or draws.
func descend(s *generatorState, root node, segments []string, sc renderScope) (node, error) {
	// docs/decisions.md#a-path-is-walked-once-without-drawing-before-it-is-walked-for-real
	var buf [16]pathStep
	steps, err := probePath(root, segments, buf[:0])
	if err != nil {
		return nil, err
	}
	return drawSteps(s, root, steps, &sc.groupDraws().pins, nil, nil), nil
}

// renderOnce renders n as one render, over draws of its own.
func renderOnce(s *generatorState, n node) string {
	var draws renderDraws
	return render(s, n, renderScope{draws: &draws})
}

// render evaluates a compiled node to a string. compile validates every node up
// front, so rendering a compiled tree cannot fail. sc holds the reference draws the
// render shares; each repeat iteration renders over draws of its own.
// A child this switch renders is one renderEdges must list too, or the fences miss it.
func render(s *generatorState, n node, sc renderScope) string {
	switch n := n.(type) {
	case *choice:
		return render(s, pick(s, n), sc)
	case *nullItem:
		return ""
	case *table:
		sc.row = renderedRow{n, sc.drawRowOf(s, n)}
		return expand(s, n.formatTemplate, sc)
	case *tableRow:
		return expand(s, n.t.formatTemplate, sc)
	case *tableColumn:
		row := sc.rowOf(n.t)
		if cell := n.t.cellTemplate(row, n.i); cell != nil {
			return render(s, cell, sc)
		}
		return n.t.cell(row, n.i)
	case *template:
		sc = sc.in(n)
		if n.repeat == 1 {
			if mark := sc.renderFrame(n); mark >= 0 {
				defer sc.draws.popFrames(mark)
			}
			if lit, fixed := n.fixedText(); fixed {
				return lit
			}
			return expand(s, n, sc)
		}
		return renderRepeat(s, n, sc)
	default:
		panic(internalError("uncompiled node %T", n))
	}
}

// renderRepeat renders each iteration of t as a render of its own, inside the name scopes
// rendering t, where a name bound outside t keeps its pick.
func renderRepeat(s *generatorState, t *template, sc renderScope) string {
	var b strings.Builder
	b.Grow(t.repeat * (t.compiled.grow + len(t.separator)))
	for i := 0; i < t.repeat; i++ {
		if i > 0 {
			b.WriteString(t.separator)
		}
		b.WriteString(expandAnew(s, t, sc.draws.frameStack, sc.base))
	}
	return b.String()
}

// expandAnew expands one repeat iteration of t as a render of its own, in no group, inside the
// name scopes of stack from base. Inlined into render's loop, its draws would move to the heap.
//
//go:noinline
func expandAnew(s *generatorState, t *template, stack *frameStack, base int) string {
	draws := renderDraws{frameStack: stack}
	if t.ownNameScope != nil {
		defer draws.popFrames(draws.pushFrame(newPickFrame(t.ownNameScope)))
	}
	return expand(s, t, renderScope{draws: &draws, base: base})
}

// pick selects one item. Uniform choices are O(1); weighted choices are an
// O(log n) search over precomputed cumulative weights.
func pick(s *generatorState, c *choice) node {
	if c.cum == nil {
		return c.items[s.IntN(len(c.items))]
	}
	return c.items[pickCum(s, c.cum)]
}

func expand(s *generatorState, t *template, sc renderScope) string {
	var b strings.Builder
	b.Grow(t.compiled.grow)
	// One draw per held name, for this expansion only: a nested template and each
	// repeat iteration get their own, since each is its own expansion. A reference
	// path reads its draw group's memo in sc instead.
	var hold *drawMemo
	if len(t.compiled.held) > 0 {
		hold = &drawMemo{
			variant: make(map[string]node, len(t.compiled.held)),
			value:   make(map[string]readValue, len(t.compiled.held)),
		}
	}
	for i := range t.compiled.ops {
		o := &t.compiled.ops[i]
		switch o.kind {
		case literalRun:
			b.WriteString(o.lit)
		case nameRead:
			b.WriteString(readField(s, t, hold, sc, o.arms[s.IntN(len(o.arms))]).text)
		case builtinCall:
			// Read before the call, so the value a calc computes is the value the
			// format showed. calcVars fixed the order op.operands holds.
			var operands []string
			if len(o.operands) > 0 {
				operands = make([]string, len(o.operands))
				for j, a := range o.operands {
					operands[j] = readField(s, t, hold, sc, a).text
				}
			}
			b.WriteString(o.call(s, b.String(), operands)) // b.String() is the output so far
		}
	}
	return b.String()
}

// readField renders one arm of a token. A name the expansion holds — a level some
// token addresses by a dotted path that is not a reference, or a field an operand
// reads — is drawn once and kept in hold, so {place.postal-code} and {place.locality}
// read one row, either read twice gives one value, and a shown operand is the operand
// computed. A field of a template rendering as part of a named pick is kept in the pick.
// Every other field is drawn afresh, so {word} {word} still draws twice.
func readField(s *generatorState, t *template, hold *drawMemo, sc renderScope, a arm) readValue {
	if a.kind == refPathRead {
		return readReference(s, t, sc, a)
	}
	if sc.draws.trace != nil {
		traceRead(sc.draws.trace, t, sc, a)
	}
	switch {
	case a.kind == namedRead:
		return readName(s, sc, a)
	case sc.pick != nil && !isRef(a.head) && sc.keeps(a):
		return readUnder(s, t, sc, a)
	}
	sc.pick = nil
	switch {
	case a.kind == freshRead && isRef(a.head):
		sc.base = sc.draws.depth()
		return readValue{text: render(s, t.head(a.head), sc)}
	case a.kind == freshRead:
		return readValue{text: render(s, t.head(a.head), sc)}
	}
	return readMemo(s, t, hold, nil, sc, a)
}

// readReference reads a reference path, kept in its draw group's memo, so its draw
// spans the render.
func readReference(s *generatorState, t *template, sc renderScope, a arm) readValue {
	if sc.draws.trace != nil {
		traceRead(sc.draws.trace, t, sc, a)
	}
	group := sc.groupDraws()
	return readMemo(s, t, &group.memo, &group.pins, sc, a)
}

func readMemo(s *generatorState, t *template, memo *drawMemo, pins *pinSet, sc renderScope, a arm) readValue {
	if r, done := memo.value[a.path]; done {
		return r
	}
	leaf := drawSteps(s, t.head(a.head), a.steps, pins, memo, a.levels)
	mark := -1
	if pins != nil {
		sc, mark = sc.at(leaf, pins).enter(leaf, memo)
	}
	r := renderLeaf(s, leaf, sc)
	if mark >= 0 {
		sc.draws.popFrames(mark)
	}
	if memo.value == nil {
		memo.value = map[string]readValue{}
	}
	memo.value[a.path] = r
	return r
}

func traceRead(trace renderTrace, t *template, sc renderScope, a arm) {
	var row renderedRow
	switch {
	case t.site.isCell():
		row = renderedRow{t.site.table, t.site.row}
	case t.site.isFormat():
		row = renderedRow{t.site.table, sc.rowOf(t.site.table)}
	}
	trace(strings.Clone(sc.group), row, a)
}

// renderLeaf draws and renders what a read lands on: null on a null item, or on a column of one
// reference alone whose read drew null.
func renderLeaf(s *generatorState, n node, sc renderScope) readValue {
	n = resolveChoice(s, n)
	switch leaf := n.(type) {
	case *nullItem:
		return readValue{null: true}
	case *template:
		if leaf.link.readsColumn != nil {
			return readReference(s, leaf, sc.in(leaf), leaf.link.readsColumn.a)
		}
	}
	return readValue{text: render(s, n, sc)}
}

// resolveChoice resolves a choice to one variant, so a held head is a concrete node the
// rest of the expansion shares. Nested choices unwrap too: a draw is one value, not
// another set to pick from.
func resolveChoice(s *generatorState, n node) node {
	for c, ok := n.(*choice); ok; c, ok = n.(*choice) {
		n = pick(s, c)
	}
	return n
}
