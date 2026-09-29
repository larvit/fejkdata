package fejkdata

import (
	"fmt"
	"sort"
	"strings"
)

// readSurvey gathers, at load, the references one render reads, by draw group, for check to compare.
type readSurvey struct {
	reads []pathRead
	read  map[readKey]bool
	seen  map[nodeVisit]bool
	clash error
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
	pins       string
	wholePins  string
}

type readKey struct {
	group     string
	path      string
	pins      string
	wholePins string
}

func newReadSurvey() *readSurvey {
	return &readSurvey{read: map[readKey]bool{}, seen: map[nodeVisit]bool{}}
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

// walk follows what rendering n renders. A repeat renders over draws of its own, so the walk stops
// there.
func (s *readSurvey) walk(n node, at surveyAt) {
	v := nodeVisit{n, at.group, at.wholeTable, at.pins.key(), at.wholePins.key()}
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
		var cell tableSite
		if isTemplate {
			cell = to.site
		}
		switch {
		case !cell.isCell():
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
		tr := tableReadOf(from.(*template).head(a.head), a, e.to)
		if k := (readKey{at.group, a.path, at.pins.key(), at.wholePins.key()}); !s.read[k] {
			s.read[k] = true
			s.reads = append(s.reads, pathRead{at, a, tr})
		}
		if tr != nil {
			pins := at.pins.clone()
			if err := tr.replay(&pins); err != nil {
				if s.clash == nil {
					s.clash = fmt.Errorf("%s: %w; read the family from a template beside it, or add the value as a column", at.route.spelled(a.spelling), err)
				}
				return
			}
			at.pins = pins
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

// check refuses what one draw per reference path cannot answer for: a read selecting a row its route
// pinned another of, a read of a level beside a path another read takes into it, and two reads of one
// table family that select different rows. Reads are compared in path order, so which pair is
// reported does not vary.
func (s *readSurvey) check() error {
	if s.clash != nil {
		return s.clash
	}
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

// checkOwnFamily refuses a table's format or cell that reads, however many templates
// away and through a repeat or a draw group too, a table of its own family: a whole
// read draws its row without pinning it, so the family would draw apart from the row
// it renders, whichever draws the reaching template holds.
// docs/decisions.md#a-table-never-reaches-its-own-family-by-any-route
func checkOwnFamily(path string, n node) error {
	t, isTemplate := n.(*template)
	if !isTemplate || t.site.table == nil {
		return nil
	}
	own := t.site.table
	seen := map[node]bool{}
	var find func(n node) (renderEdge, *table, bool)
	find = func(n node) (renderEdge, *table, bool) {
		if seen[n] {
			return renderEdge{}, nil, false
		}
		seen[n] = true
		for _, e := range renderEdges(n) {
			if e.readsRef() {
				if familyTable, isTable := n.(*template).head(e.read.head).(*table); isTable && familyTable.familyRoot() == own.familyRoot() {
					return e, familyTable, true
				}
			}
			if e, familyTable, found := find(e.to); found {
				return e, familyTable, true
			}
		}
		return renderEdge{}, nil, false
	}
	if e, familyTable, found := find(t); found {
		return fmt.Errorf("%s: %s reads %s, a table of its own family, which a bare read of %s would draw apart from the row it renders; read the family from a template beside it, or add the value as a column", path, e.reached(), familyTable.segment, own.segment)
	}
	return nil
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

// replayPairs replays, of every two reads that can render together, the later into the
// earlier's pins. Pairs find every conflict a full replay would: clash judges a row against one
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
			d := r.tr.pins.clone()
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
	apart := ""
	switch {
	case route.spelling == into.at.route.spelling:
	case isRef(into.at.route.label):
		apart = fmt.Sprintf(", or move {%s} into a field with a drawGroup", into.at.route.label)
	default:
		apart = fmt.Sprintf(", or give %s a drawGroup", into.at.route.spelling)
	}
	return fmt.Errorf("%s renders its own draw of what %s reads a path through; name the fields you want instead%s", route.spelled(ref), into.at.route.spelled(into.a.spelling), apart)
}
