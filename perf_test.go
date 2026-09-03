package fejkdata

import (
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
)

// The renderer's cost scales with the shape it renders: a deep nest descends one
// level per field, a wide format expands one token per field. The ceiling tests
// below pin the allocations one render costs for those shapes, so a change that
// adds a per-level or per-token allocation (a lost pre-size, a per-item map, an
// extra copy) fails unless the baseline is bumped as a deliberate decision.

// nestedJSON nests a template depth times: each level's format renders its field
// "a", which is the next template down, so one Fake call recurses depth levels.
func nestedJSON(depth int) string {
	s := `"leaf"`
	for i := 0; i < depth; i++ {
		s = fmt.Sprintf(`{"format":"{a}","a":%s}`, s)
	}
	return s
}

// wideTokenJSON is one format with n sibling fields, so one Fake call expands n
// tokens.
func wideTokenJSON(n int) string {
	var toks, fields strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&toks, "{f%d}", i)
		if i > 0 {
			fields.WriteByte(',')
		}
		fmt.Fprintf(&fields, `"f%d":"x"`, i)
	}
	return fmt.Sprintf(`{"format":"%s",%s}`, toks.String(), fields.String())
}

// TestNoRenderAllocRegression fails when a render allocates more than 10% past its
// recorded baseline. Allocations are deterministic across machines, so this gate
// cannot flake under CI load the way a wall-clock ceiling would; a real slowdown
// almost always costs an allocation too. Raising a baseline here is a deliberate
// "we accept this cost" decision.
func TestNoRenderAllocRegression(t *testing.T) {
	shapes := []struct {
		name string
		json string
		base float64
	}{
		{"nested depth 25", nestedJSON(25), 27},
		{"nested depth 100", nestedJSON(100), 102},
		{"wide 500 tokens", wideTokenJSON(500), 9},
	}
	for _, s := range shapes {
		f, err := New(WithoutShippedData(), WithDataFS(fstest.MapFS{"x.json": {Data: []byte(s.json)}}))
		if err != nil {
			t.Fatalf("New(%s): %v", s.name, err)
		}
		allocs := testing.AllocsPerRun(10000, func() { f.Fake("x") })
		if allocs > s.base*1.10 {
			t.Errorf("%s: %.1f allocs/op regressed past %.1f (baseline %.1f + 10%%); bump the baseline only as a deliberate change", s.name, allocs, s.base*1.10, s.base)
		}
	}
}

// A few depth/width benchmarks so the time trend stays visible next to the
// allocation gate.
func BenchmarkNestedDepth25(b *testing.B)  { benchPath(b, tmpData(b, "deep", nestedJSON(25)), "deep") }
func BenchmarkNestedDepth100(b *testing.B) { benchPath(b, tmpData(b, "deep", nestedJSON(100)), "deep") }
func BenchmarkWideTokens100(b *testing.B) {
	benchPath(b, tmpData(b, "wide", wideTokenJSON(100)), "wide")
}
func BenchmarkWideTokens500(b *testing.B) {
	benchPath(b, tmpData(b, "wide", wideTokenJSON(500)), "wide")
}
