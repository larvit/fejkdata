package fejkdata

import "testing"

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

func TestReplayPairsFindsWhatAWholeReplayFinds(t *testing.T) {
	corpus := fenceCorpus(t)
	roots := fenceRoots(t, corpus)
	shipped, err := New()
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
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

func TestReplayPairsFindsWhatAWholeReplayOfSelectedRowsFinds(t *testing.T) {
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
				if replayPairs(triple) != nil {
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
	pairs := replayPairs(reads)
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
		t.Errorf("%s: replayPairs = %v, a whole replay of every read set rendering together = %v", label, pairs, whole)
	}
	return compared
}

// eachRenderSet calls fn with every maximal set of reads that render together, by Bron–Kerbosch.
func eachRenderSet(reads []pathRead, fn func([]pathRead)) {
	together := func(a, b pathRead) bool { return a.at.group == b.at.group && !alternatives(a.at, b.at) }
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
				if together(r, c) {
					nextC = append(nextC, c)
				}
			}
			for _, x := range excluded {
				if together(r, x) {
					nextX = append(nextX, x)
				}
			}
			grow(append(set[:len(set):len(set)], r), nextC, nextX)
			candidates, excluded = candidates[1:], append(excluded, r)
		}
	}
	grow(nil, reads, nil)
}
