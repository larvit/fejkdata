package render

import (
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
)

// A path's steps draw in one of three ways, and a caller picks one by the function it calls:
// drawSibling, drawFresh, or namedPick.draw for a read kept in a name's pick.

// drawSibling draws a path with no reference sigil from n, every choice afresh, and returns the
// leaf. Only a reference reaches a table, so this path never crosses one and holds no row.
func drawSibling(s *drawstate.State, n tree.Node, steps []tree.PathStep) tree.Node {
	for _, st := range steps {
		if st.Kind != tree.StepField {
			panic(invariant.Broken("step %d: a table's step of kind %d on a sibling read, which never crosses a table", st.At, st.Kind))
		}
		n = fieldOf(drawThroughChoices(s, n), st)
	}
	return n
}

// drawFresh draws a reference's path from n, every choice and row afresh, starting from the rows
// pins holds. It returns the leaf and the pins its row is in.
func drawFresh(s *drawstate.State, n tree.Node, steps []tree.PathStep, pins *tree.PinSet) (tree.Node, *tree.PinSet) {
	return drawAcrossTables(s, n, steps, pins, nil)
}

// keptDraws is where a path drawn for a name's pick keeps its draws: in the pick's memo, under
// the path's levels, so every read of the pick sees them. A nil keptDraws keeps nothing, and each
// draw is fresh.
type keptDraws struct {
	memo   *drawMemo
	levels []tree.PickKey
}

// variantOf is the variant of c a step at level at reads.
func (k *keptDraws) variantOf(s *drawstate.State, c *tree.Choice, at int) tree.Node {
	if k == nil {
		return drawThroughChoices(s, c)
	}
	return k.memo.variantOf(s, c, k.levels[at])
}

// stepDownPins is the pins a step at level at steps down into after a ".." stepped up to climbed:
// pins holding only the rows from climbed up, not the row the path came from, so the step draws
// the child's row afresh inside the row stepped up to.
func (k *keptDraws) stepDownPins(pins *tree.PinSet, climbed *tree.RowsTable, at int) *tree.PinSet {
	if k == nil {
		return pins.Above(climbed)
	}
	return k.memo.stepDownPins(pins, climbed, k.levels[at])
}

// drawAcrossTables draws the rows and variants a path's steps read from n, into pins, and returns
// the leaf and the pins its row is in. kept is non-nil on a read for a name's pick.
func drawAcrossTables(s *drawstate.State, n tree.Node, steps []tree.PathStep, pins *tree.PinSet, kept *keptDraws) (tree.Node, *tree.PinSet) {
	var climbed *tree.RowsTable // the table a ".." stepped up to, until the next step down
	for _, st := range steps {
		if c, ok := n.(*tree.Choice); ok {
			n = kept.variantOf(s, c, st.At)
		}
		if st.Kind == tree.StepField {
			n = fieldOf(n, st)
			continue
		}
		t, ok := n.(*tree.Table)
		if !ok {
			panic(invariant.Broken("step %d: a table's step of kind %d on %T", st.At, st.Kind, n))
		}
		switch {
		case st.Kind == tree.StepParent:
			climbed = t.Rows().Parent()
		case st.Kind == tree.StepChild && climbed != nil:
			pins = kept.stepDownPins(pins, climbed, st.At)
			climbed = nil
		}
		n = followStep(t, s, st, pins)
	}
	return n, pins
}

func fieldOf(n tree.Node, st tree.PathStep) tree.Node {
	child, ok := tree.Child(n, st.Name)
	if !ok {
		panic(invariant.Broken("step %d: no field %q; a path's steps should have been compiled from the node they are drawn from", st.At, st.Name))
	}
	return child
}
