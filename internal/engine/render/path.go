package render

import (
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
)

// drawMode is the kind of read drawSteps makes.
type drawMode uint8

const (
	// drawSibling reads a path with no reference sigil. Only a reference reaches a table, so this
	// path never crosses one, and it takes no pins and no memo.
	drawSibling drawMode = iota
	// drawFresh reads a reference with pins of its own, drawing every choice and row afresh.
	drawFresh
	// drawInPick reads for a name's pick, into the pick's pins. memo keeps the variant drawn at each
	// of levels and the rows drawn by each step down after a "..", so every read of the pick sees them.
	drawInPick
)

// drawSteps draws the rows and variants a path's steps read from n, and returns the leaf and the
// pins its row is in. mode says which of pins, memo and levels it reads: drawSibling none,
// drawFresh pins, drawInPick all three.
func drawSteps(s *drawstate.State, mode drawMode, n tree.Node, steps []tree.PathStep, pins *tree.PinSet, memo *drawMemo, levels []tree.PickKey) (tree.Node, *tree.PinSet) {
	// climbed is the table a ".." stepped up to. The next step down from it draws the child's row
	// afresh inside the row stepped up to, so it takes pins holding only the rows from climbed up,
	// not the row the path came from.
	var climbed *tree.RowsTable
	for _, st := range steps {
		if c, ok := n.(*tree.Choice); ok {
			if mode == drawInPick {
				n = memo.variantOf(s, c, levels[st.At])
			} else {
				n = drawThroughChoices(s, c)
			}
		}
		if st.Kind == tree.StepField {
			var ok bool
			if n, ok = tree.Child(n, st.Name); !ok {
				panic(invariant.Broken("step %d: no field %q; a path's steps should have been compiled from the node they are drawn from", st.At, st.Name))
			}
			continue
		}
		if mode == drawSibling {
			panic(invariant.Broken("step %d: a table's step of kind %d on a sibling read, which never crosses a table", st.At, st.Kind))
		}
		t, ok := n.(*tree.Table)
		if !ok {
			panic(invariant.Broken("step %d: a table's step of kind %d on %T", st.At, st.Kind, n))
		}
		switch {
		case st.Kind == tree.StepParent:
			climbed = t.Rows().Parent()
		case st.Kind == tree.StepChild && climbed != nil:
			if mode == drawInPick {
				pins = memo.stepDownPins(pins, climbed, levels, st.At)
			} else {
				pins = pins.Above(climbed)
			}
			climbed = nil
		}
		n = followStep(t, s, st, pins)
	}
	return n, pins
}
