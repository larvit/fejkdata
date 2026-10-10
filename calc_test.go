package fejkdata

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/grammar"
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
	if got := mustRender(t, engine(1), `{"format":"{calc(b * (a - b))}","a":"10","b":"3"}`); got != "21" {
		t.Fatalf("calc(b * (a - b)) = %q, want 21", got)
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
func TestParsedCallReadsOnlyAnOperandBuiltin(t *testing.T) {
	for _, format := range []string{"{luhn()}", "{calc()}", "{calc(1 +)}", "{calc(()}"} {
		if toks, err := grammar.ParseFormat(format); err != nil || len(toks) != 1 || tokenReads(toks[0]) != nil {
			t.Errorf("ParseFormat(%q) = %+v, %v, want one call reading no operand", format, toks, err)
		}
	}
	if toks, err := grammar.ParseFormat("{calc(net * qty)}"); err != nil || len(toks) != 1 || len(tokenReads(toks[0])) != 2 {
		t.Errorf("ParseFormat({calc(net * qty)}) = %+v, %v, want both operands", toks, err)
	}
}

// --- one draw, one value: a calc operand reads the expansion's draw ---

// TestANamedCalcOperandIsTheOneShown pins how a calc computes from the value its format
// shows: the format and the calc read one name.
func TestANamedCalcOperandIsTheOneShown(t *testing.T) {
	dir := writeData(t, map[string]string{
		"inv": `{"format":"{net as n}{qty as q}{n} x {q} = {calc(n * q, 2)}","net":["19.99","5.00","100.00"],"qty":["2","3","7"]}`,
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

// TestACalcOperandDrawsAfresh pins that a calc reading a field draws it as any {…} does.
func TestACalcOperandDrawsAfresh(t *testing.T) {
	dir := writeData(t, map[string]string{"same": `{"format":"{w} {calc(w)}","w":["1","2","3","4","5"]}`})
	f := newGenerator(t, dir, WithSeed(5))
	for i := 0; i < 200; i++ {
		if p := strings.Fields(fake(t, f, "same")); p[0] != p[1] {
			return
		}
	}
	t.Fatal("{w} {calc(w)} never differed in 200 draws, want two independent draws")
}

// TestANamedCalcOperandPicksEachIteration pins that a name a repeat binds picks again on each
// iteration, staying one value within it.
func TestANamedCalcOperandPicksEachIteration(t *testing.T) {
	dir := writeData(t, map[string]string{
		"rep": `{"format":"{n as m}{m}={calc(m * 1)}","repeat":8,"separator":" ","n":["2","3","4","5","6","7","8","9"]}`,
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

// TestTwoNamesPickApart pins that a nested template's name is a pick of its own: the inner
// pair agrees with itself, not with the outer pair.
func TestTwoNamesPickApart(t *testing.T) {
	dir := writeData(t, map[string]string{
		"nest": `{"format":"{v as a}{a}={calc(a * 1)} {inner}","v":["2","3","4","5","6","7","8","9"],
			"inner":{"format":"{v as b}{b}={calc(b * 1)}","v":["2","3","4","5","6","7","8","9"]}}`,
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
		_, err := resolved(t, src)
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
		if _, err := resolved(t, ok); err != nil {
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
		if _, err := resolved(t, src); err == nil || !strings.Contains(err.Error(), want) {
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

func TestCalcCompileErrors(t *testing.T) {
	for _, bad := range []string{
		`"{calc(1/0)}"`, // a constant zero divisor
		`{"format":"{calc(x/y)}","x":"1","y":"0"}`, // a fixed zero divisor
		`"{calc()}"`,        // calc needs an expression
		`"{calc(1 +)}"`,     // dangling operator
		`"{calc((1 + 2)}"`,  // unbalanced parenthesis
		`"{calc(1 2)}"`,     // two operands, no operator
		`"{calc(price)}"`,   // operand names no field
		`"{calc(1, 2, 3)}"`, // too many args
		`"{calc(1, x)}"`,    // decimals arg not an integer
		`"{calc(1, -1)}"`,   // decimals negative
	} {
		if _, err := resolved(t, bad); err == nil {
			t.Errorf("compile(%s) = nil error, want error", bad)
		}
	}
}

func TestCalcReadsAName(t *testing.T) {
	dir := writeData(t, map[string]string{
		"n":    `["2","3","5","7"]`,
		"card": `"{/n as k}{k} x 2 = {calc(k * 2)}"`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	for i := 0; i < 50; i++ {
		var k, double int
		got := fake(t, f, "card")
		if _, err := fmt.Sscanf(got, "%d x 2 = %d", &k, &double); err != nil || double != k*2 {
			t.Fatalf("card = %q, want the calc to compute from the pick {k} prints", got)
		}
	}
}

func TestCalcReadsANameItReadsNowhereElse(t *testing.T) {
	dir := writeData(t, map[string]string{
		"n":    `["2","3"]`,
		"card": `"{/n as k}{calc(k * 2)}"`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	if got := fake(t, f, "card"); got != "4" && got != "6" {
		t.Fatalf("card = %q, want 4 or 6", got)
	}
}

func TestCalcRefusesABadOperand(t *testing.T) {
	for card, want := range map[string]string{
		`"{/word as w}{w}{calc(w * 2)}"`:    `operand "w" is never a number: it renders "def"`,
		`"{/zero as z}{z}{calc(1 / z)}"`:    `divides by z, which is always zero`,
		`"{calc(nope * 2)}"`:                `no field "nope"`,
		`"{/word as w}{w}{calc(nope * 2)}"`: `no field or name "nope"; the names bound here are "w"`,
	} {
		dir := writeData(t, map[string]string{
			"word": `["abc","def"]`,
			"zero": `"0"`,
			"card": card,
		})
		if _, err := New(withShipped(), WithDataPath(dir), WithSeed(1)); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("New with card %s = %v, want an error containing %q", card, err, want)
		}
	}
}

func TestATypedColumnRefusesACalcOverNoDigits(t *testing.T) {
	_, err := New(WithDataPath(writeData(t, map[string]string{
		"row": `{"format":"","x":{"format":"{calc(d + 1)}","d":"{digits(0)}","datatype":"integer"}}`,
	})))
	if err == nil || !strings.Contains(err.Error(), "{digits(0)} prints nothing, which is no number") {
		t.Fatalf("New = %v, want a calc over {digits(0)} refused in a typed column", err)
	}
}

func TestATypedColumnProvesACalcOverAName(t *testing.T) {
	dir := writeData(t, map[string]string{
		"n":     `"{int(1,9)}"`,
		"order": `{"format":"{/n as k}{k}","double":{"format":"{calc(k * 2)}","datatype":"integer"}}`,
	})
	f := newGenerator(t, dir, WithSeed(1))
	r, err := f.FakeRecord("order")
	if err != nil {
		t.Fatal(err)
	}
	if v, err := strconv.Atoi(r.Columns()[0].Value); err != nil || v < 2 || v > 18 || v%2 != 0 {
		t.Fatalf("double = %q, want an even integer in 2..18", r.Columns()[0].Value)
	}
	if !strings.Contains(r.JSON(), `"double":`+r.Columns()[0].Value) {
		t.Fatalf("JSON = %s, want double written as a number", r.JSON())
	}
	bad := writeData(t, map[string]string{
		"word":  `["1","x"]`,
		"order": `{"format":"{/word as w}{w}","double":{"format":"{calc(w * 2)}","datatype":"integer"}}`,
	})
	if _, err := New(withShipped(), WithDataPath(bad), WithSeed(1)); err == nil || !strings.Contains(err.Error(), `operand "w": "x" is not a number`) {
		t.Fatalf("New = %v, want the typed column refused over an operand that is not always a number", err)
	}
}

func TestARepeatOfZeroIsNeverANumber(t *testing.T) {
	if _, err := resolved(t, `{"format":"{calc(x / y)}","x":"1","y":{"format":"0","repeat":0}}`); err == nil || !strings.Contains(err.Error(), `operand "y" is never a number: it renders ""`) {
		t.Errorf("a calc over a repeat of 0 = %v, want the operand refused as never a number", err)
	}
}
