package fejkdata

import (
	"errors"

	"github.com/larvit/fejkdata/internal/avoid"
	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/drawstate"
)

// alphabetDraw draws a value holding none of a where clause's strings, for a builtin that draws each
// character from an alphabet, such as {digits(n)}. A string the clause reads from a name is read at
// each draw, and where no value avoids the strings then, the draw prints its fallback.
type alphabetDraw struct {
	runs    avoid.Runs
	quoted  []string    // the strings quoted in the clause, as patterns a value must not hold
	plan    *avoid.Plan // the draws avoiding quoted; nil where quoted is empty
	call    builtinfunc.Call
	strings readStrings
	// fallback prints where the strings read leave no value; without an else, it is a value of
	// the builtin's own shape
	fallback fallback
	ownShape bool // no else names the fallback
}

// newAlphabetDraw builds the draw of o, a call with a where clause, and refuses a clause no value
// passes: one where every value holds a quoted string, or no value has its length.
func newAlphabetDraw(o *op, draws bindingDraws) (*alphabetDraw, error) {
	w := o.Where
	runs, _ := builtinfunc.Runs(o.Fn, o.Args)
	d := &alphabetDraw{runs: runs, call: o.call,
		strings:  readStrings{clauseReads(o.clause, w.NotNames), clauseReads(o.clause, w.NotContainingNames)},
		fallback: fallbackOf(w, o.clause, builtinfunc.Default(o.Fn, o.Args)), ownShape: w.Else == nil}
	d.quoted = patterns(runs.Len(), w.Not, nonEmpty(w.NotContaining))
	if len(d.quoted) > 0 {
		if d.plan = avoid.Compile(runs, d.quoted); d.plan == nil {
			return nil, errors.New("no value passes the where clause")
		}
	}
	p := lengthProof{lengths: []int{runs.Len()}, count: 1, filters: true}
	p.lo, p.hi = lengthEnds(w, o.clause)
	return d, p.proveEveryDrawPasses(draws)
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

func (d *alphabetDraw) draw(s *drawstate.State, _ *template, env renderEnv, emitted string) string {
	not, containing := d.strings.read(s, env)
	read := patterns(d.runs.Len(), not, containing)
	switch {
	case len(read) > 0:
		plan := avoid.Compile(d.runs, append(read, d.quoted...))
		if plan == nil {
			return d.fallback.at(s, env)
		}
		return plan.Draw(s)
	case d.plan != nil:
		return d.plan.Draw(s)
	}
	return d.call(s, emitted, nil)
}

// mayFallBack leaves out a fallback no else names: the builtin's own default is a value of its shape.
func (d *alphabetDraw) mayFallBack() bool { return d.strings.any() && !d.ownShape }
