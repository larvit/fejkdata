package fejkdata

import (
	"fmt"
	"sort"
)

// readSurvey gathers, at load, the references one render reads, by draw group, for check to compare.
type readSurvey struct {
	reads  []pathRead
	read   map[readKey]bool
	seen   map[nodeVisit]bool
	pinIDs map[pinsLink]pinsID
}

// surveyAt is where a survey stands: the draw group its reads draw in, how the render's root reached it, the
// table rows it pinned, the table a whole read draws a row of, and the rows of whole draws whose
// cells the walk is in.
type surveyAt struct {
	group      string
	route      surveyRoute
	pins       pinSet
	wholeTable *table
	wholePins  pinSet
}

// surveyRoute is how a render reaches a read: as its author spells it, and the root edge's label.
type surveyRoute struct{ spelling, label string }

// pathRead is one reference a render reads: a path, or a bare reference with no tail.
type pathRead struct {
	at surveyAt
	a  arm
	tr *tableRead // set where the reference names a table
}

type nodeVisit struct {
	n          node
	group      string
	wholeTable *table
	pins       pinsID
	wholePins  pinsID
}

type readKey struct {
	group     string
	path      string
	pins      pinsID
	wholePins pinsID
}

// pinsID names a pin set within one walk: two sets pinning the same rows share one.
type pinsID int

// pinsLink is a set's last pin, in path order, and the id of the set before it.
type pinsLink struct {
	prev pinsID
	pin  tablePin
}

func newReadSurvey() *readSurvey {
	return &readSurvey{read: map[readKey]bool{}, seen: map[nodeVisit]bool{}, pinIDs: map[pinsLink]pinsID{}}
}

func surveyRender(t *template) *readSurvey {
	s := newReadSurvey()
	for _, e := range renderEdges(t) {
		s.edge(t, e, surveyAt{group: t.link.drawGroupKey, route: surveyRoute{e.reached(), e.label}})
	}
	return s
}

func surveyColumns(t *template, columns []string) *readSurvey {
	s := newReadSurvey()
	for _, name := range columns {
		s.walk(t.fields[name], surveyAt{group: t.link.drawGroupKey, route: surveyRoute{spelling: fmt.Sprintf("column %q", name)}})
	}
	return s
}

func (s *readSurvey) pinsID(p *pinSet) pinsID {
	var pins []tablePin
	p.each(func(t *table, r int) { pins = append(pins, tablePin{t, r}) })
	sort.Slice(pins, func(i, j int) bool { return pins[i].t.path < pins[j].t.path })
	id := pinsID(0)
	for _, q := range pins {
		l := pinsLink{id, q}
		next, ok := s.pinIDs[l]
		if !ok {
			next = pinsID(len(s.pinIDs) + 1)
			s.pinIDs[l] = next
		}
		id = next
	}
	return id
}

// walk follows what rendering n renders. A repeat renders over draws of its own, so the walk stops
// there.
func (s *readSurvey) walk(n node, at surveyAt) {
	v := nodeVisit{n, at.group, at.wholeTable, s.pinsID(&at.pins), s.pinsID(&at.wholePins)}
	if s.seen[v] {
		return
	}
	s.seen[v] = true
	if repeats(n) {
		return
	}
	if t, isTemplate := n.(*template); isTemplate && t.link.drawGroupKey != "" {
		at.group = t.link.drawGroupKey
	}
	if t, isTable := n.(*table); isTable {
		at.wholeTable = t
	}
	for _, e := range renderEdges(n) {
		to, isTemplate := e.to.(*template)
		cell := tableSite{}
		if isTemplate && to.site.row != formatRow {
			cell = to.site
		}
		switch {
		case cell.table == nil:
			s.edge(n, e, at)
		case cell.table == at.wholeTable:
			in := at
			in.wholePins = at.wholePins.clone()
			in.wholePins.add(cell.table, cell.row)
			s.edge(n, e, in)
		case at.pins.clash(cell.table, cell.row) == nil:
			in := at
			in.pins = at.pins.entered(cell.table, cell.row)
			s.edge(n, e, in)
		default:
			// A pinned table renders its pinned row alone, and an unpinned one draws a row inside its
			// nearest pinned ancestor's, so a cell of a row the pins clash with never renders on this route.
		}
	}
}

// edge records the reference an edge reads, then walks on with every row the read
// pins entered, so a selected row renders only its own cells.
func (s *readSurvey) edge(from node, e renderEdge, at surveyAt) {
	if e.readsRef() {
		a := e.read
		tr := tableReadOf(from.(*template).head(a.key), a, e.to)
		if k := (readKey{at.group, a.path, s.pinsID(&at.pins), s.pinsID(&at.wholePins)}); !s.read[k] {
			s.read[k] = true
			s.reads = append(s.reads, pathRead{at, a, tr})
		}
		if tr != nil {
			tr.pins.each(func(t *table, r int) { at.pins = at.pins.entered(t, r) })
		}
	}
	s.walk(e.to, at)
}

// spelled names the route, and the reference it reaches a read by where its root edge is not that
// reference.
func (r surveyRoute) spelled(ref string) string {
	if ref == "" || ref == r.label {
		return r.spelling
	}
	return fmt.Sprintf("%s with {%s}", r.spelling, ref)
}

// tableRead is what a reference path reads of a table family: the table its head
// names, the rows its selectors pin, the tables it draws — those it walks with no
// row pinned, and their unpinned ancestors — each selector's spelling, and whether
// it lands on a `tableRow`.
type tableRead struct {
	headTable *table
	pins      pinSet
	drawn     map[*table]bool
	sels      []tableSel
	landsRow  bool
}

// tableSel is one selector on the way: the table it selects a row of, and the path
// as written up to and including it.
type tableSel struct {
	t        *table
	spelling string
}

// tableReadOf replays a reference path through pathProbe, so two paths pinning one
// row by different routes compare equal. checkPath proved
// each selector names a row.
func tableReadOf(head node, a arm, leaf node) *tableRead {
	t, isTable := head.(*table)
	if !isTable {
		return nil
	}
	probe := &pathProbe{drawn: map[*table]bool{}}
	_, _ = probe.walk(t, a.tail)
	tr := &tableRead{headTable: t, pins: probe.pins, drawn: probe.drawn}
	written := a.spelling[:len(a.spelling)-len(joinSegments(a.tail))]
	cur := t
	for i, seg := range a.tail {
		switch d := cur.descendant(seg); {
		case isSelector(seg):
			tr.sels = append(tr.sels, tableSel{cur, written + joinSegments(a.tail[:i+1])})
		case d != nil:
			cur = d
		}
	}
	_, tr.landsRow = leaf.(*tableRow)
	return tr
}

// selected is the selector in r on t, or on the nearest ancestor of t it selects.
func (r *tableRead) selected(t *table) (tableSel, bool) {
	for ; t != nil; t = t.parentT {
		for _, s := range r.sels {
			if s.t == t {
				return s, true
			}
		}
	}
	return tableSel{}, false
}

// drawnOf is a table the read draws that pins holds a row of, if any.
func (r *tableRead) drawnOf(pins *pinSet) *table {
	var found *table
	pins.each(func(t *table, _ int) {
		if found == nil && r.drawn[t] {
			found = t
		}
	})
	return found
}

// replay pins the read's rows into d, where they agree with the rows pinned before.
func (r *tableRead) replay(d *pinSet) error {
	var err error
	r.pins.each(func(t *table, row int) {
		if err == nil {
			err = d.pinRow(t, row)
		}
	})
	return err
}
