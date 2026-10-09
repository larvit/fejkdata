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
// drawn at random. A path naming a folder (no value of its own) is an error.
func (f *Generator) Fake(path string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	path, segments, err := f.loadCallerPath(path)
	if err != nil {
		return "", err
	}
	n, pins, err := drawCallerPath(f.drawState, &f.root, segments)
	if err != nil {
		return "", f.noDataNote(fmt.Errorf("fejkdata: %s: %w", path, err))
	}
	if _, ok := n.(*folder); ok {
		return "", fmt.Errorf("fejkdata: %s names a folder, not a value", path)
	}
	var frames frameStack
	env, _ := renderEnv{frames: &frames}.at(n, &pins).enter(n, nil)
	return render(f.drawState, n, env), nil
}

// drawCallerPath walks a caller's path to the node it names, returning the rows its leaf renders in.
func drawCallerPath(s *drawstate.State, root node, segments []string) (node, pinSet, error) {
	// docs/decisions.md#a-path-is-walked-once-without-drawing-before-it-is-walked-for-real
	var buf [16]pathStep
	steps, err := callerPathSteps(root, segments, buf[:0])
	if err != nil {
		return nil, pinSet{}, err
	}
	var pins pinSet
	n, leafPins := drawSteps(s, root, steps, &pins, nil, nil)
	return n, *leafPins, nil
}

func renderOnce(s *drawstate.State, n node) string {
	var frames frameStack
	return render(s, n, renderEnv{frames: &frames})
}

// render evaluates a compiled node to a string. compile validates every node up
// front, so rendering a compiled tree cannot fail.
// A child this switch renders is one renderEdges must list too, or the fences miss it.
func render(s *drawstate.State, n node, env renderEnv) string {
	switch n := n.(type) {
	case *choice:
		return render(s, drawItem(s, n), env)
	case *nullItem:
		return ""
	case *table:
		env.row = renderedRow{n, env.drawRowOf(s, n)}
		return expand(s, n.formatTemplate, env)
	case *tableRow:
		return expand(s, n.t.formatTemplate, env)
	case *tableColumn:
		row := env.rowOf(n.t)
		if cell := n.t.cellTemplate(row, n.i); cell != nil {
			return render(s, cell, env)
		}
		return n.t.rows.Cell(row, n.i)
	case *template:
		if n.repeat == 1 {
			if mark := env.renderFrame(n); mark >= 0 {
				defer env.frames.pop(mark)
			}
			if lit, fixed := n.fixedText(); fixed {
				return lit
			}
			return expand(s, n, env)
		}
		return renderRepeat(s, n, env)
	default:
		panic(invariant.Broken("uncompiled node %T", n))
	}
}

// renderRepeat renders each iteration of t inside the name scopes rendering t, where a name bound
// outside t keeps its pick, and a name t binds takes a fresh pick.
func renderRepeat(s *drawstate.State, t *template, env renderEnv) string {
	var b strings.Builder
	b.Grow(t.repeat * (t.compiled.grow + len(t.separator)))
	own := t.ownScope()
	for i := 0; i < t.repeat; i++ {
		if i > 0 {
			b.WriteString(t.separator)
		}
		if own == nil {
			b.WriteString(expand(s, t, env))
			continue
		}
		mark := env.frames.push(newPickFrame(own))
		b.WriteString(expand(s, t, env))
		env.frames.pop(mark)
	}
	return b.String()
}

func drawItem(s *drawstate.State, c *choice) node {
	if c.cum == nil {
		return c.items[s.IntN(len(c.items))]
	}
	return c.items[s.Weighted(c.cum)]
}

func expand(s *drawstate.State, t *template, env renderEnv) string {
	var b strings.Builder
	b.Grow(t.compiled.grow)
	for i := range t.compiled.ops {
		o := &t.compiled.ops[i]
		switch o.Kind {
		case grammar.LiteralRun:
			b.WriteString(o.Lit)
		case grammar.PathRead:
			b.WriteString(readField(s, t, env, o.arms[s.IntN(len(o.arms))]).text)
		case grammar.BuiltinCall:
			var operands []string
			if len(o.operands) > 0 {
				operands = make([]string, len(o.operands))
				for j, a := range o.operands {
					operands[j] = readField(s, t, env, a).text
				}
			}
			b.WriteString(o.call(s, b.String(), operands)) // b.String() is the output so far
		}
	}
	return b.String()
}

// readField renders one arm of a token. A read is kept in a name's pick when it reads the name, or
// a level that a read of the name being rendered addresses; every other read draws afresh, so
// {word} {word} draws twice and {p.a} {p.b} reads two draws of p.
// A fresh read clears env.pick: were it kept, a read nested in the fresh one would key the pick's
// memo under env.pickAt and share a value with another node at that key.
func readField(s *drawstate.State, t *template, env renderEnv, a arm) readValue {
	switch {
	case a.kind == namedRead:
		return readName(s, env, a)
	case keptInPick(env, a):
		return readUnder(s, t, env, a)
	}
	env.pick = nil
	if !grammar.IsRef(a.head) {
		leaf, _ := drawSteps(s, t.startOf(a.head), a.steps, nil, nil, a.levels)
		return renderLeaf(s, leaf, env)
	}
	var pins pinSet
	leaf, leafPins := drawSteps(s, t.startOf(a.head), a.steps, &pins, nil, a.levels)
	env, mark := env.at(leaf, leafPins).enter(leaf, nil)
	r := renderLeaf(s, leaf, env)
	env.frames.pop(mark)
	return r
}

// renderLeaf draws and renders what a read lands on: null on a null item, or on a column that only
// reads one reference or name whose read drew null.
func renderLeaf(s *drawstate.State, n node, env renderEnv) readValue {
	n = drawThroughChoices(s, n)
	switch leaf := n.(type) {
	case *nullItem:
		return readValue{null: true}
	case *template:
		if leaf.readsColumn != nil {
			return readField(s, leaf, env, leaf.readsColumn.a)
		}
	}
	return readValue{text: render(s, n, env)}
}

// drawThroughChoices draws a choice's variant. Nested choices unwrap too: a draw is one value,
// not another set to draw from.
func drawThroughChoices(s *drawstate.State, n node) node {
	for c, ok := n.(*choice); ok; c, ok = n.(*choice) {
		n = drawItem(s, c)
	}
	return n
}
