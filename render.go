package fejkdata

import (
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// Fake generates a value for a dot path. Each segment descends one level: folder
// names and the category (JSON file) come first, then named fields within it,
// e.g. "sv_SE.address" or "sv_SE.address.street". Choices along the way are
// resolved at random. A path naming a folder (no value of its own) is an error.
func (f *Generator) Fake(path string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	segments, err := grammar.SplitPath(path)
	if err != nil {
		return "", fmt.Errorf("fejkdata: %w", err)
	}
	f.loadShippedAt(segments)
	n, pins, err := descend(f.drawState, &f.root, segments)
	if err != nil {
		return "", fmt.Errorf("fejkdata: %s: %w", path, err)
	}
	if _, ok := n.(*folder); ok {
		return "", fmt.Errorf("fejkdata: %s names a folder, not a value", path)
	}
	var frames frameStack
	sc, _ := renderScope{frames: &frames}.at(n, &pins).enter(n, nil)
	return render(f.drawState, n, sc), nil
}

// descend walks a caller's path to the node it names, returning the rows its leaf renders in.
func descend(s *drawstate.State, root node, segments []string) (node, pinSet, error) {
	// docs/decisions.md#a-path-is-walked-once-without-drawing-before-it-is-walked-for-real
	var buf [16]pathStep
	steps, err := probePath(root, segments, buf[:0])
	if err != nil {
		return nil, pinSet{}, err
	}
	var pins pinSet
	n, leafPins := drawSteps(s, root, steps, &pins, nil, nil)
	return n, *leafPins, nil
}

// renderOnce renders n as one render, over frames of its own.
func renderOnce(s *drawstate.State, n node) string {
	var frames frameStack
	return render(s, n, renderScope{frames: &frames})
}

// render evaluates a compiled node to a string. compile validates every node up
// front, so rendering a compiled tree cannot fail.
// A child this switch renders is one renderEdges must list too, or the fences miss it.
func render(s *drawstate.State, n node, sc renderScope) string {
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
		if n.repeat == 1 {
			if mark := sc.renderFrame(n); mark >= 0 {
				defer sc.frames.pop(mark)
			}
			if lit, fixed := n.fixedText(); fixed {
				return lit
			}
			return expand(s, n, sc)
		}
		return renderRepeat(s, n, sc)
	default:
		panic(invariant.Broken("uncompiled node %T", n))
	}
}

// renderRepeat renders each iteration of t inside the name scopes rendering t, where a name bound
// outside t keeps its pick, and a name t binds picks again.
func renderRepeat(s *drawstate.State, t *template, sc renderScope) string {
	var b strings.Builder
	b.Grow(t.repeat * (t.compiled.grow + len(t.separator)))
	for i := 0; i < t.repeat; i++ {
		if i > 0 {
			b.WriteString(t.separator)
		}
		if t.ownNameScope == nil {
			b.WriteString(expand(s, t, sc))
			continue
		}
		mark := sc.frames.push(newPickFrame(t.ownNameScope))
		b.WriteString(expand(s, t, sc))
		sc.frames.pop(mark)
	}
	return b.String()
}

// pick selects one item. Uniform choices are O(1); weighted choices are an
// O(log n) search over precomputed cumulative weights.
func pick(s *drawstate.State, c *choice) node {
	if c.cum == nil {
		return c.items[s.IntN(len(c.items))]
	}
	return c.items[pickCum(s, c.cum)]
}

func expand(s *drawstate.State, t *template, sc renderScope) string {
	var b strings.Builder
	b.Grow(t.compiled.grow)
	for i := range t.compiled.ops {
		o := &t.compiled.ops[i]
		switch o.Kind {
		case grammar.LiteralRun:
			b.WriteString(o.Lit)
		case grammar.NameRead:
			b.WriteString(readField(s, t, sc, o.arms[s.IntN(len(o.arms))]).text)
		case grammar.BuiltinCall:
			var operands []string
			if len(o.operands) > 0 {
				operands = make([]string, len(o.operands))
				for j, a := range o.operands {
					operands[j] = readField(s, t, sc, a).text
				}
			}
			b.WriteString(o.call(s, b.String(), operands)) // b.String() is the output so far
		}
	}
	return b.String()
}

// readField renders one arm of a token. A read of a name, or of a level a read of the name
// rendering addresses, is kept in that name's pick; every other read draws afresh, so {word}
// {word} draws twice and {p.a} {p.b} reads two draws of p.
func readField(s *drawstate.State, t *template, sc renderScope, a arm) readValue {
	switch {
	case a.kind == namedRead:
		return readName(s, sc, a)
	case sc.pick != nil && !grammar.IsRef(a.head) && sc.keeps(a):
		return readUnder(s, t, sc, a)
	}
	sc.pick = nil
	if !grammar.IsRef(a.head) {
		leaf, _ := drawSteps(s, t.head(a.head), a.steps, nil, nil, a.levels)
		return renderLeaf(s, leaf, sc)
	}
	var pins pinSet
	leaf, leafPins := drawSteps(s, t.head(a.head), a.steps, &pins, nil, a.levels)
	sc, mark := sc.at(leaf, leafPins).enter(leaf, nil)
	r := renderLeaf(s, leaf, sc)
	sc.frames.pop(mark)
	return r
}

// renderLeaf draws and renders what a read lands on: null on a null item, or on a column that only
// reads one reference or name whose read drew null.
func renderLeaf(s *drawstate.State, n node, sc renderScope) readValue {
	n = resolveChoice(s, n)
	switch leaf := n.(type) {
	case *nullItem:
		return readValue{null: true}
	case *template:
		if leaf.link.readsColumn != nil {
			return readField(s, leaf, sc, leaf.link.readsColumn.a)
		}
	}
	return readValue{text: render(s, n, sc)}
}

// resolveChoice resolves a choice to one variant. Nested choices unwrap too: a draw is one value,
// not another set to pick from.
func resolveChoice(s *drawstate.State, n node) node {
	for c, ok := n.(*choice); ok; c, ok = n.(*choice) {
		n = pick(s, c)
	}
	return n
}
