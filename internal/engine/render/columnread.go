package render

import (
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// readColumnItem reads the column t reads, where t is a column read. Rendering t's format
// instead would print a null the column drew as "". t's own name scope opens first, since the read
// may be of a name t binds, such as {/c as n}{n.x}.
func readColumnItem(s *drawstate.State, t *tree.Template, env renderEnv) (r readValue, reads bool) {
	if t.ReadsColumn() == nil {
		return readValue{}, false
	}
	if mark := env.renderFrame(t); mark >= 0 {
		defer env.frames.pop(mark)
	}
	return readField(s, t, env, t.ReadsColumn().Arm()), true
}
