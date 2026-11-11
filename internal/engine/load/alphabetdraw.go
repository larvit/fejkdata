package load

import (
	"errors"

	"github.com/larvit/fejkdata/internal/avoid"
	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// newAlphabetDraw builds the draw of o, a call with a where clause, and refuses a clause where
// every value holds a quoted string.
func newAlphabetDraw(o *tree.Op) (*tree.AlphabetDraw, error) {
	tok := o.Token()
	w := tok.Where
	runs, _ := builtinfunc.Runs(tok.Fn, tok.Args)
	quoted := avoid.Patterns(runs.Len(), w.Not, nonEmpty(w.NotContaining))
	var plan *avoid.Plan
	if len(quoted) > 0 {
		if plan = avoid.Compile(runs, quoted); plan == nil {
			return nil, errors.New("no value passes the where clause")
		}
	}
	strings := tree.NewReadStrings(clauseReads(o.Clause(), w.NotNames), clauseReads(o.Clause(), w.NotContainingNames))
	fallback := fallbackOf(w, o.Clause(), builtinfunc.Default(tok.Fn, tok.Args))
	return tree.NewAlphabetDraw(tree.AlphabetDrawSpec{
		Runs:     runs,
		Quoted:   quoted,
		Plan:     plan,
		Call:     o.Call(),
		Strings:  strings,
		Fallback: fallback,
		OwnShape: w.Else == nil,
	}), nil
}
