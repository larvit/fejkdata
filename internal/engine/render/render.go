package render

import (
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// Walk draws steps from root, starting from the rows pins holds, and returns the node they
// reach with the rows it renders in.
func Walk(s *drawstate.State, root tree.Node, steps []tree.PathStep, pins *tree.PinSet) (tree.Node, *tree.PinSet) {
	return drawSteps(s, drawFresh, root, steps, pins, nil, nil)
}

// Pinned renders n once, in the rows pins holds.
func Pinned(s *drawstate.State, n tree.Node, pins *tree.PinSet) string {
	var frames frameStack
	env, _ := renderEnv{frames: &frames}.at(n, pins).enter(n, nil)
	return render(s, n, env)
}

// Once renders n once, outside any row.
func Once(s *drawstate.State, n tree.Node) string {
	var frames frameStack
	return render(s, n, renderEnv{frames: &frames})
}

// render steps into a child only through the Rendered methods of tree/renderedge.go, which
// tree.RenderEdges lists for the fences.
func render(s *drawstate.State, n tree.Node, env renderEnv) string {
	switch n := n.(type) {
	case *tree.Choice:
		return render(s, drawItem(s, n), env)
	case *tree.NullItem:
		return ""
	case *tree.Table:
		env.row = Row{Table: n, Index: env.drawRowOf(s, n)}
		return expand(s, n.RenderedFormat(), env)
	case *tree.TableRow:
		return expand(s, n.RenderedFormat(), env)
	case *tree.TableColumn:
		row := env.rowOf(n.Table())
		if cell := n.RenderedCell(row); cell != nil {
			return render(s, cell, env)
		}
		return n.Table().Rows().Cell(row, n.Index())
	case *tree.Template:
		if n.Repeat() == 1 {
			if mark := env.renderFrame(n); mark >= 0 {
				defer env.frames.pop(mark)
			}
			if lit, fixed := n.FixedText(); fixed {
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
func renderRepeat(s *drawstate.State, t *tree.Template, env renderEnv) string {
	var b strings.Builder
	b.Grow(t.Repeat() * (t.Compiled().Grow() + len(t.Separator())))
	own := t.OwnScope()
	for i := 0; i < t.Repeat(); i++ {
		if i > 0 {
			b.WriteString(t.Separator())
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

func drawItem(s *drawstate.State, c *tree.Choice) tree.Node {
	if c.Cum() == nil {
		return c.RenderedItem(s.IntN(len(c.Items())))
	}
	return c.RenderedItem(s.Weighted(c.Cum()))
}

func expand(s *drawstate.State, t *tree.Template, env renderEnv) string {
	var b strings.Builder
	b.Grow(t.Compiled().Grow())
	ops := t.Compiled().Ops()
	for i := range ops {
		o := &ops[i]
		switch tok := o.Token(); {
		case tok.Kind == grammar.LiteralRun:
			b.WriteString(tok.Lit)
		case tok.Kind == grammar.NameBind: // its draw prints where the name is read
		case o.Draw() != nil:
			b.WriteString(drawToken(s, t, env, o.Draw(), b.String()))
		case tok.Kind == grammar.PathRead:
			b.WriteString(readField(s, t, env, o.Alternatives()[s.IntN(len(o.Alternatives()))]).text)
		case tok.Kind == grammar.BuiltinCall:
			var operands []string
			if len(o.Operands()) > 0 {
				operands = make([]string, len(o.Operands()))
				for j, a := range o.Operands() {
					operands[j] = readField(s, t, env, a).text
				}
			}
			b.WriteString(o.Call()(s, b.String(), operands)) // b.String() is the output so far
		}
	}
	return b.String()
}

// readField renders one read of a token. A read is kept in a name's pick when it reads the name, or
// a level that a read of the name being rendered addresses; every other read draws afresh, so
// {word} {word} draws twice and {p.a} {p.b} reads two draws of p.
// A fresh read clears env.pick, which stops a read nested in it from keying the pick's memo under
// env.pickAt and sharing a value with another node at that key; a reference's read clears it in enter.
func readField(s *drawstate.State, t *tree.Template, env renderEnv, a tree.Read) readValue {
	switch {
	case a.Kind() == tree.NamedRead:
		return readName(s, env, a)
	case keptInPick(env, a):
		return readUnder(s, t, env, a)
	}
	if !grammar.IsRef(a.Head()) {
		env.pick = nil
		leaf, _ := drawSteps(s, drawSibling, t.StartOf(a.Head(), t.Refs()), a.Steps(), nil, nil, nil)
		return renderLeaf(s, leaf, env)
	}
	var pins tree.PinSet
	leaf, leafPins := drawSteps(s, drawFresh, t.StartOf(a.Head(), t.Refs()), a.Steps(), &pins, nil, nil)
	env, mark := env.at(leaf, leafPins).enter(leaf, nil)
	r := renderLeaf(s, leaf, env)
	env.frames.pop(mark)
	return r
}

// renderLeaf draws and renders what a read lands on: null on a null item, or on a column that only
// reads one reference or name whose read drew null.
func renderLeaf(s *drawstate.State, n tree.Node, env renderEnv) readValue {
	n = drawThroughChoices(s, n)
	switch leaf := n.(type) {
	case *tree.NullItem:
		return readValue{null: true}
	case *tree.Template:
		if r, reads := readColumnItem(s, leaf, env); reads {
			return r
		}
	}
	return readValue{text: render(s, n, env)}
}

// drawThroughChoices draws a choice's variant. Nested choices unwrap too: a draw is one value,
// not another set to draw from.
func drawThroughChoices(s *drawstate.State, n tree.Node) tree.Node {
	for c, ok := n.(*tree.Choice); ok; c, ok = n.(*tree.Choice) {
		n = drawItem(s, c)
	}
	return n
}
