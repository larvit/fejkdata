package fejkdata

import (
	"fmt"
	"strings"
)

// tableRead is what a reference path reads of a table family: the table its head
// names, the rows its selectors pin, the tables it draws — those it walks with no
// row pinned, and their unpinned ancestors — each selector's spelling, and whether
// it lands on a row rendered whole.
type tableRead struct {
	head  *table
	pins  pinSet
	drawn map[*table]bool
	sels  []tableSel
	whole bool
}

// tableSel is one selector on the way: the table it selects a row of, and the path
// as written up to and including it.
type tableSel struct {
	t        *table
	spelling string
}

// tableReadOf replays a reference's selectors through the walk a render uses, so
// two paths pinning one row by different routes compare equal. checkPath proved
// each selector names a row.
func tableReadOf(head node, a arm, leaf node) *tableRead {
	t, isTable := head.(*table)
	if !isTable {
		return nil
	}
	tr := &tableRead{head: t, drawn: map[*table]bool{}}
	_, _ = walkPath(t, a.tail, pathWalk{mode: walkProbe, pins: &tr.pins, drawn: tr.drawn})
	written := a.name[:len(a.name)-len(joinSegments(a.tail))]
	cur := t
	for i, seg := range a.tail {
		switch d := cur.descendant(seg); {
		case isSelector(seg):
			tr.sels = append(tr.sels, tableSel{cur, written + joinSegments(a.tail[:i+1])})
		case d != nil:
			cur = d
		}
	}
	_, tr.whole = leaf.(*row)
	return tr
}

// joinSegments spells segments as a path: a selector attaches to the name before it.
func joinSegments(segs []string) string {
	var b strings.Builder
	for _, s := range segs {
		if b.Len() > 0 && !isSelector(s) {
			b.WriteByte('.')
		}
		b.WriteString(s)
	}
	return b.String()
}

// family is the table a chain of parents ends at.
func (t *table) family() *table {
	for t.parentT != nil {
		t = t.parentT
	}
	return t
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

// checkFamilies refuses reads of one table family in one draw group that cannot
// read one consistent draw: a table rendered whole beside a path into the family,
// a table one read draws that another pins, and two reads pinning different rows.
// The reads come sorted by group and path, so which pair is reported does not vary.
// docs/decisions.md#every-reference-path-into-one-family-selects-the-same-rows-per-render-and-group
func checkFamilies(reads []pathRead) error {
	for i, r := range reads {
		if r.tr == nil {
			continue
		}
		for _, o := range reads[:i] {
			if o.tr == nil || o.at.group != r.at.group || alternatives(o.at, r.at) || o.tr.head.family() != r.tr.head.family() {
				continue
			}
			if err := checkFamilyPair(o, r); err != nil {
				return err
			}
		}
	}
	return replayPairs(reads)
}

// replayPairs replays every two reads of one draw group that can render together into
// one draws, the earlier read first. A pin conflicts with one earlier pin, never with a
// combination, so pairs find every conflict a full replay would.
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
	return fmt.Errorf("%s: %w; select the same rows in every path into the family, or draw them apart with a drawGroup", r.at.route.spelled(r.a.name), err)
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

// checkFamilyPair refuses a table read whole beside a path into its family, and a
// table one read draws that the other pins, since which token renders first would
// then decide the row.
// docs/decisions.md#a-table-read-into-is-pinned-a-table-read-whole-draws-afresh
func checkFamilyPair(a, b pathRead) error {
	for _, pair := range [][2]pathRead{{a, b}, {b, a}} {
		x, y := pair[0], pair[1]
		if len(x.a.tail) == 0 && len(y.a.tail) > 0 {
			return overlapError(x.at.route, x.a.name, y)
		}
		if drawn := x.tr.drawnOf(&y.tr.pins); drawn != nil {
			if s, ok := y.tr.selected(x.tr.head); ok && len(x.tr.sels) == 0 {
				tail := x.a.tail
				if s.t != x.tr.head {
					tail = append([]string{x.tr.head.category}, tail...)
				}
				return fmt.Errorf("%s draws %s, which %s selects a row of; write {%s.%s}, or draw them apart with a drawGroup",
					x.at.route.spelled(x.a.name), drawn.category, y.at.route.spelled(y.a.name), s.spelling, joinSegments(tail))
			}
			return fmt.Errorf("%s draws %s, which %s selects a row of; select that row in both, or draw them apart with a drawGroup",
				x.at.route.spelled(x.a.name), drawn.category, y.at.route.spelled(y.a.name))
		}
	}
	return nil
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

// checkOwnFamily refuses a table's format or cell that reads, however many templates
// away and through a repeat or a draw group too, a table of its own family: a row
// rendered whole draws its row without pinning it, so the family would draw apart
// from the row being rendered, whichever draws the reaching template holds.
// docs/decisions.md#a-table-never-reaches-its-own-family-by-any-route
func checkOwnFamily(t *template) error {
	own := t.table
	if own == nil {
		own = t.cellOf
	}
	if own == nil {
		return nil
	}
	seen := map[node]bool{}
	var find func(n node) (renderEdge, *table, bool)
	find = func(n node) (renderEdge, *table, bool) {
		if seen[n] {
			return renderEdge{}, nil, false
		}
		seen[n] = true
		for _, e := range renderEdges(n) {
			if a, isRef := refRead(n, e.label); isRef {
				if head, isTable := n.(*template).head(a.key).(*table); isTable && head.family() == own.family() {
					return e, head, true
				}
			}
			if e, head, found := find(e.to); found {
				return e, head, true
			}
		}
		return renderEdge{}, nil, false
	}
	if e, head, found := find(t); found {
		return fmt.Errorf("%s reads %s, a table of its own family, which a row of %s rendered whole would draw apart from; read the family from a template beside it, or add the value as a column", e.reached(), head.category, own.category)
	}
	return nil
}

// alternatives reports whether two reads sit in different rows of one table: rows the walks pinned,
// or rows of one whole draw.
// docs/decisions.md#the-rows-of-a-table-are-alternatives
func alternatives(a, b drawAt) bool { return a.pinned.differs(&b.pinned) || a.whole.differs(&b.whole) }
