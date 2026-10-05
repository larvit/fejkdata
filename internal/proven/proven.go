// Package proven holds what a proof knows of every render of a value, and the arithmetic
// that bounds a calc from its operands.
package proven

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// Facts is what a proof knows of every render of a node: bounds on the number each render
// reads as, and, per datatype, why some render's text is not of it ("" when every render's is).
type Facts struct {
	Lo, Hi     float64
	NonZero    float64 // every value is at least this far from zero; 0 when one can be zero
	Integral   bool
	NotOperand string // why some render reads as no finite number, the way calc reads it
	Not        [datatype.Count]string
	Nullable   bool // some draw of a column is null
}

// ShortestDecimals prints the fewest digits that read back as the same float64.
const ShortestDecimals = -1

// limit is the largest magnitude a proof accepts as finite, far enough below
// math.MaxFloat64 that rounding in the bounds cannot hide an overflow.
const limit = 1e300

// Or is what a proof knows of a render that is either v or w.
func (v Facts) Or(w Facts) Facts {
	v.Lo, v.Hi, v.NonZero = min(v.Lo, w.Lo), max(v.Hi, w.Hi), min(v.NonZero, w.NonZero)
	v.Integral, v.Nullable = v.Integral && w.Integral, v.Nullable || w.Nullable
	if v.NotOperand == "" {
		v.NotOperand = w.NotOperand
	}
	for d := range v.Not {
		if v.Not[d] == "" {
			v.Not[d] = w.Not[d]
		}
	}
	return v
}

// OfCalc bounds a calc expression from its operands, or says why it cannot.
func OfCalc(expr grammar.CalcNode, operand func(name string) Facts) (Facts, string) {
	v, doubt := bound(expr, operand)
	if doubt == "" && !(magnitude(v) <= limit) {
		doubt = grammar.CalcText(expr) + " is not proven within 1e300"
	}
	return v, doubt
}

func bound(n grammar.CalcNode, operand func(name string) Facts) (Facts, string) {
	switch n := n.(type) {
	case grammar.CalcNum:
		v := float64(n)
		return Bounded(v, v, v == math.Trunc(v)), ""
	case grammar.CalcVar:
		v := operand(n.Name)
		if v.NotOperand != "" {
			return Facts{}, fmt.Sprintf("operand %q: %s", n.Name, v.NotOperand)
		}
		return Facts{Lo: v.Lo, Hi: v.Hi, NonZero: v.NonZero, Integral: v.Integral}, ""
	case grammar.CalcNeg:
		v, doubt := bound(n.X, operand)
		v.Lo, v.Hi = -v.Hi, -v.Lo
		return v, doubt
	case grammar.CalcBin:
		l, doubt := bound(n.L, operand)
		if doubt != "" {
			return l, doubt
		}
		r, doubt := bound(n.R, operand)
		if doubt != "" {
			return r, doubt
		}
		return combine(n, l, r)
	}
	panic(invariant.Broken("calc node %T has no bound", n))
}

// combine bounds one operation from the bounds of its sides.
func combine(n grammar.CalcBin, l, r Facts) (Facts, string) {
	var v Facts
	integral := l.Integral && r.Integral
	switch n.Operator {
	case '+':
		v = Bounded(l.Lo+r.Lo, l.Hi+r.Hi, integral)
	case '-':
		v = Bounded(l.Lo-r.Hi, l.Hi-r.Lo, integral)
	case '*':
		v = Bounded(min(l.Lo*r.Lo, l.Lo*r.Hi, l.Hi*r.Lo, l.Hi*r.Hi), max(l.Lo*r.Lo, l.Lo*r.Hi, l.Hi*r.Lo, l.Hi*r.Hi), integral)
		v.NonZero = max(v.NonZero, l.NonZero*r.NonZero)
	default:
		if r.NonZero == 0 {
			return v, fmt.Sprintf("divides by %s, which is not proven nonzero", grammar.CalcText(n.R))
		}
		m := magnitude(l) / r.NonZero
		v = Facts{Lo: -m, Hi: m, NonZero: l.NonZero / magnitude(r)}
	}
	if !(magnitude(v) <= limit) {
		return v, grammar.CalcText(n) + " is not proven within 1e300"
	}
	return v, ""
}

// Bounded is a number in [lo, hi], its distance from zero read off the bounds.
func Bounded(lo, hi float64, integral bool) Facts {
	v := Facts{Lo: lo, Hi: hi, Integral: integral}
	switch {
	case lo > 0:
		v.NonZero = lo
	case hi < 0:
		v.NonZero = -hi
	}
	return v
}

func magnitude(v Facts) float64 { return math.Max(math.Abs(v.Lo), math.Abs(v.Hi)) }

// PrintedNumber is what a token printing v to dp decimals holds.
func PrintedNumber(token string, v Facts, dp int) Facts {
	if dp == ShortestDecimals {
		if v.Integral { // a whole value prints with no point
			return printedInteger(token, v)
		}
		return Printing(token, datatype.Number, v)
	}
	half, _ := strconv.ParseFloat("5e-"+strconv.Itoa(dp+1), 64)
	v = Facts{Lo: v.Lo - half, Hi: v.Hi + half, NonZero: math.Max(0, v.NonZero-half), Integral: v.Integral || dp == 0}
	if dp == 0 {
		return printedInteger(token, v)
	}
	return Printing(token, datatype.Number, v)
}

func printedInteger(token string, v Facts) Facts {
	v = Printing(token, datatype.Integer, v)
	if !(magnitude(v) < math.MaxInt64) {
		v.Not[datatype.Integer] = fmt.Sprintf("{%s} is not proven within int64", token)
	}
	return v
}

// Printing is v for a token whose every render is text of datatype prints, with a reason
// against each datatype that text is not.
func Printing(token string, prints datatype.DataType, v Facts) Facts {
	for d := datatype.Integer; d < datatype.Count; d++ {
		if prints != d && !(prints == datatype.Integer && d == datatype.Number) {
			v.Not[d] = fmt.Sprintf("{%s} prints %s, not %s", token, datatype.Noun(prints), datatype.Noun(d))
		}
	}
	return v
}

// Unproven is a render no datatype and no calc can take, for why.
func Unproven(why string) Facts {
	v := Facts{NotOperand: why}
	for d := datatype.Integer; d < datatype.Count; d++ {
		v.Not[d] = why
	}
	return v
}

var (
	integerText = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)
	numberText  = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)
)

// Literal proves fixed text: the number calc reads it as, and each datatype it is.
func Literal(text string) Facts {
	var v Facts
	if f, err := strconv.ParseFloat(strings.TrimSpace(text), 64); err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		v.NotOperand = fmt.Sprintf("%q is not a number", text)
	} else {
		v = Bounded(f, f, f == math.Trunc(f))
	}
	if _, err := strconv.ParseInt(text, 10, 64); !integerText.MatchString(text) {
		v.Not[datatype.Integer] = fmt.Sprintf("%q is not an integer", text)
	} else if err != nil {
		v.Not[datatype.Integer] = fmt.Sprintf("%q is past the int64 range", text)
	}
	if v.NotOperand != "" || !numberText.MatchString(text) {
		v.Not[datatype.Number] = fmt.Sprintf("%q is not a number", text)
	}
	if text != "true" && text != "false" {
		v.Not[datatype.Boolean] = fmt.Sprintf("%q is not a boolean", text)
	}
	return signedZero(text, v)
}

// signedZero refuses a zero written with a sign as a typed value, naming it unsigned.
func signedZero(text string, v Facts) Facts {
	mantissa, _, _ := strings.Cut(strings.ToLower(text), "e")
	if !strings.HasPrefix(text, "-") || v.NotOperand != "" || strings.Trim(mantissa, "-0.") != "" {
		return v
	}
	for _, d := range []datatype.DataType{datatype.Integer, datatype.Number} {
		if v.Not[d] == "" {
			v.Not[d] = fmt.Sprintf("%q is zero written with a sign; write %q", text, text[1:])
		}
	}
	return v
}
