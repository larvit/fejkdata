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
	f.set = holdSet{}
	sc := renderScope{set: &f.set}
	f.root.children = f.categories
	n, err := descend(f.rand, &f.root, segments, sc)
	if err != nil {
		return "", fmt.Errorf("fejkdata: %s: %w", path, err)
	}
	if _, ok := n.(*folder); ok {
		return "", fmt.Errorf("fejkdata: %s names a folder, not a value", path)
	}
	return render(f.rand, n, sc), nil
}

// descend walks named fields to the node a path names, pinning the table rows it
// selects or draws in sc. It is the one render-side step that can fail, because the
// path comes from the caller and may name a field or a row that does not exist. A
// choice consumes no segment, so the rest of the path must be one every variant
// carries before a variant is picked — a path that resolves at all resolves on
// every call.
func descend(s *session, root node, segments []string, sc renderScope) (node, error) {
	// Walked once without drawing first, so a path that fails moves no seeded stream.
	var probe pinSet
	if _, err := walkPath(root, segments, pathWalk{mode: walkProbe, pins: &probe}); err != nil {
		return nil, err
	}
	return drawPath(root, segments, "", &sc.hold().pins, &pathDraws{s: s}), nil
}

// render evaluates a compiled node to a string. compile validates every node up
// front, so rendering a compiled tree cannot fail. sc holds the reference draws the
// render shares; each repeat iteration renders over a hold set of its own.
func render(s *session, n node, sc renderScope) string {
	switch n := n.(type) {
	case *choice:
		return render(s, pick(s, n), sc)
	case *null:
		return ""
	case *table:
		sc.t, sc.row = n, n.draw(s)
		return expand(s, n.format, sc)
	case *row:
		sc.t, sc.row = n.t, sc.hold().pins.mustRow(n.t)
		return expand(s, n.t.format, sc)
	case *column:
		if sc.t != n.t {
			sc.t, sc.row = n.t, sc.hold().pins.mustRow(n.t)
		}
		if cell := n.t.cellNode(sc.row, n.i); cell != nil {
			return render(s, cell, sc)
		}
		return n.t.cell(sc.row, n.i)
	case *template:
		sc = sc.in(n)
		if n.repeat == 1 {
			if n.fixed {
				return n.lit
			}
			return expand(s, n, sc)
		}
		var b strings.Builder
		b.Grow(n.repeat * (n.grow + len(n.separator)))
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
	b.Grow(t.grow)
	// One draw per held name, for this expansion only: a nested template and each
	// repeat iteration get their own, since each is its own expansion. A reference
	// path reads the render's hold in sc instead.
	var held *hold
	if t.heldLocal {
		held = &hold{
			variant: make(map[string]node, len(t.held)),
			value:   make(map[string]draw, len(t.held)),
		}
	}
	for i := range t.ops {
		o := &t.ops[i]
		switch o.kind {
		case 'l':
			b.WriteString(o.lit)
		case 'f':
			b.WriteString(readField(s, t, held, sc, o.arms[s.IntN(len(o.arms))]).text)
		case 'b':
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
