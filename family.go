package fejkdata

import (
	"fmt"
	"sort"
	"strings"
)

// tablePin is one table's pinned row.
type tablePin struct {
	t   *table
	row int
}

// pinned is the row the render pinned for t, if any.
func (d *draws) pinned(t *table) (int, bool) {
	for _, p := range d.pins[:d.npins] {
		if p.t == t {
			return p.row, true
		}
	}
	r, ok := d.more[t]
	return r, ok
}

// mustRow is the row pinned for t, which the walk reaching a column pinned.
func (d *draws) mustRow(t *table) int {
	r, ok := d.pinned(t)
	if !ok {
		panic(fmt.Sprintf("fejkdata: a column of %s is rendered with no row pinned", t.category))
	}
	return r
}

// pin pins row r of t, and the rows of t's ancestors it links to.
func (d *draws) pin(t *table, r int) {
	for {
		if _, done := d.pinned(t); done {
			return
		}
		if d.npins < len(d.pins) {
			d.pins[d.npins] = tablePin{t, r}
			d.npins++
		} else {
			if d.more == nil {
				d.more = map[*table]int{}
			}
			d.more[t] = r
		}
		if t.parentT == nil {
			return
		}
		t, r = t.parentT, t.parentRow(r)
	}
}

// each calls fn for every pinned row, in pin order, the spilled ones by table name.
func (d *draws) each(fn func(t *table, r int)) {
	for _, p := range d.pins[:d.npins] {
		fn(p.t, p.row)
	}
	spilled := make([]*table, 0, len(d.more))
	for t := range d.more {
		spilled = append(spilled, t)
	}
	sort.Slice(spilled, func(i, j int) bool { return spilled[i].category < spilled[j].category })
	for _, t := range spilled {
		fn(t, d.more[t])
	}
}

// rowOf is the render's row of t: the one pinned, else one drawn inside the
// nearest pinned ancestor — its parent drawn inside that first where the ancestor
// is further up — or over the whole table, and pinned with its ancestors.
func (d *draws) rowOf(t *table) int {
	if r, ok := d.pinned(t); ok {
		return r
	}
	r := -1
	for a := t.parentT; a != nil && r < 0; a = a.parentT {
		if _, ok := d.pinned(a); !ok {
			continue
		}
		if t.parentT != a {
			d.rowOf(t.parentT)
		}
		pr, _ := d.pinned(t.parentT)
		r = t.drawUnder(d.s, pr)
	}
	if r < 0 {
		r = t.draw(d.s)
	}
	d.pin(t, r)
	return r
}

// pinRow pins row r of t where it agrees with the rows pinned before it.
func (d *draws) pinRow(t *table, r int) error {
	if pr, ok := d.pinned(t); ok && pr != r {
		return fmt.Errorf("%s and %s are two rows of %s", t.selectorSpelling(pr), t.selectorSpelling(r), t.category)
	}
	for a := t.parentT; a != nil; a = a.parentT {
		if pa, ok := d.pinned(a); ok && !t.under(r, a, pa) {
			return fmt.Errorf("%s is not inside %s", t.selectorSpelling(r), a.selectorSpelling(pa))
		}
	}
	d.pin(t, r)
	return nil
}

// selectRow pins the row a selector names.
func (d *draws) selectRow(t *table, sel string) error {
	r, err := t.find(sel, d)
	if err != nil {
		return err
	}
	return d.pinRow(t, r)
}

// inside keeps the rows of t that sit inside every pinned ancestor.
func (d *draws) inside(t *table, rows []int) []int {
	for a := t.parentT; a != nil; a = a.parentT {
		pa, ok := d.pinned(a)
		if !ok {
			continue
		}
		var kept []int
		for _, r := range rows {
			if t.under(r, a, pa) {
				kept = append(kept, r)
			}
		}
		rows = kept
	}
	return rows
}

// tableRead is what a reference path reads of a table family: the table its head
// names, the rows its selectors pin, the tables it draws — those it walks with no
// row pinned, and their unpinned ancestors — each selector's spelling, and whether
// it lands on a row rendered whole.
type tableRead struct {
	head  *table
	pins  draws
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
	_, _ = walkPath(t, a.tail, pathWalk{pins: &tr.pins})
	written := a.name[:len(a.name)-len(joinSegments(a.tail))]
	cur := t
	tr.draws(cur)
	for i, seg := range a.tail {
		switch d := cur.descendant(seg); {
		case isSelector(seg):
			tr.sels = append(tr.sels, tableSel{cur, written + joinSegments(a.tail[:i+1])})
		case d != nil:
			cur = d
			tr.draws(cur)
		}
	}
	_, tr.whole = leaf.(*row)
	return tr
}

// draws marks t and its ancestors drawn, up to the first the read pins.
func (r *tableRead) draws(t *table) {
	for ; t != nil; t = t.parentT {
		if _, pinned := r.pins.pinned(t); pinned {
			return
		}
		r.drawn[t] = true
	}
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
			var d draws
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
func (r *tableRead) replay(d *draws) error {
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
func (r *tableRead) drawnOf(pins *draws) *table {
	var found *table
	pins.each(func(t *table, _ int) {
		if found == nil && r.drawn[t] {
			found = t
		}
	})
	return found
}
