package fejkdata

import (
	"regexp"
	"strings"
	"testing"
)

func TestAnOptionOfTheWrongKindIsNamedByItsJSONKind(t *testing.T) {
	if _, err := resolved(t, `{"format":"x","repeat":2,"separator":5}`); err == nil || !strings.Contains(err.Error(), "separator must be a string, not a number") {
		t.Errorf("separator 5 = %v, want the kind named", err)
	}
	_, err := New(WithoutShippedData(), WithDataPath(writeFiles(t, map[string]string{"t.json": `{"format":"{a}","rows":"t.tsv","key":5}`, "t.tsv": "a\nx\n"})))
	if err == nil || !strings.Contains(err.Error(), "key must be a string, not a number") {
		t.Errorf("key 5 = %v, want the kind named", err)
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

func TestNodeCompileErrors(t *testing.T) {
	// Every structural problem is caught up front, at compile/New time, never
	// deferred to a random render that happens to hit the bad branch.
	for _, bad := range []string{
		`{"x":"Q"}`,                        // object without "format"
		`{"format":"{y}","x":1}`,           // a field is a bare number
		`[]`,                               // empty choice
		`[{"format":"A","weight":-1},"B"]`, // negative weight
		`[{"format":"A","weight":0},{"format":"B","weight":0}]`,         // weights sum to zero
		`[{"format":"A","weight":1e308},{"format":"B","weight":1e308}]`, // weights overflow to +Inf
		`{"format":"A","weight":"heavy"}`,                               // non-numeric weight
		`{"format":"x","repeat":-2}`,                                    // negative repeat
		`{"format":"x","repeat":1.5}`,                                   // non-integer repeat
		`{"format":"x","repeat":"two"}`,                                 // non-numeric repeat
		`{"format":"x","repeat":2,"separator":5}`,                       // non-string separator
	} {
		if _, err := resolved(t, bad); err == nil {
			t.Errorf("compile(%s) = nil error, want error", bad)
		}
	}
}

func TestEveryFormatCompilesAtResolve(t *testing.T) {
	n, err := compile(parse(t, `{"format":"{a} {b}","a":{"format":"x{c}","c":["1","2"]},"b":"{/w}"}`))
	if err != nil {
		t.Fatal(err)
	}
	w, err := compile(parse(t, `["p","q"]`))
	if err != nil {
		t.Fatal(err)
	}
	compiledByPath := func() map[string]bool {
		compiled := map[string]bool{}
		if err := eachNode(n, "t", func(label string, m node) error {
			if tm, ok := m.(*template); ok {
				compiled[label+" "+tm.format] = tm.compiled.ops != nil
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		return compiled
	}
	for format, has := range compiledByPath() {
		if has {
			t.Errorf("%s compiled before resolve", format)
		}
	}
	if err := resolveInlineTemplates(inlineNodes(n, "t"), map[string]node{"w": w}); err != nil {
		t.Fatal(err)
	}
	for format, has := range compiledByPath() {
		if !has {
			t.Errorf("%s not compiled at resolve", format)
		}
	}
}
