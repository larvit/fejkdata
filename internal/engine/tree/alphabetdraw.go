package tree

import (
	"github.com/larvit/fejkdata/internal/avoid"
	"github.com/larvit/fejkdata/internal/builtinfunc"
)

// AlphabetDraw draws a value holding none of a where clause's strings, for a builtin that draws each
// character from an alphabet, such as {digits(n)}. A string the clause reads from a name is read at
// each draw, and where no value avoids the strings then, the draw prints its fallback.
type AlphabetDraw struct {
	runs    avoid.Runs
	quoted  []string    // the strings quoted in the clause, as patterns a value must not hold
	plan    *avoid.Plan // the draws avoiding quoted; nil where quoted is empty
	call    builtinfunc.Call
	strings ReadStrings
	// fallback prints where the strings read leave no value; without an else, it is a value of
	// the builtin's own shape
	fallback Fallback
	ownShape bool // no else names the fallback
}

func NewAlphabetDraw(runs avoid.Runs, quoted []string, plan *avoid.Plan, call builtinfunc.Call, strings ReadStrings, fallback Fallback, ownShape bool) *AlphabetDraw {
	return &AlphabetDraw{runs, quoted, plan, call, strings, fallback, ownShape}
}

func (*AlphabetDraw) isTokenDraw() {}

// MayFallBack counts only a fallback an else names, since the builtin's own default has its shape.
func (d *AlphabetDraw) MayFallBack() bool { return d.strings.Any() && !d.ownShape }

func (r *AlphabetDraw) Call() builtinfunc.Call { return r.call }

func (r *AlphabetDraw) Fallback() Fallback { return r.fallback }

func (r *AlphabetDraw) Plan() *avoid.Plan { return r.plan }

func (r *AlphabetDraw) Quoted() []string { return r.quoted }

func (r *AlphabetDraw) Runs() avoid.Runs { return r.runs }

func (r *AlphabetDraw) Strings() ReadStrings { return r.strings }
