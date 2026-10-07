package proven

import (
	"math"
	"testing"

	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/grammar"
)

func TestLiteralSaysWhichDatatypesItsTextIs(t *testing.T) {
	for _, c := range []struct {
		text string
		not  [datatype.Count]bool
	}{
		{"42", [datatype.Count]bool{datatype.Boolean: true}},
		{"1.5", [datatype.Count]bool{datatype.Integer: true, datatype.Boolean: true}},
		{"true", [datatype.Count]bool{datatype.Integer: true, datatype.Number: true}},
		{"007", [datatype.Count]bool{datatype.Integer: true, datatype.Number: true, datatype.Boolean: true}},
		{"-0", [datatype.Count]bool{datatype.Boolean: true}},
	} {
		v := Literal(c.text)
		for d := datatype.Integer; d < datatype.Count; d++ {
			if (v.Not[d] != "") != c.not[d] {
				t.Errorf("Literal(%q).Not[%s] = %q", c.text, d, v.Not[d])
			}
		}
	}
}

func TestOrKeepsTheBoundsOfBothAndTheFirstReason(t *testing.T) {
	v := Bounded(1, 2, true).Or(Bounded(-3, 0.5, false))
	if v.Lo != -3 || v.Hi != 2 || v.Integral || v.NonZero != 0 {
		t.Errorf("Or = %+v, want [-3, 2], not integral, possibly zero", v)
	}
	if v := Unproven("first").Or(Unproven("second")); v.NotOperand != "first" || v.Not[datatype.Number] != "first" {
		t.Errorf("Or = %+v, want the first reason kept", v)
	}
}

func TestOfCalcBoundsAnExpressionFromItsOperands(t *testing.T) {
	operand := func(name string) Facts {
		return map[string]Facts{"a": Bounded(1, 9, true), "b": Bounded(0, 9, true), "big": Bounded(0, 1e299, true), "word": Unproven("it is text")}[name]
	}
	for expr, want := range map[string]string{
		"a * 2 + 1": "",
		"a / b":     "divides by b, which is not proven nonzero",
		"big * 100": "(big * 100) is not proven within 1e300",
		"big":       "",
		"word + 1":  `operand "word": it is text`,
		"-(a - 10)": "",
	} {
		c, err := grammar.ParseCalc(expr)
		if err != nil {
			t.Fatal(err)
		}
		v, doubt := OfCalc(c.Expr, operand)
		if doubt != want {
			t.Errorf("OfCalc(%s) doubts %q, want %q", expr, doubt, want)
		}
		if expr == "a * 2 + 1" && (v.Lo != 3 || v.Hi != 19 || !v.Integral || v.NonZero != 3) {
			t.Errorf("OfCalc(%s) = %+v, want [3, 19], integral, nonzero by 3", expr, v)
		}
		if expr == "-(a - 10)" && (v.Lo != 1 || v.Hi != 9) {
			t.Errorf("OfCalc(%s) = %+v, want [1, 9]", expr, v)
		}
	}
	lone, _ := grammar.ParseCalc("x")
	if _, doubt := OfCalc(lone.Expr, func(string) Facts { return Bounded(0, 1e301, true) }); doubt != "x is not proven within 1e300" {
		t.Errorf("OfCalc(x) doubts %q, want a lone operand held to the limit", doubt)
	}
}

func TestPrintedNumberWidensItsBoundsByTheRounding(t *testing.T) {
	v := PrintedNumber("float(0,1,1)", Bounded(0.2, 0.8, false), 1)
	if math.Abs(v.Lo-0.15) > 1e-12 || math.Abs(v.Hi-0.85) > 1e-12 || v.Not[datatype.Integer] == "" || v.Not[datatype.Number] != "" {
		t.Errorf("PrintedNumber = %+v, want [0.15, 0.85], a number and no integer", v)
	}
	if v := PrintedNumber("calc(a)", Bounded(1, 2, true), ShortestDecimals); v.Not[datatype.Integer] != "" {
		t.Errorf("PrintedNumber = %+v, want a whole value printed as an integer", v)
	}
	if v := Printing("digits(3)", datatype.String, Bounded(0, 999, true)); v.Not[datatype.Integer] != "{digits(3)} prints text, not an integer" {
		t.Errorf("Printing = %q, want the text named", v.Not[datatype.Integer])
	}
}
