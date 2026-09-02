package fejkdata

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

// engine builds a seeded generator with no loaded categories, for rendering tests.
func engine(seed uint64) *Generator {
	s, err := newRand(seed, true)
	if err != nil {
		panic(err)
	}
	return &Generator{rand: s}
}

// parse unmarshals a JSON template fragment into its dynamic form.
func parse(t *testing.T, s string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return v
}

// compiled parses and compiles a JSON fragment into a node.
func compiled(t *testing.T, s string) node {
	t.Helper()
	n, err := compile(parse(t, s))
	if err != nil {
		t.Fatalf("compile %q: %v", s, err)
	}
	return n
}

func mustRender(t *testing.T, f *Generator, s string) string {
	t.Helper()
	return render(f.rand, compiled(t, s))
}

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

func TestClassBuiltins(t *testing.T) {
	cases := map[string]*regexp.Regexp{
		`"{digits(1)}"`:          regexp.MustCompile(`^[0-9]$`),
		`"{digits(3)}"`:          regexp.MustCompile(`^[0-9]{3}$`),
		`"{int(1,9)}"`:           regexp.MustCompile(`^[1-9]$`),
		`"{upper(1)}"`:           regexp.MustCompile(`^[A-Z]$`),
		`"{lower(1)}"`:           regexp.MustCompile(`^[a-z]$`),
		`"{upper(2)}{lower(2)}"`: regexp.MustCompile(`^[A-Z]{2}[a-z]{2}$`),
	}
	f := engine(7)
	for tmpl, re := range cases {
		for i := 0; i < 100; i++ {
			if got := mustRender(t, f, tmpl); !re.MatchString(got) {
				t.Fatalf("%s produced %q, want %s", tmpl, got, re)
			}
		}
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

func TestWeightZeroIsRejected(t *testing.T) {
	_, err := compile(parse(t, `[{"format":"X","weight":0},"Y"]`))
	if err == nil || !strings.Contains(err.Error(), "weight 0") || !strings.Contains(err.Error(), "remove") {
		t.Fatalf("compile(weight 0) = %v, want it rejected naming the fix", err)
	}
}

func TestWeightSkewsDistribution(t *testing.T) {
	f := engine(9)
	heavy := 0
	for i := 0; i < 1000; i++ {
		if mustRender(t, f, `[{"format":"H","weight":10},"L"]`) == "H" {
			heavy++
		}
	}
	if heavy < 800 { // expected ~909
		t.Fatalf("heavy variant chosen %d/1000, want a clear majority", heavy)
	}
}

func TestRepeatRendersFormatNTimes(t *testing.T) {
	// repeat N concatenates N independent renders of the format ('x','y' are
	// literal, not character classes).
	if got := mustRender(t, engine(1), `{"format":"xy","repeat":3}`); got != "xyxyxy" {
		t.Fatalf("repeat literal = %q, want xyxyxy", got)
	}
	// separator joins renders (N-1 of them, no trailing one).
	if got := mustRender(t, engine(1), `{"format":"xy","repeat":3,"separator":"-"}`); got != "xy-xy-xy" {
		t.Fatalf("repeat with separator = %q, want xy-xy-xy", got)
	}
	f, re := engine(7), regexp.MustCompile(`^[0-9]{4}$`)
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		got := mustRender(t, f, `{"format":"{digits(1)}","repeat":4}`)
		if !re.MatchString(got) {
			t.Fatalf("repeat-4 = %q, want 4 digits", got)
		}
		seen[got] = true
	}
	if len(seen) < 2 {
		t.Fatalf("repeat never varied across renders: %v", seen)
	}
}

func TestFunctionTokenLuhn(t *testing.T) {
	// {luhn()} appends a Luhn check digit over the digits emitted so far in the
	// current expansion (non-digits skipped but kept). It reads the output
	// buffer, so a value is never re-rendered. Bodies are escaped to fix input.
	f := engine(1)
	cases := map[string]string{
		`"811218987{luhn()}"`:                      "8112189876",  // personnummer body
		`"7992739871{luhn()}"`:                     "79927398713", // classic Luhn vector
		`"811218-987{luhn()}"`:                     "811218-9876", // '-' skipped, kept
		`{"format":"{n}{luhn()}","n":"811218987"}`: "8112189876",  // over a rendered token
	}
	for tmpl, want := range cases {
		if got := mustRender(t, f, tmpl); got != want {
			t.Fatalf("%s = %q, want %q", tmpl, got, want)
		}
	}
}

func TestRecursionHasNoDepthLimit(t *testing.T) {
	// Build {format:{a}, a:[{format:{a}, a:[ ... "deep" ]]}} 50 levels deep.
	tmpl := `"deep"`
	for i := 0; i < 50; i++ {
		tmpl = `{"format":"{a}","a":` + tmpl + `}`
	}
	if got := mustRender(t, engine(1), tmpl); got != "deep" {
		t.Fatalf("deep recursion = %q, want deep", got)
	}
}

func TestCompileErrors(t *testing.T) {
	// Every structural problem is caught up front, at compile/New time, never
	// deferred to a random render that happens to hit the bad branch.
	for _, bad := range []string{
		`{"x":"Q"}`,                           // object without "format"
		`{"format":"{y}","x":1}`,              // a field is a bare number
		`[1, 2]`,                              // a choice of numbers
		`5`,                                   // unsupported node type
		`[]`,                                  // empty choice
		`"{x"`,                                // unterminated brace
		`"x}y"`,                               // a lone } must be written }}
		`["x"]`,                               // a one-item choice is its item
		`["a","a"]`,                           // a repeated item is a weight
		`{"format":"x"}`,                      // an object holding only a format is a string
		`[{"format":"a","weight":1},"b"]`,     // weight 1 is the default
		`{"format":"{x}","x":"v","repeat":1}`, // repeat 1 is the default
		`{"format":"{a{b}","a":"Q"}`,          // a brace inside a token
		`"{x}"`,                               // a bare string has no fields to name
		`"{digits(0)}"`,                       // count must be positive
		`"{upper(x)}"`,                        // count must be an integer
		`"{lower()}"`,                         // wrong arity
		`{"format":"{y}","x":"Q"}`,            // token names a missing field
		`"{}"`,                                // empty token name
		`{"format":"{a|}","a":"Q"}`,           // empty alternation segment
		`[{"format":"A","weight":-1},"B"]`,    // negative weight
		`[{"format":"A","weight":0},{"format":"B","weight":0}]`,         // weights sum to zero
		`[{"format":"A","weight":1e308},{"format":"B","weight":1e308}]`, // weights overflow to +Inf
		`{"format":"A","weight":"heavy"}`,                               // non-numeric weight
		`{"format":"x","repeat":0}`,                                     // repeat below 1
		`{"format":"x","repeat":-2}`,                                    // negative repeat
		`{"format":"x","repeat":1.5}`,                                   // non-integer repeat
		`{"format":"x","repeat":"two"}`,                                 // non-numeric repeat
		`{"format":"x","repeat":2,"separator":5}`,                       // non-string separator
		`"{nope()}"`,           // unknown function
		`"{luhn(x)}"`,          // function given args it takes none of
		`"{luhn(}"`,            // malformed function token
		`"{int(1)}"`,           // wrong arity
		`"{int(a,b)}"`,         // non-integer args
		`"{int(5,1)}"`,         // min > max
		`"{hex(0)}"`,           // count must be positive
		`"{nanoid(-1)}"`,       // negative count
		`"{base64(0)}"`,        // count must be positive
		`"{float(1,2)}"`,       // wrong arity
		`"{float(1,2,-1)}"`,    // negative decimals
		`"{float(NaN,NaN,2)}"`, // bounds must be finite
		`"{float(Inf,Inf,2)}"`, // same-sign infinities
		`"{float(1,NaN,2)}"`,   // one NaN bound
		`"{float(-Inf,1,2)}"`,  // one infinite bound
		`"{digits(+5)}"`,       // a count is a plain integer
		`"{digits(05)}"`,       // no leading zero
		`"{int(+1,5)}"`,        // a bound is a plain integer
		`"{int(5,5)}"`,         // a constant is written as text
		`"{float(1,1,2)}"`,     // a constant is written as text
		`{"format":"{x}","x":"v","repeat":2,"separator":""}`, // separator "" is the default
		`"{calc(1/0)}"`, // a constant zero divisor
		`{"format":"{calc(x/y)}","x":"1","y":"0"}`, // a fixed zero divisor
		`"{iban(US)}"`,      // unsupported country
		`"{seq(a,b)}"`,      // seq takes at most one name
		`"{calc()}"`,        // calc needs an expression
		`"{calc(1 +)}"`,     // dangling operator
		`"{calc((1 + 2)}"`,  // unbalanced parenthesis
		`"{calc(1 2)}"`,     // two operands, no operator
		`"{calc(price)}"`,   // operand names no field
		`"{calc(1, 2, 3)}"`, // too many args
		`"{calc(1, x)}"`,    // decimals arg not an integer
		`"{calc(1, -1)}"`,   // decimals negative
	} {
		if _, err := compile(parse(t, bad)); err == nil {
			t.Errorf("compile(%s) = nil error, want error", bad)
		}
	}
}

func TestFakePathNavigation(t *testing.T) {
	f := engine(1)
	f.categories = map[string]node{
		"addr": compiled(t, `{"format":"{street}","street":"Main"}`),
	}
	if got, err := f.Fake("addr"); err != nil || !strings.Contains(got, "Main") {
		t.Fatalf("Fake(addr) = %q, %v", got, err)
	}
	if got, err := f.Fake("addr.street"); err != nil || got != "Main" {
		t.Fatalf("Fake(addr.street) = %q, %v, want Main", got, err)
	}
	if _, err := f.Fake("addr.nope"); err == nil {
		t.Error("Fake(addr.nope) = nil error, want missing-field error")
	}
	if _, err := f.Fake("missing"); err == nil {
		t.Error("Fake(missing) = nil error, want unknown-category error")
	}
}

// TestGrowIsALowerBound pins the property that makes the pre-sized render buffer
// safe: grow must never exceed what expand emits. render multiplies it by repeat,
// so an over-estimate would amplify up to a million-fold.
func TestGrowIsALowerBound(t *testing.T) {
	f := engine(3)
	for _, format := range []string{
		"",
		"plain literal",
		"{digits(2)}-{int(1,9)}{int(1,9)}-{upper(2)}-{lower(2)}",
		"01Aa# literal",
		"{{}} {{{x}}}",
		"Ö dag åäö 日本語",
		"{x}{x}{x}",
		"{hex(8)}-{int(10,99)}-{nanoid(5)}",
		"9{d}{luhn()}",
		"{a|b} and {a|b}",
	} {
		src := `{"format":` + quote(format) + `,"x":"1","a":"A","b":"B","d":"012345678901234"}`
		tmpl, ok := compiled(t, src).(*template)
		if !ok {
			t.Fatalf("format %q did not compile to a template", format)
		}
		for i := 0; i < 50; i++ {
			if got := len(expand(f.rand, tmpl)); got < tmpl.grow {
				t.Errorf("format %q: expand emitted %d bytes, below grow %d", format, got, tmpl.grow)
			}
		}
	}
}

func quote(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(b)
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
