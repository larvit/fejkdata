package fejkdata

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

// TestCalcArithmetic pins the operators, precedence, parentheses and unary minus
// over number literals.
func TestCalcArithmetic(t *testing.T) {
	f := engine(1)
	cases := map[string]string{
		`"{calc(2 + 3)}"`:       "5",
		`"{calc(2 * 3 + 4)}"`:   "10", // * binds tighter than +
		`"{calc(2 + 3 * 4)}"`:   "14",
		`"{calc((2 + 3) * 4)}"`: "20", // parentheses override
		`"{calc(10 / 4)}"`:      "2.5",
		`"{calc(-2 + 5)}"`:      "3", // unary minus
		`"{calc(2 - -3)}"`:      "5",
		`"{calc(1.5 * 2)}"`:     "3", // whole result drops the decimals
	}
	for tmpl, want := range cases {
		if got := mustRender(t, f, tmpl); got != want {
			t.Errorf("%s = %q, want %q", tmpl, got, want)
		}
	}
}

// TestCalcAuto pins the default (no-dp) rendering: minimal decimal form, no
// scientific notation, whole numbers without a fraction.
func TestCalcAuto(t *testing.T) {
	f := engine(1)
	cases := map[string]string{
		`"{calc(10 / 3)}"`: "3.3333333333333335",
		`"{calc(6 / 2)}"`:  "3",
		`"{calc(1 / 4)}"`:  "0.25",
		`"{calc(0 * -1)}"`: "0",
	}
	for tmpl, want := range cases {
		if got := mustRender(t, f, tmpl); got != want {
			t.Errorf("%s = %q, want %q", tmpl, got, want)
		}
	}
}

// TestCalcDecimals pins the optional decimals arg: rounds to dp places, dp 0
// drops the fraction.
func TestCalcDecimals(t *testing.T) {
	f := engine(1)
	cases := map[string]string{
		`"{calc(10 / 3, 2)}"`: "3.33",
		`"{calc(10 / 3, 0)}"`: "3",
		`"{calc(2 * 3, 2)}"`:  "6.00",
		`"{calc(-0.001, 2)}"`: "0.00",
	}
	for tmpl, want := range cases {
		if got := mustRender(t, f, tmpl); got != want {
			t.Errorf("%s = %q, want %q", tmpl, got, want)
		}
	}
}

// TestCalcFields pins that bare names resolve to sibling fields, rendered then
// parsed as numbers.
func TestCalcFields(t *testing.T) {
	if got := mustRender(t, engine(1), `{"format":"{calc(price * qty, 2)}","price":"19.99","qty":"3"}`); got != "59.97" {
		t.Fatalf("calc over fields = %q, want 59.97", got)
	}
	// A field that is itself a template renders before parsing.
	if got := mustRender(t, engine(1), `{"format":"{calc(a - b)}","a":{"format":"{n}","n":"10"},"b":"3"}`); got != "7" {
		t.Fatalf("calc(a - b) = %q, want 7", got)
	}
}

// TestCalcNonNumericIsNaN pins the never-fail rule: a field that sometimes does
// not render to a number becomes NaN then, which propagates and prints visibly.
func TestCalcNonNumericIsNaN(t *testing.T) {
	f := engine(1)
	for i := 0; i < 50; i++ {
		if got := mustRender(t, f, `{"format":"{calc(x * 2)}","x":["abc","1"]}`); got == "NaN" {
			return
		}
	}
	t.Fatal("calc over a sometimes non-numeric field never printed NaN in 50 draws")
}

// TestCalcReproducible pins that a calc over a random operand stays seed-stable.
func TestCalcReproducible(t *testing.T) {
	tmpl := `{"format":"{calc(q * 2 + 1)}","q":"{int(1,1000000)}"}`
	if a, b := mustRender(t, engine(7), tmpl), mustRender(t, engine(7), tmpl); a != b {
		t.Fatalf("calc not reproducible: %q != %q", a, b)
	}
}

// TestParsedCallReadsOnlyAnOperandBuiltin pins the operands a parsed call carries for a
// call that is not a calc, and for one whose expression does not parse: none, either way.
// checkCalc is what reports a bad expression, so the callers that run after it
// never meet one — but they must not have to depend on that order to be safe.
func TestParsedCallReadsOnlyAnOperandBuiltin(t *testing.T) {
	for _, format := range []string{"{luhn()}", "{calc()}", "{calc(1 +)}", "{calc(()}"} {
		if toks, err := parseFormat(format); err != nil || len(toks) != 1 || toks[0].names != nil {
			t.Errorf("parseFormat(%q) = %+v, %v, want one call reading no operand", format, toks, err)
		}
	}
	if toks, err := parseFormat("{calc(net * qty)}"); err != nil || len(toks) != 1 || len(toks[0].names) != 2 {
		t.Errorf("parseFormat({calc(net * qty)}) = %+v, %v, want both operands", toks, err)
	}
}

// TestCalcGuardsPanic pins the guards on calc's own invariants: checkCalc parsed
// the expression before prep sees it, and indexVars places every name calcVars
// read, so a break in either is reported rather than computed around.
func TestCalcGuardsPanic(t *testing.T) {
	for name, call := range map[string]func(){
		"prep on an expression that does not parse": func() { calcPrep([]string{"1 +"}) },
		"an operand name never placed":              func() { calcVar("n").eval(nil) },
		"indexVars on a name it did not read":       func() { indexVars(calcVar("n"), map[string]int{}) },
	} {
		mustPanic(t, name, call)
	}
}

// --- one draw, one value: a calc operand reads the expansion's draw ---

// TestCalcOperandReadsTheExpansionsDraw pins the one-draw rule for a calc operand.
// A field the format renders and a calc reads is drawn once per expansion, so the
// operand shown is the operand computed — the correlation the dotted-path rule
// already gives a level (see bound_test.go), applied to a plain sibling.
func TestCalcOperandReadsTheExpansionsDraw(t *testing.T) {
	dir := writeData(t, map[string]string{
		"inv": `{"format":"{net} x {qty} = {calc(net * qty, 2)}","net":["19.99","5.00","100.00"],"qty":["2","3","7"]}`,
	})
	f := newGenerator(t, dir, WithSeed(3))
	for i := 0; i < 300; i++ {
		got := fake(t, f, "inv")
		var net, qty, want float64
		if _, err := fmt.Sscanf(got, "%g x %g = %g", &net, &qty, &want); err != nil {
			t.Fatalf("inv = %q, unparseable: %v", got, err)
		}
		if math.Abs(net*qty-want) > 1e-9 {
			t.Fatalf("inv = %q: the shown %g x %g is not the computed %g", got, net, qty, want)
		}
	}
}

// TestCalcOperandSharesOneDraw pins the reach of that hold: the draw belongs to the
// expansion, not to the calc, so the bare token rendering the same name reads it too.
func TestCalcOperandSharesOneDraw(t *testing.T) {
	dir := writeData(t, map[string]string{
		"same": `{"format":"{w} {calc(w)}","w":["1","2","3","4","5"]}`,
	})
	f := newGenerator(t, dir, WithSeed(5))
	for i := 0; i < 200; i++ {
		got := fake(t, f, "same")
		if p := strings.Fields(got); len(p) != 2 || p[0] != p[1] {
			t.Fatalf("same = %q, want one value twice", got)
		}
	}
}

// TestFieldNoCalcReadsDrawsEachTime guards the boundary: only a name a calc reads is
// held, so an ordinary {w} {w} still draws twice.
func TestFieldNoCalcReadsDrawsEachTime(t *testing.T) {
	dir := writeData(t, map[string]string{"two": `{"format":"{w} {w}","w":["1","2","3","4","5"]}`})
	f := newGenerator(t, dir, WithSeed(5))
	for i := 0; i < 200; i++ {
		if p := strings.Fields(fake(t, f, "two")); p[0] != p[1] {
			return
		}
	}
	t.Fatal("{w} {w} never differed in 200 draws, want two independent draws")
}

// TestCalcHoldIsPerExpansion pins the scope of the hold: each repeat iteration is
// its own expansion, so it draws again while staying self-consistent.
func TestCalcHoldIsPerExpansion(t *testing.T) {
	dir := writeData(t, map[string]string{
		"rep": `{"format":"{n}={calc(n * 1)}","repeat":8,"separator":" ","n":["2","3","4","5","6","7","8","9"]}`,
	})
	f := newGenerator(t, dir, WithSeed(11))
	varied := false
	for i := 0; i < 50; i++ {
		got := fake(t, f, "rep")
		seen := map[string]bool{}
		for _, pair := range strings.Fields(got) {
			shown, computed, ok := strings.Cut(pair, "=")
			if !ok || shown != computed {
				t.Fatalf("rep = %q: %q disagrees within one iteration", got, pair)
			}
			seen[shown] = true
		}
		varied = varied || len(seen) > 1
	}
	if !varied {
		t.Fatal("every repeat iteration drew alike, want an independent draw each")
	}
}

// TestCalcHoldIsPerTemplate pins that a nested template holds its own: the inner
// {v} and its calc agree with each other, not with the outer pair.
func TestCalcHoldIsPerTemplate(t *testing.T) {
	dir := writeData(t, map[string]string{
		"nest": `{"format":"{v}={calc(v * 1)} {inner}","v":["2","3","4","5","6","7","8","9"],
			"inner":{"format":"{v}={calc(v * 1)}","v":["2","3","4","5","6","7","8","9"]}}`,
	})
	f := newGenerator(t, dir, WithSeed(13))
	differed := false
	for i := 0; i < 200; i++ {
		got := fake(t, f, "nest")
		outer, inner, ok := strings.Cut(got, " ")
		if !ok {
			t.Fatalf("nest = %q, want two pairs", got)
		}
		for _, pair := range []string{outer, inner} {
			if shown, computed, ok := strings.Cut(pair, "="); !ok || shown != computed {
				t.Fatalf("nest = %q: %q disagrees within its own expansion", got, pair)
			}
		}
		differed = differed || outer != inner
	}
	if !differed {
		t.Fatal("the nested template never differed from its parent, want its own draw")
	}
}

func TestCalcOverANeverNumericOperandIsRejected(t *testing.T) {
	for src, want := range map[string]string{
		`{"format":"{calc(x * 2)}","x":"abc"}`:           `"abc"`,
		`{"format":"{calc(x * 2)}","x":["a","b"]}`:       `"x"`,
		`{"format":"{calc(x + y)}","x":"1","y":"{{2}}"}`: `"{2}"`,
	} {
		_, err := compile(parse(t, src))
		if err == nil || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), "never a number") {
			t.Errorf("compile(%s) = %v, want the operand rejected naming %s", src, err, want)
		}
	}
	for _, ok := range []string{
		`{"format":"{calc(x * 2)}","x":"3"}`,
		`{"format":"{calc(x * 2)}","x":" 2.5 "}`,
		`{"format":"{calc(x * 2)}","x":["1","abc"]}`,
		`{"format":"{calc(x * 2)}","x":"{digits(2)}"}`,
	} {
		if _, err := compile(parse(t, ok)); err != nil {
			t.Errorf("compile(%s) = %v, want it accepted", ok, err)
		}
	}
}

func TestCalcConstantZeroDivisorIsRejected(t *testing.T) {
	for src, want := range map[string]string{
		`"{calc(1/0)}"`:     "divides by 0",
		`"{calc(2/(1-1))}"`: "divides by (1 - 1)",
		`{"format":"{calc(x/y)}","x":"1","y":"0"}`:       "divides by y",
		`{"format":"{calc(x/(y*2))}","x":"1","y":" 0 "}`: "divides by (y * 2)",
		`{"format":"{calc((a/0)+b)}","a":"1","b":"2"}`:   "divides by 0",
		`{"format":"{calc(a/-0)}","a":"1"}`:              "divides by -0",
	} {
		if _, err := compile(parse(t, src)); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("compile(%s) = %v, want the constant zero divisor rejected naming %q", src, err, want)
		}
	}
	if _, err := compile(parse(t, `{"format":"{calc(a/(b*c))}","a":"1","b":"0","c":["1","2"]}`)); err != nil {
		t.Errorf("compile(a/(b*c)) = %v, want a divisor that varies accepted", err)
	}
	f := engine(1)
	for i := 0; i < 50; i++ {
		if got := mustRender(t, f, `{"format":"{calc(x/y)}","x":"1","y":["0","1"]}`); got == "+Inf" {
			return
		}
	}
	t.Fatal("a sometimes-zero divisor never printed +Inf in 50 draws")
}
