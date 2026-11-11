package render

import (
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
)

func followStep(t *tree.Table, s *drawstate.State, st tree.PathStep, pins *tree.PinSet) tree.Node {
	switch st.Kind {
	case tree.StepSelect:
		if err := pins.PinRow(t.Rows(), st.Row); err != nil {
			panic(invariant.Broken("%s[%s]: %v", t.Rows().Segment(), st.Name, err))
		}
		return t
	case tree.StepDraw:
		t.Rows().DrawIn(s, pins)
		return t
	case tree.StepRow:
		return t.RowNode()
	case tree.StepColumn:
		return t.FormatTemplate().Fields()[st.Name]
	case tree.StepChild:
		return t.Descendant(st.Name)
	case tree.StepParent:
		return t.Rows().Parent().Owner()
	}
	panic(invariant.Broken("followStep has no case for step kind %d", st.Kind))
}
