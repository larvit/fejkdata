package fejkdata

import (
	"fmt"
	"sort"
	"strings"
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
	probe := &pathProbe{a: &a}
	_, _ = probe.walk(t, a.tail)
	_, landsRow := leaf.(*tableRow)
	return &tableRead{headTable: t, pins: probe.pins, drawn: probe.drawn, sels: probe.sels, landsRow: landsRow}
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

// check refuses what one draw per reference path cannot answer for: a read of a level beside a path
// another read takes into it, and two reads of one table family that select different rows. Reads
// are compared in path order, so which pair is reported does not vary.
func (s *readSurvey) check() error {
	sort.SliceStable(s.reads, func(i, j int) bool {
		if s.reads[i].at.group != s.reads[j].at.group {
			return s.reads[i].at.group < s.reads[j].at.group
		}
		return s.reads[i].a.path < s.reads[j].a.path
	})
	for i, level := range s.reads {
		for _, into := range s.reads[i+1:] {
			if into.at.group != level.at.group || alternatives(level.at, into.at) {
				continue
			}
			if strings.HasPrefix(into.a.path, level.a.path+".") && !(level.tr != nil && level.tr.landsRow) {
				return overlapError(level.at.route, level.a.spelling, into)
			}
		}
	}
	return checkFamilies(s.reads)
}

// checkFamilies refuses reads of one table family in one draw group that cannot
// read one consistent draw: a table rendered whole beside a path into the family,
// a table one read draws that another pins, and two reads pinning different rows.
// The reads come sorted by group and path, so which pair is reported does not vary.
func checkFamilies(reads []pathRead) error {
	for i, r := range reads {
		if r.tr == nil {
			continue
		}
		for _, o := range reads[:i] {
			if o.tr == nil || o.at.group != r.at.group || alternatives(o.at, r.at) || o.tr.headTable.familyRoot() != r.tr.headTable.familyRoot() {
				continue
			}
			if err := checkFamilyPair(o, r); err != nil {
				return err
			}
		}
	}
	return replayPairs(reads)
}

// checkFamilyPair refuses a table read whole beside a path into its family, and a
// table one read draws that the other pins, since which token renders first would
// then decide the row.
// docs/decisions.md#a-bare-reference-draws-each-time-a-reference-path-is-held
func checkFamilyPair(a, b pathRead) error {
	for _, pair := range [][2]pathRead{{a, b}, {b, a}} {
		x, y := pair[0], pair[1]
		if len(x.a.tail) == 0 && len(y.a.tail) > 0 {
			return overlapError(x.at.route, x.a.spelling, y)
		}
		if drawn := x.tr.drawnOf(&y.tr.pins); drawn != nil {
			if s, ok := y.tr.selected(x.tr.headTable); ok && len(x.tr.sels) == 0 {
				tail := x.a.tail
				if s.t != x.tr.headTable {
					tail = append([]string{x.tr.headTable.segment}, tail...)
				}
				return fmt.Errorf("%s draws %s, which %s selects a row of; write {%s.%s}, or draw them apart with a drawGroup",
					x.at.route.spelled(x.a.spelling), drawn.segment, y.at.route.spelled(y.a.spelling), s.spelling, joinSegments(tail))
			}
			return fmt.Errorf("%s draws %s, which %s selects a row of; select that row in both, or draw them apart with a drawGroup",
				x.at.route.spelled(x.a.spelling), drawn.segment, y.at.route.spelled(y.a.spelling))
		}
	}
	return nil
}

// replayPairs replays every two reads that can render together into one draws, the earlier
// read first. Pairs find every conflict a full replay would: clash judges a row against one
// pinned table, and the read that pinned it holds that pin itself, since a read's pins carry
// its rows' ancestors, so the pair of those two reads clashes the same way.
func replayPairs(reads []pathRead) error {
	for i, r := range reads {
		if r.tr == nil {
			continue
		}
		for _, o := range reads[i+1:] {
			if o.tr == nil || o.at.group != r.at.group || alternatives(r.at, o.at) {
				continue
			}
			var d pinSet
			if err := r.tr.replay(&d); err != nil {
				return conflict(r, err)
			}
			if err := o.tr.replay(&d); err != nil {
				return conflict(o, err)
			}
		}
	}
	return nil
}

func conflict(r pathRead, err error) error {
	return fmt.Errorf("%s: %w; select the same rows in every path into the family, or draw them apart with a drawGroup", r.at.route.spelled(r.a.spelling), err)
}

// alternatives reports whether two reads never render together: one read renders one row of a
// table, so reads under different rows the walks pinned, or under different rows of one whole
// draw, never meet.
func alternatives(a, b surveyAt) bool {
	return a.pins.differs(&b.pins) || a.wholePins.differs(&b.wholePins)
}

func overlapError(route surveyRoute, ref string, into pathRead) error {
	return fmt.Errorf("%s renders a level that %s reads a path into; name the fields you want instead, or draw them apart with a drawGroup", route.spelled(ref), into.at.route.spelled(into.a.spelling))
}
