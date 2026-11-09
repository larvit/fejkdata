package fejkdata

import (
	"errors"

	"github.com/larvit/fejkdata/internal/avoid"
	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/drawstate"
)

// sampleDraw draws a value of a builtin that draws each character from an alphabet, such as
// {digits(n)}, holding none of its where clause's strings. A string the clause reads from a name is
// read at each draw, and where no value avoids the strings then, the draw prints its fallback.
type sampleDraw struct {
	runs    avoid.Runs
	quoted  []string    // the strings quoted in the clause, as patterns a value must not hold
	plan    *avoid.Plan // the draws avoiding quoted; nil where quoted is empty
	call    builtinfunc.Call
	reads   []arm
	strings readStrings
	// fallback prints where the strings read leave no value; without an else, it is a value of
	// the builtin's own shape
	fallback fallback
}

// newSampleDraw builds the draw of o, a call with a where clause, and refuses a clause no value
// passes: one whose quoted strings every value holds, or whose length no value has.
func newSampleDraw(o *op) (*sampleDraw, error) {
	w := o.Where
	runs, _ := builtinfunc.Runs(o.Fn, o.Args)
	names := clauseNames(w)
	d := &sampleDraw{runs: runs, call: o.call, reads: o.operands,
		strings:  readStrings{clauseReads(names, w.NotNames), clauseReads(names, w.NotContainingNames)},
		fallback: fallbackOf(w, names, builtinfunc.Default(o.Fn, o.Args))}
	d.quoted = patterns(runs.Len(), w.Not, nonEmpty(w.NotContaining))
	if len(d.quoted) > 0 {
		if d.plan = avoid.Compile(runs, d.quoted); d.plan == nil {
			return nil, errors.New("no value passes the where clause")
		}
	}
	p := lengthProof{lengths: []int{runs.Len()}, reads: o.operands, count: 1, filters: true}
	p.lo, p.hi = lengthEnds(w, names)
	return d, p.proveEveryDrawPasses()
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

func (d *sampleDraw) draw(s *drawstate.State, env renderEnv, emitted string) string {
	not, containing := d.strings.read(s, env, d.reads)
	read := patterns(d.runs.Len(), not, containing)
	switch {
	case len(read) > 0:
		plan := avoid.Compile(d.runs, append(read, d.quoted...))
		if plan == nil {
			return d.fallback.at(s, env, d.reads)
		}
		return plan.Draw(s)
	case d.plan != nil:
		return d.plan.Draw(s)
	}
	return d.call(s, emitted, nil)
}
