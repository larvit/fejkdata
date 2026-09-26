package fejkdata

import (
	"strings"
	"testing"
)

// fenceCorpus layers over the shipped set what it lacks: draw groups, a repeat beside one, token
// cells selecting a row, a table read whole and records reading a family.
func fenceCorpus(t *testing.T) *Generator {
	t.Helper()
	dir := writeFiles(t, with(geo(), map[string]string{
		"addr.json":  `{"format":"{a} {b} {c}","a":{"format":"{/locality.name}, {/municipality.name}"},"b":{"format":"{/region[12].name} {/place}","drawGroup":"g"},"c":{"format":"{/place.zip} ","repeat":2}}`,
		"place.json": `{"format":"{zip} {name} {tag}","rows":"place.tsv","key":"name"}`,
		"place.tsv":  "name\tzip\ttag\nStockholm\t1{digits(2)} {digits(2)}\t{/w}\nTranås\t573 {digits(2)}\t{/region[12].name}\n",
		"rec.json":   `{"format":"","l":"{/locality.name}","m":"{/municipality.name}","t":"{/place[Stockholm].tag}"}`,
		"w.json":     `["x","y"]`,
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
	_ = walkNodes(f.categories, func(path string, n node) error {
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

func TestEveryReadARenderMakesIsGathered(t *testing.T) {
	f := fenceCorpus(t)
	for _, root := range fenceRoots(t, f) {
		wantGathered(t, f.rand, root.label, renderDraws(root.t).reads, func() { renderRoot(f.rand, root.t) })
		if !root.t.isRecord {
			continue
		}
		columns := sortedNames(root.t.fields)
		wantGathered(t, f.rand, root.label+" as a record", columnDraws(root.t, columns).reads, func() { renderRecordRoot(f.rand, root.t) })
	}
}

// wantGathered renders a root five times over the seeded stream and fails where a reference the
// render read is not among gathered: same draw group and path, under rows that can render together.
func wantGathered(t *testing.T, s *session, label string, gathered []pathRead, renderOnce func()) {
	t.Helper()
	var reads []pathRead
	s.trace = &renderTrace{read: func(_ *template, group string, pins pinSet, a arm) {
		if s.trace.depth == 0 && isRef(a.key) {
			reads = append(reads, pathRead{at: drawAt{group: group, pins: pins}, a: a})
		}
	}}
	defer func() { s.trace = nil }()
	for i := 0; i < 5; i++ {
		renderOnce()
	}
	for _, r := range reads {
		if !gathers(gathered, r) {
			t.Errorf("%s: the render read {%s} in draw group %q with %s pinned, and drawWalk gathered no such read", label, r.a.spelling, r.at.group, pinned(&r.at.pins))
		}
	}
}

func gathers(gathered []pathRead, r pathRead) bool {
	for _, g := range gathered {
		if g.at.group == r.at.group && g.a.path == r.a.path && !g.at.pins.differs(&r.at.pins) {
			return true
		}
	}
	return false
}

func pinned(p *pinSet) string {
	var rows []string
	p.each(func(t *table, r int) { rows = append(rows, t.selectorSpelling(r)) })
	if len(rows) == 0 {
		return "no row"
	}
	return strings.Join(rows, ", ")
}

// renderRoot renders t the way the render reaches it: a table's format through its table, a cell in
// its row, one iteration of a repeat in t's own draw group, and anything else as itself.
func renderRoot(s *session, t *template) {
	var set holdSet
	sc := renderScope{set: &set}
	switch {
	case t.table != nil:
		render(s, t.table, sc)
	case t.cellOf != nil:
		sc.t, sc.row = t.cellOf, t.cellRow
		render(s, t, sc)
	case t.repeat > 1:
		expand(s, t, sc.in(t))
	default:
		render(s, t, sc)
	}
}

func renderRecordRoot(s *session, t *template) {
	set := eagerHoldSet()
	sc := renderScope{set: &set}
	if t.table != nil {
		t.table.drawIn(s, &sc.in(t).hold().pins)
	}
	_, columns, err := recordOf(t)
	if err != nil {
		panic(err)
	}
	renderRecord(s, t, columns, sc)
}
