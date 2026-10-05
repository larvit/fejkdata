package grammar

import (
	"reflect"
	"testing"
)

func TestParseCalcPlacesEachOperandAtItsFirstReading(t *testing.T) {
	n, err := ParseCalc("b * (a - b) / -2")
	if err != nil {
		t.Fatal(err)
	}
	want := CalcBin{'/', CalcBin{'*', CalcVar{"b", 0}, CalcBin{'-', CalcVar{"a", 1}, CalcVar{"b", 0}}}, CalcNeg{CalcNum(2)}}
	if !reflect.DeepEqual(n, want) {
		t.Errorf("ParseCalc = %#v, want %#v", n, want)
	}
	if got := CalcVars(n); !reflect.DeepEqual(got, []string{"b", "a"}) {
		t.Errorf("CalcVars = %q, want [b a]", got)
	}
	if got := CalcText(n); got != "((b * (a - b)) / -2)" {
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
