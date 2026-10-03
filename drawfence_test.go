package fejkdata

import (
	"strings"
	"testing"
	"testing/fstest"
)

func fenceCorpus(t *testing.T) *Generator {
	t.Helper()
	dir := writeFiles(t, with(geo(), map[string]string{
		"addr.json":  `{"format":"{a} {b} {c}","a":"{/locality.name}, {/municipality.name}","b":{"format":"{/region[12].name} {/place}","drawGroup":"g"},"c":{"format":"{/place.zip} ","repeat":2}}`,
		"place.json": `{"format":"{zip} {name} {tag} {/w}","rows":"place.tsv","key":"name"}`,
		"place.tsv":  "name\tzip\ttag\nStockholm\t1{digits(2)} {digits(2)}\t{/x}\nTranås\t573 {digits(2)}\t{/region[12].name}\n",
		"rec.json":   `{"format":"","l":"{/locality.name}","m":"{/municipality.name}","t":"{/place[Stockholm].tag}"}`,
		"w.json":     `["x","y"]`,
		"x.json":     `"{/w}"`,
	}))
	f, err := New(WithDataPath(dir), WithSeed(1))
	if err != nil {
		t.Fatalf("New = %v", err)
	}
	return f
}

type fenceRoot struct {
	label string
	t     *template
}

func fenceRoots(t *testing.T, f *Generator) []fenceRoot {
	t.Helper()
	var roots []fenceRoot
	_ = walkNodes(f.root.children, func(path string, n node) error {
		if tm, ok := n.(*template); ok {
			roots = append(roots, fenceRoot{path, tm})
		}
		return nil
	})
	inline, err := f.NewTemplate(`{"format":"{x}|{y}","x":{"format":"{/locality.name}","drawGroup":"g"},"y":"{/region[01].name}"}`)
	if err != nil {
		t.Fatalf("NewTemplate = %v", err)
	}
	record, err := f.NewRecordTemplate(`{"format":"","a":"{/locality.name}","b":"{/municipality.name}"}`)
	if err != nil {
		t.Fatalf("NewRecordTemplate = %v", err)
	}
	_ = eachNode(inline.n, "inline", func(path string, n node) error {
		if tm, ok := n.(*template); ok {
			roots = append(roots, fenceRoot{path, tm})
		}
		return nil
	})
	return append(roots, fenceRoot{"inline record", record.template})
}

func TestPairsFindWhatAWholeReplayFinds(t *testing.T) {
	corpus := fenceCorpus(t)
	roots := fenceRoots(t, corpus)
	shipped := newShippedWhole(t)
	_ = walkNodes(shipped.root.children, func(path string, n node) error {
		if tm, ok := n.(*template); ok {
			roots = append(roots, fenceRoot{path, tm})
		}
		return nil
	})
	compared := 0
	fold := newReadFold()
	for _, root := range roots {
		compared += wantPairsFindAll(t, root.label, surveyRender(fold, root.t).reads)
		if root.t.isRecord && len(root.t.fields) > 0 {
			compared += wantPairsFindAll(t, root.label+" as a record", surveyColumns(fold, root.t, sortedNames(root.t.fields)).reads)
		}
	}
	if compared == 0 {
		t.Error("no read set rendering together holds two table reads, so the comparison proved nothing")
	}
}

func TestPairsFindWhatAWholeReplayOfSelectedRowsFinds(t *testing.T) {
	f, err := New(WithDataPath(writeFiles(t, siblings())))
	if err != nil {
		t.Fatalf("New = %v", err)
	}
	var reads []pathRead
	fold := newReadFold()
	_ = walkNodes(f.root.children, func(_ string, n node) error {
		tb, isTable := n.(*table)
		if !isTable || tb.keyIndex < 0 || tb.familyRoot().path != "region" {
			return nil
		}
		for r := 0; r < tb.rowCount(); r++ {
			inline, err := f.NewTemplate(`"{/` + tb.selectorSpelling(r) + `}"`)
			if err != nil {
				t.Fatalf("NewTemplate(%s) = %v", tb.selectorSpelling(r), err)
			}
			root, isTemplate := inline.n.(*template)
			if !isTemplate {
				t.Fatalf("NewTemplate(%s) compiled to %T", tb.selectorSpelling(r), inline.n)
			}
			reads = append(reads, surveyRender(fold, root).reads...)
		}
		return nil
	})
	conflicts := 0
	for _, a := range reads {
		for _, b := range reads {
			for _, c := range reads {
				triple := []pathRead{a, b, c}
				wantPairsFindAll(t, a.a.spelling+", "+b.a.spelling+", "+c.a.spelling, triple)
				if (&readSurvey{reads: triple}).check() != nil {
					conflicts++
				}
			}
		}
	}
	if conflicts == 0 {
		t.Error("no triple of selected rows conflicts, so the comparison proved nothing")
	}
}

// wantPairsFindAll counts the read sets rendering together that hold two table reads or more.
func wantPairsFindAll(t *testing.T, label string, reads []pathRead) int {
	t.Helper()
	var tabled []pathRead
	for _, r := range reads {
		if r.tr != nil {
			tabled = append(tabled, r)
		}
	}
	pairs := (&readSurvey{reads: reads}).check()
	var whole error
	compared := 0
	eachRenderSet(tabled, func(set []pathRead) {
		if len(set) > 1 {
			compared++
		}
		var d pinSet
		for _, r := range set {
			if err := r.tr.replay(&d); err != nil && whole == nil {
				whole = err
			}
		}
	})
	if (pairs == nil) != (whole == nil) {
		t.Errorf("%s: check = %v, a whole replay of every read set rendering together = %v", label, pairs, whole)
	}
	return compared
}

// eachRenderSet calls fn with every maximal set of reads that render together, by Bron–Kerbosch.
func eachRenderSet(reads []pathRead, fn func([]pathRead)) {
	var grow func(set, candidates, excluded []pathRead)
	grow = func(set, candidates, excluded []pathRead) {
		if len(candidates) == 0 && len(excluded) == 0 {
			fn(set)
			return
		}
		for len(candidates) > 0 {
			r := candidates[0]
			var nextC, nextX []pathRead
			for _, c := range candidates[1:] {
				if coRender(r, c) {
					nextC = append(nextC, c)
				}
			}
			for _, x := range excluded {
				if coRender(r, x) {
					nextX = append(nextX, x)
				}
			}
			grow(append(set[:len(set):len(set)], r), nextC, nextX)
			candidates, excluded = candidates[1:], append(excluded, r)
		}
	}
	grow(nil, reads, nil)
}

func TestReadSurveyReportsAPinClashAsAnError(t *testing.T) {
	fsys := fstest.MapFS{
		"r.json": {Data: []byte(`"{/y[a].v}"`)},
		"m.json": {Data: []byte(`{"format":"{g}{f}","f":"{/y[a].v}","g":"{/y[a].v}"}`)},
		"n.json": {Data: []byte(`{"format":"{f}","f":"{/y[a].v}"}`)},
		"x.json": {Data: []byte(`{"format":"{v}","rows":"x.tsv","key":"code"}`)},
		"x.tsv":  {Data: []byte("code\tv\n1\tb\n2\tc\n")},
		"y.json": {Data: []byte(`{"format":"{v}","rows":"y.tsv","key":"code","parent":"x"}`)},
		"y.tsv":  {Data: []byte("code\tv\tx\na\t{/x[2].v}\t1\nb\tz\t2\n")},
	}
	g, err := loadDir(dataSource{fsys: fsys}, ".")
	if err != nil {
		t.Fatalf("loadDir = %v", err)
	}
	if err := treeBinding(g.children).link(); err != nil {
		t.Fatalf("link = %v", err)
	}
	r, isTemplate := g.children["r"].(*template)
	if !isTemplate {
		t.Fatalf("r is a %T, want a template", g.children["r"])
	}
	if err := surveyRender(newReadFold(), r).check(); err == nil || !strings.Contains(err.Error(), "two rows of") {
		t.Errorf("check = %v, want y[a]'s row of x clashing with x[2] named, whatever order the fences run in", err)
	}
	n, isTemplate := g.children["n"].(*template)
	if !isTemplate {
		t.Fatalf("n is a %T, want a template", g.children["n"])
	}
	if err := surveyRender(newReadFold(), n).check(); err == nil || !strings.HasPrefix(err.Error(), "{f} with {/x[2].v}: x[1] and x[2] are two rows of x") {
		t.Errorf("check = %v, want the clash below {f} named by the route the root reaches it by", err)
	}
	m, isTemplate := g.children["m"].(*template)
	if !isTemplate {
		t.Fatalf("m is a %T, want a template", g.children["m"])
	}
	if err := surveyRender(newReadFold(), m).check(); err == nil || !strings.HasPrefix(err.Error(), "{f} with") {
		t.Errorf("check = %v, want the clash of the lowest route named, {f}, whichever the walk reaches first", err)
	}
}

func surveyRender(f *readFold, t *template) *readSurvey {
	return survey(f.rootReads(t), t.link.drawGroupKey)
}

func surveyColumns(f *readFold, t *template, columns []string) *readSurvey {
	return survey(f.columnReads(t, columns), t.link.drawGroupKey)
}
