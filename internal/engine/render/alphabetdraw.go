package render

import (
	"github.com/larvit/fejkdata/internal/avoid"
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

func drawAlphabet(d *tree.AlphabetDraw, s *drawstate.State, env renderEnv, emitted string) string {
	not, containing := readClauseStrings(d.Strings(), s, env)
	read := avoid.Patterns(d.Runs().Len(), not, containing)
	switch {
	case len(read) > 0:
		plan := avoid.Compile(d.Runs(), append(read, d.Quoted()...))
		if plan == nil {
			return fallbackAt(d.Fallback(), s, env)
		}
		return plan.Draw(s)
	case d.Plan() != nil:
		return d.Plan().Draw(s)
	}
	return d.Call()(s, emitted, nil)
}
