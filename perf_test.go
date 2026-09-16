package fejkdata

import (
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
)

// One Fake call descends depth levels: each level's "a" is the next template down.
func nestedJSON(depth int) string {
	s := `"leaf"`
	for i := 0; i < depth; i++ {
		s = fmt.Sprintf(`{"format":"{a}","a":%s}`, s)
	}
	return s
}

// One Fake call expands n sibling tokens.
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

// The repeat shape prices both escape measures: dropping either costs an alloc an iteration.
func TestNoReferenceAllocRegression(t *testing.T) {
	word := `{"format":"{w}","w":["alpha","beta","gamma","delta"]}`
	for _, s := range []struct {
		name, json string
		base       float64
	}{
		{"a repeat of a reference path", `{"format":"{r}","r":{"format":"{/word.w}","repeat":20,"separator":", "}}`, 66},
		{"a named draw group", `{"format":"{a}","a":{"format":"{/word.w}","drawGroup":"g"}}`, 11},
	} {
		f, err := New(WithoutShippedData(), WithDataFS(fstest.MapFS{
			"word.json": {Data: []byte(word)},
			"x.json":    {Data: []byte(s.json)},
		}))
		if err != nil {
			t.Fatalf("New(%s): %v", s.name, err)
		}
		if allocs := testing.AllocsPerRun(10000, func() { f.Fake("x") }); allocs > s.base*1.10 {
			t.Errorf("%s: %.1f allocs/op regressed past %.1f (baseline %.1f + 10%%); a draw set reaching the heap is the usual cause", s.name, allocs, s.base*1.10, s.base)
		}
	}
}

// A record's fences read the compiled tree, so they belong to New, not to a draw.
func TestNoRecordAllocRegression(t *testing.T) {
	for _, s := range []struct{ name, json string }{
		{"record 3 columns", `{"format":"","a":"x","b":"y","c":"z"}`},
		{"record 50 columns", wideTokenJSON(50)},
	} {
		f, err := New(WithoutShippedData(), WithDataFS(fstest.MapFS{"x.json": {Data: []byte(s.json)}}))
		if err != nil {
			t.Fatalf("New(%s): %v", s.name, err)
		}
		if _, err := f.FakeRecord("x"); err != nil {
			t.Fatalf("FakeRecord(%s): %v", s.name, err) // else the gate would measure the error path
		}
		const base = 4.0
		if allocs := testing.AllocsPerRun(10000, func() { f.FakeRecord("x") }); allocs > base*1.10 {
			t.Errorf("%s: %.1f allocs/op regressed past %.1f (baseline %.1f + 10%%); a record fence running per draw is the usual cause", s.name, allocs, base*1.10, base)
		}
		r, _ := f.FakeRecord("x")
		if allocs := testing.AllocsPerRun(10000, func() { _ = r.CSVLine() }); allocs > 2 {
			t.Errorf("%s: CSVLine() makes %.1f allocs/op, want 2: the fields slice and the joined line", s.name, allocs)
		}
	}
}

func TestNoStructAllocRegression(t *testing.T) {
	f, err := New(WithoutShippedData(), WithDataFS(fstest.MapFS{"x.json": {Data: []byte(`{"format":"","a":"x","b":"y","c":"z"}`)}}))
	if err != nil {
		t.Fatal(err)
	}
	var v struct {
		A string `fake:"x.a"`
		B string `fake:"x.b"`
		C string `fake:"x.c"`
	}
	if err := f.FakeStruct(&v); err != nil {
		t.Fatal(err)
	}
	const base = 5.0
	if allocs := testing.AllocsPerRun(10000, func() { f.FakeStruct(&v) }); allocs > base*1.10 {
		t.Errorf("FakeStruct: %.1f allocs/op regressed past %.1f (baseline %.1f + 10%%); compiling the type per call is the usual cause", allocs, base*1.10, base)
	}
}

func BenchmarkNestedDepth25(b *testing.B)  { benchPath(b, tmpData(b, "deep", nestedJSON(25)), "deep") }
func BenchmarkNestedDepth100(b *testing.B) { benchPath(b, tmpData(b, "deep", nestedJSON(100)), "deep") }
func BenchmarkWideTokens100(b *testing.B) {
	benchPath(b, tmpData(b, "wide", wideTokenJSON(100)), "wide")
}
func BenchmarkWideTokens500(b *testing.B) {
	benchPath(b, tmpData(b, "wide", wideTokenJSON(500)), "wide")
}
