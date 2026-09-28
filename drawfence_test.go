package fejkdata

import (
	"strings"
	"testing"
)

func fenceCorpus(t *testing.T) *Generator {
	t.Helper()
	dir := writeFiles(t, with(geo(), map[string]string{
		"addr.json":  `{"format":"{a} {b} {c}","a":"{/locality.name}, {/municipality.name}","b":{"format":"{/region[12].name} {/place}","drawGroup":"g"},"c":{"format":"{/place.zip} ","repeat":2}}`,
		"place.json": `{"format":"{zip} {name} {tag}","rows":"place.tsv","key":"name"}`,
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

func fenceRoots(t *testing.T, f *Generator) ([]fenceRoot, map[string]*table) {
	t.Helper()
	var roots []fenceRoot
	tables := map[string]*table{}
	_ = walkNodes(f.categories, func(path string, n node) error {
		switch n := n.(type) {
		case *template:
			roots = append(roots, fenceRoot{path, n})
		case *table:
			tables[n.path] = n
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
	return append(roots, fenceRoot{"inline record", record.template}), tables
}

func TestEveryReadARenderMakesIsGathered(t *testing.T) {
	f := fenceCorpus(t)
	roots, tables := fenceRoots(t, f)
	for _, root := range roots {
		wantGathered(t, f.rand, tables, root.label, renderDraws(root.t).reads, func() { renderRoot(f.rand, root.t) })
		if !root.t.isRecord || len(root.t.fields) == 0 {
			continue
		}
		_, columns, err := recordOf(root.t)
		if err != nil {
			t.Fatalf("%s: recordOf = %v", root.label, err)
		}
		wantGathered(t, f.rand, tables, root.label+" as a record", columnDraws(root.t, sortedNames(root.t.fields)).reads, func() { renderRecordRoot(f.rand, root.t, columns) })
	}
}

func wantGathered(t *testing.T, s *session, tables map[string]*table, label string, gathered []pathRead, renderOnce func()) {
	t.Helper()
	var reads []pathRead
	seen := map[readKey]bool{}
	s.trace = &renderTrace{read: func(group, table string, row int, a arm) {
		if s.trace.repeatDepth != 0 || !isRef(a.key) {
			return
		}
		r := pathRead{at: drawAt{group: group}, a: a}
		if table != "" {
			r.at.pins.add(tables[table], row)
		}
		if k := (readKey{group, a.path, r.at.rowsKey()}); !seen[k] {
			seen[k] = true
			reads = append(reads, r)
		}
	}}
	defer func() { s.trace = nil }()
	for i := 0; i < 5; i++ {
		renderOnce()
	}
	for _, r := range reads {
		if !gathers(gathered, r) {
			t.Errorf("%s: the render read {%s} in draw group %q from %s, and drawWalk gathered no such read", label, r.a.spelling, r.at.group, spellPins(&r.at.pins))
		}
	}
}

func gathers(gathered []pathRead, r pathRead) bool {
	for _, g := range gathered {
		if g.at.group == r.at.group && g.a.path == r.a.path && !g.at.pins.differs(&r.at.pins) && !g.at.wholePins.differs(&r.at.pins) {
			return true
		}
	}
	return false
}

func spellPins(p *pinSet) string {
	var rows []string
	p.each(func(t *table, r int) { rows = append(rows, t.selectorSpelling(r)) })
	if len(rows) == 0 {
		return "no row"
	}
	return "row " + strings.Join(rows, ", ")
}

func renderRoot(s *session, t *template) {
	var set holdSet
	sc := renderScope{set: &set}
	switch {
	case t.table != nil:
		render(s, t.table, sc)
		for r := 0; r < t.table.rowCount(); r++ {
			var set holdSet
			sc := renderScope{set: &set}
			sc.hold().pins.pin(t.table, r)
			render(s, t.table.pinnedRow, sc)
		}
	case t.repeat > 1:
		expand(s, t, sc)
	default:
		render(s, t, sc)
	}
}

func renderRecordRoot(s *session, t *template, columns []Column) {
	var set holdSet
	sc := renderScope{set: &set}
	if t.table != nil {
		t.table.drawIn(s, &sc.hold().pins)
	}
	renderRecord(s, t, columns, sc)
}

func TestReplayPairsFindsWhatAWholeReplayFinds(t *testing.T) {
	corpus := fenceCorpus(t)
	roots, _ := fenceRoots(t, corpus)
	shipped, err := New()
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	_ = walkNodes(shipped.categories, func(path string, n node) error {
		if tm, ok := n.(*template); ok {
			roots = append(roots, fenceRoot{path, tm})
		}
		return nil
	})
	for _, root := range roots {
		wantPairsFindAll(t, root.label, renderDraws(root.t).reads)
		if root.t.isRecord && len(root.t.fields) > 0 {
			wantPairsFindAll(t, root.label+" as a record", columnDraws(root.t, sortedNames(root.t.fields)).reads)
		}
	}
}

func TestReplayPairsFindsWhatAWholeReplayOfSelectedRowsFinds(t *testing.T) {
	f, err := New(WithDataPath(writeFiles(t, siblings())))
	if err != nil {
		t.Fatalf("New = %v", err)
	}
	var reads []pathRead
	_ = walkNodes(f.categories, func(_ string, n node) error {
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
			reads = append(reads, renderDraws(root).reads...)
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

func wantPairsFindAll(t *testing.T, label string, reads []pathRead) {
	t.Helper()
	var tabled []pathRead
	for _, r := range reads {
		if r.tr != nil {
			tabled = append(tabled, r)
		}
	}
	pairs := replayPairs(reads)
	var whole error
	eachRenderSet(tabled, func(set []pathRead) {
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
