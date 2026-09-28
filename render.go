package fejkdata

import (
	"fmt"
	"strings"
)

// rng is the randomness a builtin sample draws from; *rand.Rand satisfies it. The
// render path takes the concrete *session instead, for the reason below.
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
	f.root.children = f.categories
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
	if _, err := (&pathProbe{drawn: map[*table]bool{}}).walk(root, segments); err != nil {
		return nil, err
	}
	return drawPath(root, segments, "", &pathDraw{s: s, pins: &sc.groupHold().pins}), nil
}

// render evaluates a compiled node to a string. compile validates every node up
// front, so rendering a compiled tree cannot fail. sc holds the reference draws the
// render shares; each repeat iteration renders over a hold set of its own.
// renderEdges mirrors this switch for the fences, so a new node kind goes in both.
func render(s *session, n node, sc renderScope) string {
	switch n := n.(type) {
	case *choice:
		return render(s, pick(s, n), sc)
	case *nullItem:
		return ""
	case *table:
		sc.row = tablePin{n, n.drawRow(s)}
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
		panic(fmt.Sprintf("fejkdata: uncompiled node %T", n))
	}
}

// pick selects one item. Uniform choices are O(1); weighted choices are an
// O(log n) search over precomputed cumulative weights. compile guarantees a
// non-empty choice and a finite positive total, so the index is always in range.
// The session is concrete rather than the rng interface, which would make the
// walk that draws through it leak its hold set to the heap.
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
		case fieldAlternation:
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
