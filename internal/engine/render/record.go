package render

import (
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// Record draws each column of t once, in name order, as one render, so the columns read one
// pick of each name. It calls emit with each column's index, its text and whether it drew null;
// a table's columns read row.
func Record(s *drawstate.State, t *tree.Template, row Row, emit func(i int, text string, null bool)) {
	var frames frameStack
	env := renderEnv{frames: &frames, row: row}
	if mark := env.renderFrame(t); mark >= 0 {
		defer env.frames.pop(mark)
	}
	for i, c := range t.Columns() {
		var v readValue
		if c.BoundWhole() != nil {
			v = readName(s, env, tree.NewRead("", "", nil, wholeLevels, nil, nil).ThroughName(c.BoundWhole(), 0))
		} else {
			v = renderLeaf(s, c.Field(), env)
		}
		emit(i, v.text, v.null)
	}
}

// wholeLevels is the levels of a read of a whole name, as load compiles {n}.
var wholeLevels = tree.NamedLevels(nil)
