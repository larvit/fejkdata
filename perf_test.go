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
		{"two paths from one held draw", `{"format":"{p.a} {p.b}","p":[{"format":"x","a":"1","b":"2"},{"format":"y","a":"A","b":"B"}]}`, 3},
		{"ten paths from one held draw", `{"format":"{r.a}{r.b}{r.c}{r.d}{r.e}{r.f}{r.g}{r.h}{r.i}{r.j}","r":[
			{"format":"x","a":"1","b":"2","c":"3","d":"4","e":"5","f":"6","g":"7","h":"8","i":"9","j":"0"},
			{"format":"y","a":"A","b":"B","c":"C","d":"D","e":"E","f":"F","g":"G","h":"H","i":"I","j":"J"}]}`, 6},
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

// The repeat shape prices what keeps a hold set off the heap: a copy dropped costs an alloc an iteration.
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
			t.Errorf("%s: %.1f allocs/op regressed past %.1f (baseline %.1f + 10%%); a hold set reaching the heap is the usual cause", s.name, allocs, s.base*1.10, s.base)
		}
	}
}

// A table read pins a row in the render's hold and reads its cells in place.
func TestNoTableAllocRegression(t *testing.T) {
	f, err := New(WithoutShippedData(), WithDataFS(fstest.MapFS{
		"region.json": {Data: []byte(`{"format":"{name}","rows":"region.tsv","key":"code","weight":"population"}`)},
		"region.tsv":  {Data: []byte("code\tname\tpopulation\n01\tStockholms län\t2400000\n12\tSkåne län\t1400000\n14\tVästra Götalands län\t1750000\n")},
		"x.json":      {Data: []byte(`"{/region.name}, {/region.code}"`)},
	}))
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []struct {
		name, path string
		base       float64
	}{
		{"a table drawn", "region", 2},
		{"a row selected", "region[12]", 2},
		{"two columns of one draw", "x", 10},
	} {
		if allocs := testing.AllocsPerRun(10000, func() { f.Fake(s.path) }); allocs > s.base*1.10 {
			t.Errorf("%s: %.1f allocs/op regressed past %.1f (baseline %.1f + 10%%); a row index built per draw is the usual cause", s.name, allocs, s.base*1.10, s.base)
		}
	}
	// Five linked tables, the depth a country's geo tree has: every pin must stay inline.
	deep := fstest.MapFS{"addr.json": {Data: []byte(`"{/e.v} {/d.v} {/c.v} {/b.v} {/a.v}"`)}}
	for i, name := range []string{"a", "b", "c", "d", "e"} {
		rows, category := "k\tv\n1\tx\n2\ty\n", `{"format":"{v}","rows":"`+name+`.tsv","key":"k"}`
		if i > 0 {
			parent := string(rune('a' + i - 1))
			rows = "k\tv\t" + parent + "\n1\tx\t1\n2\ty\t2\n"
			category = `{"format":"{v}","rows":"` + name + `.tsv","key":"k","parent":"` + parent + `"}`
		}
		deep[name+".json"], deep[name+".tsv"] = &fstest.MapFile{Data: []byte(category)}, &fstest.MapFile{Data: []byte(rows)}
	}
	f, err = New(WithoutShippedData(), WithDataFS(deep))
	if err != nil {
		t.Fatal(err)
	}
	const base = 13.0
	if allocs := testing.AllocsPerRun(10000, func() { f.Fake("addr") }); allocs > base*1.10 {
		t.Errorf("five linked tables: %.1f allocs/op regressed past %.1f (baseline %.1f + 10%%); a pin spilling past the inline set is the usual cause", allocs, base*1.10, base)
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
