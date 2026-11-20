package render

import (
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
)

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

// drawFresh draws a reference's path from n, every choice afresh and each row inside the rows
// pins holds. It returns the leaf and the pins its row is in.
func drawFresh(s *drawstate.State, n tree.Node, steps []tree.PathStep, pins *tree.PinSet) (tree.Node, *tree.PinSet) {
	return drawAcrossTables(s, n, steps, pins, nil)
}

// keptDraws is where a path drawn for a name's pick keeps its draws: in the pick's memo, under
// the path's levels, so every read of the pick sees them. A nil keptDraws doesn't keep
// anything, and each draw is fresh.
type keptDraws struct {
	memo   *drawMemo
	levels []tree.PickKey
}

// variantOf is the variant of c that a step reads, level being the index of the step's level.
func (k *keptDraws) variantOf(s *drawstate.State, c *tree.Choice, level int) tree.Node {
	if k == nil {
		return drawThroughChoices(s, c)
	}
	return k.memo.variantOf(s, c, k.levels[level])
}

// pinsBelow is the pins a step down to a child table draws in, level being the index of the
// step's level. Inside, they are pins. Climbed, they hold only the rows from climbedTo up, not
// the row the path came from, so the step draws the child's row afresh inside the row stepped up
// to; a pick's path keeps them under the level, so every path of the pick stepping down there
// reads one draw.
func (k *keptDraws) pinsBelow(pins *tree.PinSet, climbedTo *tree.RowsTable, level int) *tree.PinSet {
	switch {
	case climbedTo == nil:
		return pins
	case k == nil:
		return pins.Above(climbedTo)
	}
	return k.memo.stepDownPins(pins, climbedTo, k.levels[level])
}

// drawAcrossTables draws the rows and variants a path's steps read from n, into pins, and returns
// the leaf and the pins its row is in. kept is non-nil on a read for a name's pick
// (namedPick.draw), and nil on a fresh read (drawFresh).
//
// Among the rows it draws, the walk is in one of two states, which climbedTo holds:
//   - inside, climbedTo nil, where a path starts: every step draws inside the rows pins holds.
//   - climbed, after a ".." stepped up to a row of climbedTo: the next step down leaves the row
//     the path came from (pinsBelow), and puts the walk back inside.
//
// The state is a local beside pins, not a struct holding both: a struct would mix kept and pins
// for the compiler, and move every render's reword fakes to the heap.
func drawAcrossTables(s *drawstate.State, n tree.Node, steps []tree.PathStep, pins *tree.PinSet, kept *keptDraws) (tree.Node, *tree.PinSet) {
	var climbedTo *tree.RowsTable
	for _, st := range steps {
		if c, ok := n.(*tree.Choice); ok {
			n = kept.variantOf(s, c, st.At)
		}
		if st.Kind == tree.StepField {
			n = fieldOf(n, st)
			continue
		}
		t := tableOf(n, st)
		switch st.Kind {
		case tree.StepParent:
			climbedTo = t.Rows().Parent()
		case tree.StepChild:
			pins, climbedTo = kept.pinsBelow(pins, climbedTo, st.At), nil
		}
		n = followStep(t, s, st, pins)
	}
	return n, pins
}

func tableOf(n tree.Node, st tree.PathStep) *tree.Table {
	t, ok := n.(*tree.Table)
	if !ok {
		panic(invariant.Broken("step %d: a table's step of kind %d on %T", st.At, st.Kind, n))
	}
	return t
}

func fieldOf(n tree.Node, st tree.PathStep) tree.Node {
	child, ok := tree.Child(n, st.Name)
	if !ok {
		panic(invariant.Broken("step %d: no field %q; a path's steps should have been compiled from the node they are drawn from", st.At, st.Name))
	}
	return child
}
