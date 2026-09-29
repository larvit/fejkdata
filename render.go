package fejkdata

import (
	"fmt"
	"strings"
)

// rng is the randomness a builtin sample draws from; *rand.Rand satisfies it. The
// render path takes the concrete *session instead, which keeps the hold set of the walk
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
	var set holdSet
	sc := renderScope{set: &set}
	n, err := descend(f.rand, &f.root, segments, sc)
	if err != nil {
		return "", fmt.Errorf("fejkdata: %s: %w", path, err)
	}
	if _, ok := n.(*folder); ok {
		return "", fmt.Errorf("fejkdata: %s names a folder, not a value", path)
	}
	return render(f.rand, n, sc.at(n, &sc.groupHold().pins)), nil
}

// descend walks a caller's path to the node it names, pinning in sc the rows it
// selects or draws.
func descend(s *session, root node, segments []string, sc renderScope) (node, error) {
	// docs/decisions.md#a-path-is-walked-once-without-drawing-before-it-is-walked-for-real
	if _, err := (&pathProbe{}).walk(root, segments); err != nil {
		return nil, err
	}
	return drawPath(root, segments, "", &pathDraw{s: s, pins: &sc.groupHold().pins}), nil
}

// renderOnce renders n as one render, over a hold set of its own.
func renderOnce(s *session, n node) string {
	var set holdSet
	return render(s, n, renderScope{set: &set})
}

// render evaluates a compiled node to a string. compile validates every node up
// front, so rendering a compiled tree cannot fail. sc holds the reference draws the
// render shares; each repeat iteration renders over a hold set of its own.
// A child this switch renders is one renderEdges must list too, or the fences miss it.
func render(s *session, n node, sc renderScope) string {
	switch n := n.(type) {
	case *choice:
		return render(s, pick(s, n), sc)
	case *nullItem:
		return ""
	case *table:
		sc.row = renderedRow{n, n.drawRow(s)}
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
			if lit, fixed := n.fixedText(); fixed {
				return lit
			}
			return expand(s, n, sc)
		}
		var b strings.Builder
		b.Grow(n.repeat * (n.compiled.grow + len(n.separator)))
		for i := 0; i < n.repeat; i++ {
			if i > 0 {
				b.WriteString(n.separator)
			}
			b.WriteString(expandAnew(s, n))
		}
		return b.String()
	default:
		panic(internalError("uncompiled node %T", n))
	}
}

// expandAnew expands one repeat iteration of t as a render of its own, in no group. Inlined into
// render's loop, its hold set would move to the heap.
//
//go:noinline
func expandAnew(s *session, t *template) string {
	var set holdSet
	return expand(s, t, renderScope{set: &set})
}

// pick selects one item. Uniform choices are O(1); weighted choices are an
// O(log n) search over precomputed cumulative weights. compile guarantees a
// non-empty choice and a finite positive total, so the index is always in range.
func pick(s *session, c *choice) node {
	if c.cum == nil {
		return c.items[s.IntN(len(c.items))]
	}
	return c.items[pickCum(s, c.cum)]
}

// expand renders a template's compiled ops. compile validated every token, so this
// cannot fail.
func expand(s *session, t *template, sc renderScope) string {
	var b strings.Builder
	b.Grow(t.compiled.grow)
	// One draw per held name, for this expansion only: a nested template and each
	// repeat iteration get their own, since each is its own expansion. A reference
	// path reads the render's hold in sc instead.
	var held *hold
	if t.compiled.heldLocal {
		held = &hold{
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
			b.WriteString(readField(s, t, held, sc, o.arms[s.IntN(len(o.arms))]).text)
		case builtinCall:
			// Read before the call, so the value a calc computes is the value the
			// format showed. calcVars fixed the order op.operands holds.
			var operands []string
			if len(o.operands) > 0 {
				operands = make([]string, len(o.operands))
				for j, a := range o.operands {
					operands[j] = readField(s, t, held, sc, a).text
				}
			}
			b.WriteString(o.call(s, b.String(), operands)) // b.String() is the output so far
		}
	}
	return b.String()
}

// readField renders one arm of a token. A name the expansion holds — a level some
// token addresses by a dotted path that is not a reference, or a field an operand
// reads — is drawn once and kept in held, so {place.postal-code} and {place.locality}
// read one row, either read twice gives one value, and a shown operand is the operand
// computed. Every other name is drawn afresh, so {word} {word} still draws twice.
func readField(s *session, t *template, held *hold, sc renderScope, a arm) readValue {
	if isRef(a.head) && len(a.tail) > 0 {
		return readReference(s, t, sc, a)
	}
	if sc.set.trace != nil {
		traceRead(sc.set.trace, t, sc, a)
	}
	if _, expansionHolds := t.compiled.held[a.head]; !expansionHolds {
		if len(a.tail) > 0 {
			panic(internalError("%q reads a path into %q, which the expansion does not hold", a.spelling, a.head))
		}
		return readValue{text: render(s, t.head(a.head), sc)}
	}
	if held == nil {
		panic(internalError("%q reads a name the expansion holds, with no hold to keep it in", a.spelling))
	}
	return readHeld(s, t, held, nil, sc, a)
}

// readReference reads a reference path, kept in the render's hold for its group, so its draw
// spans the render.
func readReference(s *session, t *template, sc renderScope, a arm) readValue {
	if sc.set.trace != nil {
		traceRead(sc.set.trace, t, sc, a)
	}
	group := sc.groupHold()
	return readHeld(s, t, &group.hold, &group.pins, sc, a)
}

func readHeld(s *session, t *template, held *hold, pins *pinSet, sc renderScope, a arm) readValue {
	if r, done := held.value[a.path]; done {
		return r
	}
	leaf := drawPath(t.head(a.head), a.tail, a.head, &pathDraw{s: s, held: held, pins: pins, a: &a})
	if pins != nil {
		sc = sc.at(leaf, pins)
	}
	r := renderLeaf(s, leaf, sc)
	if held.value == nil {
		held.value = map[string]readValue{}
	}
	held.value[a.path] = r
	return r
}

func traceRead(trace renderTrace, t *template, sc renderScope, a arm) {
	var row renderedRow
	switch {
	case t.site.isCell():
		row = renderedRow{t.site.table, t.site.row}
	case t.site.table != nil:
		row = renderedRow{t.site.table, sc.rowOf(t.site.table)}
	}
	trace(strings.Clone(sc.group), row, a)
}

// renderLeaf draws and renders what a read lands on: null on a null item, or on a column of one
// reference alone whose read drew null.
func renderLeaf(s *session, n node, sc renderScope) readValue {
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
func resolveChoice(s *session, n node) node {
	for c, ok := n.(*choice); ok; c, ok = n.(*choice) {
		n = pick(s, c)
	}
	return n
}
