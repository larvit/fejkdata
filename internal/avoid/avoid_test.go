package avoid

import (
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

func TestDrawPrintsEveryValueHoldingNoPatternAndNoOther(t *testing.T) {
	tens := []string{"10", "11", "12", "13", "14", "15", "16", "17", "18", "19"}
	var bigrams []string
	for _, a := range "abcdef" {
		for _, b := range "abcdef" {
			if s := string(a) + string(b); !slices.Contains([]string{"ab", "bc", "ca", "de", "ef", "fd", "da"}, s) {
				bigrams = append(bigrams, s)
			}
		}
	}
	for _, c := range []struct {
		rs       Runs
		patterns []string
	}{
		{Runs{{"ab", 6}}, []string{"aa", "bab"}},
		{Runs{{"abc", 5}}, []string{"bab", "abc", "cc", "aca"}},
		{Runs{{"ab", 2}, {"-", 1}, {"ab", 2}}, []string{"a-a", "b-b"}},
		{Runs{{"ab", 7}, {"b", 1}}, []string{"aa", "bb"}},
		{Runs{{"7", 1}, {"89ab", 1}}, []string{"79"}},
		{Runs{{"0123456789", 2}}, tens},
		{Runs{{"abcdef", 6}, {"b", 1}}, bigrams},
	} {
		var want []string
		for _, v := range every(c.rs) {
			if !slices.ContainsFunc(c.patterns, func(p string) bool { return strings.Contains(v, p) }) {
				want = append(want, v)
			}
		}
		seen := drawn(t, c.rs, c.patterns, 20000)
		for v := range seen {
			if !slices.Contains(want, v) {
				t.Errorf("%v avoiding %q drew %q", c.rs, c.patterns, v)
			}
		}
		for _, v := range want {
			if !seen[v] {
				t.Errorf("%v avoiding %q never drew %q", c.rs, c.patterns, v)
			}
		}
	}
}

func TestCompileFindsNoPlanWhereEveryValueHoldsAPattern(t *testing.T) {
	for _, c := range []struct {
		rs       Runs
		patterns []string
	}{
		{Runs{{"0123456789", 1}}, []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}},
		{Runs{{"ab", 3}}, []string{"a", "bb"}},
		{Runs{{"ab", 2}, {"-", 1}, {"ab", 2}}, []string{"a-", "b-"}},
		{Runs{{"ab", 1 << 50}, {"a", 1}}, []string{"aa", "bb", "ba"}},
		{Runs{{"ab", 2}}, []string{""}},
	} {
		if p := Compile(c.rs, c.patterns); p != nil {
			t.Errorf("Compile(%v, %q) found a plan, want none", c.rs, c.patterns)
		}
	}
}

func TestCompileIgnoresAPatternNoValueCanHold(t *testing.T) {
	seen := drawn(t, Runs{{"01", 2}}, []string{"x0", "011"}, 2000)
	for _, v := range every(Runs{{"01", 2}}) {
		if !seen[v] {
			t.Errorf("never drew %q", v)
		}
	}
}

func TestCompilePlansALongRunWithoutAStepPerCharacter(t *testing.T) {
	for _, rs := range []Runs{{{"ab", 1 << 50}}, {{"ab", 1 << 50}, {"b", 1}}} {
		if Compile(rs, []string{"aa", "bb"}) == nil {
			t.Errorf("Compile(%v) found no plan, want one", rs)
		}
	}
}
