package grammar

import (
	"reflect"
	"testing"
)

func TestParseCalcPlacesEachOperandAtItsFirstReading(t *testing.T) {
	c, err := ParseCalc("b * (a - b) / -2")
	if err != nil {
		t.Fatal(err)
	}
	want := CalcBin{'/', CalcBin{'*', CalcVar{"b", 0}, CalcBin{'-', CalcVar{"a", 1}, CalcVar{"b", 0}}}, CalcNeg{CalcNum(2)}}
	if !reflect.DeepEqual(c.Expr, want) {
		t.Errorf("ParseCalc = %#v, want %#v", c.Expr, want)
	}
	if !reflect.DeepEqual(c.Operands, []string{"b", "a"}) {
		t.Errorf("Operands = %q, want [b a]", c.Operands)
	}
	if got := CalcText(c.Expr); got != "((b * (a - b)) / -2)" {
		t.Errorf("CalcText = %q", got)
	}
}

func TestParseCalcRefusesAnUnfinishedExpression(t *testing.T) {
	for _, expr := range []string{"", " ", "1 +", "(1", "1 2", "a.b", "postal-code *"} {
		if _, err := ParseCalc(expr); err == nil {
			t.Errorf("ParseCalc(%q) = nil error, want one", expr)
		}
	}
}
