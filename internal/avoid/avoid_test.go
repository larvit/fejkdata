package avoid

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

// every lists each value of rs.
func every(rs Runs) []string {
	values := []string{""}
	for _, run := range rs {
		for range run.N {
			var next []string
			for _, v := range values {
				for _, c := range []byte(run.Alphabet) {
					next = append(next, v+string(c))
				}
			}
			values = next
		}
	}
	return values
}

func drawn(t *testing.T, rs Runs, patterns []string, draws int) map[string]bool {
	t.Helper()
	p := Compile(rs, patterns)
	if p == nil {
		t.Fatalf("Compile(%v, %q) = nil, want a plan", rs, patterns)
	}
	r := rand.New(rand.NewPCG(1, 2))
	seen := map[string]bool{}
	for range draws {
		seen[p.Draw(r)] = true
	}
	return seen
}

func bigramsOutsideTheCycles() []string {
	var bigrams []string
	for _, a := range "abcdef" {
		for _, b := range "abcdef" {
			if s := string(a) + string(b); !slices.Contains([]string{"ab", "bc", "ca", "de", "ef", "fd", "da"}, s) {
				bigrams = append(bigrams, s)
			}
		}
	}
	return bigrams
}

// valuesAvoiding lists each value of rs that holds none of the patterns.
func valuesAvoiding(rs Runs, patterns []string) []string {
	var want []string
	for _, v := range every(rs) {
		if !slices.ContainsFunc(patterns, func(p string) bool { return strings.Contains(v, p) }) {
			want = append(want, v)
		}
	}
	return want
}

func drewNoOther(t *testing.T, rs Runs, patterns []string, seen map[string]bool, want []string) {
	t.Helper()
	for v := range seen {
		if !slices.Contains(want, v) {
			t.Errorf("%v avoiding %q drew %q", rs, patterns, v)
		}
	}
}

func drewEvery(t *testing.T, what string, seen map[string]bool, want []string) {
	t.Helper()
	for _, v := range want {
		if !seen[v] {
			t.Errorf("%s never drew %q", what, v)
		}
	}
}

var avoidingCases = []struct {
	name     string
	rs       Runs
	patterns []string
}{
	{"two patterns in six of ab", Runs{{"ab", 6}}, []string{"aa", "bab"}},
	{"four patterns in five of abc", Runs{{"abc", 5}}, []string{"bab", "abc", "cc", "aca"}},
	{"a hyphen between two runs", Runs{{"ab", 2}, {"-", 1}, {"ab", 2}}, []string{"a-a", "b-b"}},
	{"a long run then a single b", Runs{{"ab", 7}, {"b", 1}}, []string{"aa", "bb"}},
	{"a pattern across two runs", Runs{{"7", 1}, {"89ab", 1}}, []string{"79"}},
	{"the tens of two digits", Runs{{"0123456789", 2}}, []string{"10", "11", "12", "13", "14", "15", "16", "17", "18", "19"}},
	{"bigrams outside the cycles", Runs{{"abcdef", 6}, {"b", 1}}, bigramsOutsideTheCycles()},
}

func TestDrawPrintsNoValueHoldingAPattern(t *testing.T) {
	for _, c := range avoidingCases {
		t.Run(c.name, func(t *testing.T) {
			drewNoOther(t, c.rs, c.patterns, drawn(t, c.rs, c.patterns, 20000), valuesAvoiding(c.rs, c.patterns))
		})
	}
}

func TestDrawPrintsEveryValueHoldingNoPattern(t *testing.T) {
	for _, c := range avoidingCases {
		t.Run(c.name, func(t *testing.T) {
			drewEvery(t, fmt.Sprintf("%v avoiding %q", c.rs, c.patterns), drawn(t, c.rs, c.patterns, 20000), valuesAvoiding(c.rs, c.patterns))
		})
	}
}

func TestCompileFindsNoPlanWhereEveryValueHoldsAPattern(t *testing.T) {
	for _, c := range []struct {
		name     string
		rs       Runs
		patterns []string
	}{
		{"every digit", Runs{{"0123456789", 1}}, []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}},
		{"a and bb in three of ab", Runs{{"ab", 3}}, []string{"a", "bb"}},
		{"a hyphen after either letter", Runs{{"ab", 2}, {"-", 1}, {"ab", 2}}, []string{"a-", "b-"}},
		{"a long run with every pair", Runs{{"ab", 1 << 50}, {"a", 1}}, []string{"aa", "bb", "ba"}},
		{"the empty pattern", Runs{{"ab", 2}}, []string{""}},
	} {
		t.Run(c.name, func(t *testing.T) {
			if p := Compile(c.rs, c.patterns); p != nil {
				t.Errorf("Compile(%v, %q) found a plan, want none", c.rs, c.patterns)
			}
		})
	}
}

func TestCompileIgnoresAPatternNoValueCanHold(t *testing.T) {
	rs, patterns := Runs{{"01", 2}}, []string{"x0", "011"}
	drewEvery(t, fmt.Sprintf("%v avoiding %q", rs, patterns), drawn(t, rs, patterns, 2000), every(rs))
}

func TestCompilePlansALongRunWithoutAStepPerCharacter(t *testing.T) {
	for _, c := range []struct {
		name string
		rs   Runs
	}{
		{"a long run", Runs{{"ab", 1 << 50}}},
		{"a long run beside a short one", Runs{{"ab", 1 << 50}, {"b", 1}}},
	} {
		t.Run(c.name, func(t *testing.T) {
			if Compile(c.rs, []string{"aa", "bb"}) == nil {
				t.Errorf("Compile(%v) found no plan, want one", c.rs)
			}
		})
	}
}
