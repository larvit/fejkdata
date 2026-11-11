package fejkdata

import (
	"errors"

	"github.com/larvit/fejkdata/internal/avoid"
	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// newAlphabetDraw builds the draw of o, a call with a where clause, and refuses a clause where
// every value holds a quoted string.
func newAlphabetDraw(o *tree.Op) (*tree.AlphabetDraw, error) {
	tok := o.Token()
	w := tok.Where
	runs, _ := builtinfunc.Runs(tok.Fn, tok.Args)
	quoted := patterns(runs.Len(), w.Not, nonEmpty(w.NotContaining))
	var plan *avoid.Plan
	if len(quoted) > 0 {
		if plan = avoid.Compile(runs, quoted); plan == nil {
			return nil, errors.New("no value passes the where clause")
		}
	}
	strings := tree.NewReadStrings(clauseReads(o.Clause(), w.NotNames), clauseReads(o.Clause(), w.NotContainingNames))
	fallback := fallbackOf(w, o.Clause(), builtinfunc.Default(tok.Fn, tok.Args))
	return tree.NewAlphabetDraw(runs, quoted, plan, o.Call(), strings, fallback, w.Else == nil), nil
}

// patterns is what a value n bytes long must not hold: each of not that is n bytes long, which a
// value holds only by equalling it, and each of containing.
func patterns(n int, not, containing []string) []string {
	var out []string
	for _, s := range not {
		if len(s) == n {
			out = append(out, s)
		}
	}
	return append(out, containing...)
}

func drawAlphabet(d *tree.AlphabetDraw, s *drawstate.State, env renderEnv, emitted string) string {
	not, containing := readClauseStrings(d.Strings(), s, env)
	read := patterns(d.Runs().Len(), not, containing)
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
