package fejkdata

import (
	"regexp"
	"strings"
	"testing"
)

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

func TestNodeCompileErrors(t *testing.T) {
	// Every structural problem is caught up front, at compile/New time, never
	// deferred to a random render that happens to hit the bad branch.
	for _, bad := range []string{
		`{"x":"Q"}`,                           // object without "format"
		`{"format":"{y}","x":1}`,              // a field is a bare number
		`[1, 2]`,                              // a choice of numbers
		`5`,                                   // unsupported node type
		`[]`,                                  // empty choice
		`["x"]`,                               // a one-item choice is its item
		`["a","a"]`,                           // a repeated item is a weight
		`{"format":"x"}`,                      // an object holding only a format is a string
		`[{"format":"a","weight":1},"b"]`,     // weight 1 is the default
		`{"format":"{x}","x":"v","repeat":1}`, // repeat 1 is the default
		`[{"format":"A","weight":-1},"B"]`,    // negative weight
		`[{"format":"A","weight":0},{"format":"B","weight":0}]`,         // weights sum to zero
		`[{"format":"A","weight":1e308},{"format":"B","weight":1e308}]`, // weights overflow to +Inf
		`{"format":"A","weight":"heavy"}`,                               // non-numeric weight
		`{"format":"x","repeat":0}`,                                     // repeat below 1
		`{"format":"x","repeat":-2}`,                                    // negative repeat
		`{"format":"x","repeat":1.5}`,                                   // non-integer repeat
		`{"format":"x","repeat":"two"}`,                                 // non-numeric repeat
		`{"format":"x","repeat":2,"separator":5}`,                       // non-string separator
		`{"format":"{x}","x":"v","repeat":2,"separator":""}`,            // separator "" is the default
	} {
		if _, err := compile(parse(t, bad)); err == nil {
			t.Errorf("compile(%s) = nil error, want error", bad)
		}
	}
}
