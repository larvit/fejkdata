package tree

import (
	"github.com/larvit/fejkdata/internal/avoid"
	"github.com/larvit/fejkdata/internal/builtinfunc"
)

// AlphabetDraw draws a value holding none of a where clause's strings, for a builtin that draws each
// character from an alphabet, such as {digits(n)}. A string the clause reads from a name is read at
// each draw, and where no value avoids the strings then, the draw prints its fallback.
type AlphabetDraw struct{ spec AlphabetDrawSpec }

// AlphabetDrawSpec is what NewAlphabetDraw builds an AlphabetDraw of.
type AlphabetDrawSpec struct {
	Runs    avoid.Runs
	Quoted  []string    // the strings quoted in the clause, as patterns a value must not hold
	Plan    *avoid.Plan // the draws avoiding Quoted; nil where Quoted is empty
	Call    builtinfunc.Call
	Strings ReadStrings
	// Fallback prints where the strings read leave no value; without an else, it is a value of
	// the builtin's own shape
	Fallback Fallback
	OwnShape bool // no else names the fallback
}

func NewAlphabetDraw(s AlphabetDrawSpec) *AlphabetDraw { return &AlphabetDraw{s} }

func (*AlphabetDraw) isTokenDraw() {}

// MayFallBack counts only a fallback an else names, since the builtin's own default has its shape.
func (d *AlphabetDraw) MayFallBack() bool { return d.spec.Strings.Any() && !d.spec.OwnShape }

func (d *AlphabetDraw) Call() builtinfunc.Call { return d.spec.Call }

func (d *AlphabetDraw) Fallback() Fallback { return d.spec.Fallback }

func (d *AlphabetDraw) Plan() *avoid.Plan { return d.spec.Plan }

func (d *AlphabetDraw) Quoted() []string { return d.spec.Quoted }

func (d *AlphabetDraw) Runs() avoid.Runs { return d.spec.Runs }

func (d *AlphabetDraw) Strings() ReadStrings { return d.spec.Strings }
