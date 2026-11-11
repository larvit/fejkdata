package fejkdata

import (
	"regexp"
	"strings"
	"testing"
)

func TestStringIsAFormat(t *testing.T) {
	f := emptyGenerator(1)
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
	if _, err := resolved(t, `"{x}"`); err == nil || !strings.Contains(err.Error(), `no field "x"`) {
		t.Errorf(`compile("{x}") = %v, want a no-field error`, err)
	}
}

func TestTextIsLiteral(t *testing.T) {
	if got := mustRender(t, emptyGenerator(1), `{"format":"100 Main St #1 {x}","x":"A"}`); got != "100 Main St #1 A" {
		t.Fatalf("text = %q, want it verbatim", got)
	}
}

func TestBraceEscapes(t *testing.T) {
	f := emptyGenerator(1)
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
		if _, err := resolved(t, src); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("compile(%s) = %v, want an error mentioning %s", src, err, want)
		}
	}
}

func TestTokenSubstitution(t *testing.T) {
	if got := mustRender(t, emptyGenerator(1), `{"format":"{x}sson","x":"Erik"}`); got != "Eriksson" {
		t.Fatalf("token = %q, want Eriksson", got)
	}
}

func TestAlternationDrawsOneField(t *testing.T) {
	f := emptyGenerator(3)
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		seen[mustRender(t, f, `{"format":"{a|b}","a":"A","b":"B"}`)] = true
	}
	if !seen["A"] || !seen["B"] || len(seen) != 2 {
		t.Fatalf("alternation produced %v, want both A and B", seen)
	}
}

func TestAlternationThreeWay(t *testing.T) {
	f := emptyGenerator(4)
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		seen[mustRender(t, f, `{"format":"{a|b|c}","a":"A","b":"B","c":"C"}`)] = true
	}
	if !seen["A"] || !seen["B"] || !seen["C"] || len(seen) != 3 {
		t.Fatalf("3-way alternation produced %v, want A, B and C", seen)
	}
}

func TestHashIsLiteral(t *testing.T) {
	f := emptyGenerator(1)
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
	got := mustRender(t, emptyGenerator(2), `{"format":"Öster{x}-{digits(1)}å","x":"väg"}`)
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
		`"{lower()}"`,                // wrong arity
		`"{nope()}"`,                 // unknown function
		`"{luhn(x)}"`,                // function given args it takes none of
		`"{int(1)}"`,                 // wrong arity
		`"{float(1,2)}"`,             // wrong arity
	} {
		if _, err := resolved(t, bad); err == nil {
			t.Errorf("compile(%s) = nil error, want error", bad)
		}
	}
}

func TestAReadOfAFieldBesideItsTemplateNamesTheRuleAndTheFix(t *testing.T) {
	const rule = "whose format holds it"
	for _, c := range []struct {
		name, src, holder, bind string
	}{
		{"a bare string", `{"format":"{a}","a":"x{b}","b":"y"}`, `field "a"`, `{b as n}`},
		{"a bare string reading a name", `{"format":"{w as p}{a}","w":["v","z"],"a":"{p}{b}","b":"y"}`, `field "a"`, `{b as n}`},
		{"a path into the field", `{"format":"{a}","a":"x{b.c}","b":{"format":"{c}","c":"y"}}`, `field "a"`, `{b as n}`},
		{"an item of a choice in the field", `{"format":"{a}","a":["x{b}","z"],"b":"y"}`, `the item of field "a"`, `{b as n}`},
		{"a field of a choice's item", `{"format":"{c}","c":[{"format":"{a}","a":"x{b}","b":"y"},"z"]}`, `field "a"`, ""},
	} {
		_, err := resolved(t, c.src)
		if err == nil {
			t.Errorf("%s: compile(%s) loaded, want a refusal", c.name, c.src)
			continue
		}
		msg := err.Error()
		if !strings.Contains(msg, rule) || !strings.Contains(msg, c.holder) || !strings.Contains(msg, `"b"`) {
			t.Errorf("%s: compile(%s) = %v, want it to state the rule %q and name %s and \"b\"", c.name, c.src, err, rule, c.holder)
		}
		if c.bind == "" && strings.Contains(msg, " as n}") {
			t.Errorf("%s: compile(%s) = %v, want no binding fix inside a choice's item", c.name, c.src, err)
		} else if !strings.Contains(msg, c.bind) {
			t.Errorf("%s: compile(%s) = %v, want the binding fix %s", c.name, c.src, err, c.bind)
		}
	}
	if _, err := resolved(t, `{"format":"{a}","a":"x{zzz}","b":"y"}`); err == nil || strings.Contains(err.Error(), rule) {
		t.Errorf("compile(a read of a field nowhere) = %v, want a refusal that names no field beside", err)
	}
	for _, src := range []string{
		`{"format":"{a}","a":{"format":"x{b}","b":"y"}}`,
		`{"format":"{b as n}{a}","a":"x{n}","b":"y"}`,
		`{"format":"{a}","a":[{"format":"x{b}","b":"y"}]}`,
		`{"format":"{b as n}{a}","a":["x{n}"],"b":"y"}`,
		`{"format":"{b as n}{a}","a":"x{n.c}","b":{"format":"{c}","c":"y"}}`,
	} {
		if got := mustRender(t, emptyGenerator(1), src); got != "xy" {
			t.Errorf("render(%s) = %q, want the fix the refusal names to print xy", src, got)
		}
	}
}
