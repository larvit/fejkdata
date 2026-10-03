package fejkdata

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestEveryReadARenderMakesIsGathered(t *testing.T) {
	f := fenceCorpus(t)
	roots := fenceRoots(t, f)
	traced := 0
	fold := newReadFold()
	for _, root := range roots {
		traced += wantGathered(t, root.label, surveyRender(fold, root.t).reads, func(trace renderTrace) { renderRoot(f.rand, root.t, trace) })
		if !root.t.isRecord || len(root.t.fields) == 0 {
			continue
		}
		_, columns, err := recordOf(root.t)
		if err != nil {
			t.Fatalf("%s: recordOf = %v", root.label, err)
		}
		traced += wantGathered(t, root.label+" as a record", surveyColumns(fold, root.t, sortedNames(root.t.fields)).reads, func(trace renderTrace) { renderRecordRoot(f.rand, root.t, columns, trace) })
	}
	if traced == 0 {
		t.Error("no render traced a reference read, so the comparison proved nothing")
	}
}

func wantGathered(t *testing.T, label string, gathered []pathRead, renderWith func(renderTrace)) int {
	t.Helper()
	var reads []pathRead
	type traceKey struct {
		group, path string
		row         renderedRow
	}
	seen := map[traceKey]bool{}
	trace := func(group string, row renderedRow, a arm) {
		if !isRef(a.head) {
			return
		}
		r := pathRead{group: group, a: a}
		if row.t != nil {
			r.branches.pins.add(row.t, row.index)
		}
		if k := (traceKey{group, a.path, row}); !seen[k] {
			seen[k] = true
			reads = append(reads, r)
		}
	}
	for i := 0; i < 5; i++ {
		renderWith(trace)
	}
	for _, r := range reads {
		if !gathers(gathered, r) {
			t.Errorf("%s: the render read {%s} in draw group %q from %s, and the fold gathered no such read", label, r.a.spelling, r.group, spellPins(&r.branches.pins))
		}
	}
	return len(reads)
}

func gathers(gathered []pathRead, r pathRead) bool {
	for _, g := range gathered {
		if g.group == r.group && g.a.path == r.a.path && !g.branches.pins.differs(&r.branches.pins) && !g.branches.wholePins.differs(&r.branches.pins) {
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

func renderRoot(s *generatorState, t *template, trace renderTrace) {
	draws := renderDraws{trace: trace}
	sc := renderScope{draws: &draws}
	switch {
	case t.site.isFormat():
		tbl := t.site.table
		render(s, tbl, sc)
		for r := 0; r < tbl.rowCount(); r++ {
			draws := renderDraws{trace: trace}
			sc := renderScope{draws: &draws}
			sc.groupDraws().pins.pin(tbl, r)
			render(s, tbl.rowNode, sc.at(tbl.rowNode, &sc.groupDraws().pins))
		}
	case t.repeat > 1:
		expand(s, t, sc)
	default:
		render(s, t, sc)
	}
}

func renderRecordRoot(s *generatorState, t *template, columns []Column, trace renderTrace) {
	draws := renderDraws{trace: trace}
	sc := renderScope{draws: &draws}
	if t.site.isFormat() {
		t.site.table.drawIn(s, &sc.groupDraws().pins)
	}
	renderRecord(s, t, columns, sc)
}

func TestReadFoldWalksACellByEveryRoute(t *testing.T) {
	fsys := fstest.MapFS{
		"g.json": {Data: []byte(`{"format":"{/x.v}","drawGroup":"g2"}`)},
		"r.json": {Data: []byte(`"{/g} {/x}"`)},
		"x.json": {Data: []byte(`{"format":"{v}","rows":"x.tsv","key":"code"}`)},
		"x.tsv":  {Data: []byte("code\tv\n1\t{/z}\n2\t{/z.a}\n")},
		"z.json": {Data: []byte(`{"format":"{a}","a":"hi"}`)},
	}
	g, err := loadDir(dataSource{fsys: fsys}, ".")
	if err != nil {
		t.Fatalf("loadDir = %v", err)
	}
	if err := categoryBinding(categorySites(g), g.children).link(); err != nil {
		t.Fatalf("link = %v", err)
	}
	r, isTemplate := g.children["r"].(*template)
	if !isTemplate {
		t.Fatalf("r is a %T, want a template", g.children["r"])
	}
	var pinned, whole bool
	for _, read := range surveyRender(newReadFold(), r).reads {
		if read.a.path != "/z" {
			continue
		}
		pinned = pinned || read.branches.pins.used > 0
		whole = whole || read.branches.wholePins.used > 0
	}
	if !pinned || !whole {
		t.Errorf("{/z} gathered under a pinned row %v and a whole read's row %v, want both: {/g} reaches x's cells from the root, and {/x} renders them whole", pinned, whole)
	}
}
