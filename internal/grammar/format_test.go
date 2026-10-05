package grammar

import (
	"reflect"
	"testing"
)

// TestSplitArgsQuotesOutsideSelectors pins the arg grammar: a comma splits outside a
// selector and outside a quoted layout, and a quote inside a selector is a name's text.
func TestSplitArgsQuotesOutsideSelectors(t *testing.T) {
	for in, want := range map[string][]string{
		"a, b": {"a", "b"},
		"1990-01-01,1990-12-31,'January 2, 2006'": {"1990-01-01", "1990-12-31", "'January 2, 2006'"},
		"/geo.US.locality[O'Fallon].name, 2":      {"/geo.US.locality[O'Fallon].name", "2"},
		"'[a,b]'":                                 {"'[a,b]'"},
	} {
		if got := splitArgs(in); !reflect.DeepEqual(got, want) {
			t.Errorf("splitArgs(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseFormatReadsEachTokenKind(t *testing.T) {
	toks, err := ParseFormat("a{{{x|/y.z[a.b]}{/p as q}{f(1, n)}")
	if err != nil {
		t.Fatal(err)
	}
	want := []Token{
		{Kind: LiteralRun, Lit: "a{"},
		{Kind: NameRead, Body: "x|/y.z[a.b]", Arms: []string{"x", "/y.z[a.b]"}},
		{Kind: NameBind, Body: "/p as q", BoundRef: "/p", Bound: "q"},
		{Kind: BuiltinCall, Body: "f(1, n)", Fn: "f", Args: []string{"1", "n"}},
	}
	if !reflect.DeepEqual(toks, want) {
		t.Errorf("ParseFormat = %+v, want %+v", toks, want)
	}
}

func TestParseFormatRefusesUnbalancedBraces(t *testing.T) {
	for _, format := range []string{"{a", "a}", "{a{b}}", "{f(}"} {
		if _, err := ParseFormat(format); err == nil {
			t.Errorf("ParseFormat(%q) = nil error, want one", format)
		}
	}
}
