package fejkdata

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestStringIsAFormat(t *testing.T) {
	f := engine(1)
	for src, want := range map[string]string{
		`"Malmö"`:                  "Malmö",
		`"100 Main St, Apt 1A #0"`: "100 Main St, Apt 1A #0",
		`"{{x}}"`:                  "{x}",
	} {
		if got := mustRender(t, f, src); got != want {
			t.Errorf("render(%s) = %q, want %q", src, got, want)
		}
	}
	if got := mustRender(t, f, `"{digits(3)}"`); !regexp.MustCompile(`^[0-9]{3}$`).MatchString(got) {
		t.Errorf(`render("{digits(3)}") = %q, want three digits`, got)
	}
	if _, err := compile(parse(t, `"{x}"`)); err == nil || !strings.Contains(err.Error(), `no field "x"`) {
		t.Errorf(`compile("{x}") = %v, want a no-field error`, err)
	}
}

func TestTextIsLiteral(t *testing.T) {
	if got := mustRender(t, engine(1), `{"format":"100 Main St #1 {x}","x":"A"}`); got != "100 Main St #1 A" {
		t.Fatalf("text = %q, want it verbatim", got)
	}
}

func TestBraceEscapes(t *testing.T) {
	f := engine(1)
	for src, want := range map[string]string{
		`{"format":"{{","x":"v"}`:         "{",
		`{"format":"}}","x":"v"}`:         "}",
		`{"format":"{{x}}","x":"v"}`:      "{x}",
		`{"format":"{{{x}}}","x":"v"}`:    "{v}",
		`{"format":"a{{{{b}}}}","x":"v"}`: "a{{b}}",
	} {
		if got := mustRender(t, f, src); got != want {
			t.Errorf("render(%s) = %q, want %q", src, got, want)
		}
	}
	for src, want := range map[string]string{
		`"x}y"`:                      "}}",
		`"}"`:                        "}}",
		`{"format":"{a{b}","a":"Q"}`: "'{'",
		`"{x"`:                       "unterminated",
	} {
		if _, err := compile(parse(t, src)); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("compile(%s) = %v, want an error mentioning %s", src, err, want)
		}
	}
}

func TestTokenSubstitution(t *testing.T) {
	if got := mustRender(t, engine(1), `{"format":"{x}sson","x":"Erik"}`); got != "Eriksson" {
		t.Fatalf("token = %q, want Eriksson", got)
	}
}

func TestAlternationPicksOneField(t *testing.T) {
	f := engine(3)
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		seen[mustRender(t, f, `{"format":"{a|b}","a":"A","b":"B"}`)] = true
	}
	if !seen["A"] || !seen["B"] || len(seen) != 2 {
		t.Fatalf("alternation produced %v, want both A and B", seen)
	}
}

func TestAlternationThreeWay(t *testing.T) {
	f := engine(4)
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		seen[mustRender(t, f, `{"format":"{a|b|c}","a":"A","b":"B","c":"C"}`)] = true
	}
	if !seen["A"] || !seen["B"] || !seen["C"] || len(seen) != 3 {
		t.Fatalf("3-way alternation produced %v, want A, B and C", seen)
	}
}

func TestHashIsLiteral(t *testing.T) {
	f := engine(1)
	cases := map[string]string{
		`{"format":"","x":"v"}`:         "",
		`{"format":"#","x":"v"}`:        "#",
		`{"format":"##","x":"v"}`:       "##",
		`{"format":"#0#1#A#a","x":"v"}`: "#0#1#A#a",
		`{"format":"#{x}","x":"v"}`:     "#v",
	}
	for tmpl, want := range cases {
		if got := mustRender(t, f, tmpl); got != want {
			t.Errorf("render(%s) = %q, want %q", tmpl, got, want)
		}
	}
}

func TestMultibyteFormat(t *testing.T) {
	// Scanning is rune-aware: multibyte literals coexist with class chars and
	// tokens without corrupting indices.
	got := mustRender(t, engine(2), `{"format":"Öster{x}-{digits(1)}å","x":"väg"}`)
	if !regexp.MustCompile(`^Österväg-[0-9]å$`).MatchString(got) {
		t.Fatalf("multibyte format = %q", got)
	}
}

func TestFormatCompileErrors(t *testing.T) {
	for _, bad := range []string{
		`"{x"`,                       // unterminated brace
		`"x}y"`,                      // a lone } must be written }}
		`{"format":"{a{b}","a":"Q"}`, // a brace inside a token
		`"{x}"`,                      // a bare string has no fields to name
		`{"format":"{y}","x":"Q"}`,   // token names a missing field
		`"{}"`,                       // empty token name
		`{"format":"{a|}","a":"Q"}`,  // empty alternation segment
		`"{luhn(}"`,                  // malformed function token
	} {
		if _, err := compile(parse(t, bad)); err == nil {
			t.Errorf("compile(%s) = nil error, want error", bad)
		}
	}
}

func TestArgErrorsNameTheSpelling(t *testing.T) {
	for src, want := range map[string]string{
		`"{float(1,2,02)}"`:                 "write 2",
		`{"format":"{calc(a,02)}","a":"1"}`: "write 2",
		`"{digits(+5)}"`:                    "write 5",
		`"{hex(99999999999999999999)}"`:     "exceeds the maximum",
		`"{int(007,9)}"`:                    "write 7",
	} {
		if _, err := compile(parse(t, src)); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("compile(%s) = %v, want an error saying %q", src, err, want)
		}
	}
}

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
