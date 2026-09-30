package fejkdata

import "fmt"

// readRoute is how a render reaches a read: as its author spells it, and the root edge's label.
type readRoute struct{ spelling, label string }

// branches is the rows a read sits under, where two reads under different rows of one table never
// render together. pins holds the rows pinned above it with their ancestors, which a pair of reads
// needs to find what a replay of all of them does; wholePins the rows of whole draws whose cells it
// sits in, without ancestors, since each whole draw draws apart from every other.
type branches struct {
	pins      pinSet
	wholePins pinSet
}

// pathRead is one reference a render reads, a path or a bare reference with no tail: in its draw
// group ("" until the root names one), by the route the render's root reached it, under its
// branches.
type pathRead struct {
	group    string
	route    readRoute
	branches branches
	a        arm
	tr       *tableRead // set where the reference names a table
	clash    error      // set where its selectors clash with a row its route pinned
}

type readKey struct {
	group     string
	path      string
	pins      string
	wholePins string
	clash     bool
}

// readFold gathers, once per node, the reference reads rendering it makes.
type readFold struct {
	memo map[node][]pathRead
}

func newReadFold() *readFold { return &readFold{memo: map[node][]pathRead{}} }

// rootReads is what rendering t as a render of its own reads, each read named by the edge of t
// that reaches it, and in "" where no template below t names a draw group.
func (f *readFold) rootReads(t *template) []pathRead {
	var reads []pathRead
	for _, e := range renderEdges(t) {
		reads = append(reads, f.viaEdge(t, e, readRoute{e.reached(), e.label})...)
	}
	return reads
}

// columnReads is what rendering t's columns as one record reads.
func (f *readFold) columnReads(t *template, columns []string) []pathRead {
	var reads []pathRead
	for _, name := range columns {
		reads = append(reads, routed(f.reads(t.fields[name]), readRoute{spelling: fmt.Sprintf("column %q", name)})...)
	}
	return reads
}

// reads is what rendering n reads, short of a repeat, which renders over draws of its own: each
// read with the rows pinned above it inside n, and in its draw group where a template inside n
// names one.
func (f *readFold) reads(n node) []pathRead {
	if r, done := f.memo[n]; done {
		return r
	}
	var out []pathRead
	switch n := n.(type) {
	case *choice:
		for _, it := range n.items {
			out = append(out, f.reads(it)...)
		}
	case *template:
		if repeats(n) {
			break
		}
		for _, e := range renderEdges(n) {
			out = append(out, f.viaEdge(n, e, readRoute{})...)
		}
		if n.link.drawGroupKey != "" {
			for i := range out {
				if out[i].group == "" {
					out[i].group = n.link.drawGroupKey
				}
			}
		}
	case *table:
		out = f.rowReads(n, func(b *branches, row int) bool {
			b.wholePins = b.wholePins.clone()
			b.wholePins.add(n, row)
			return true
		})
	case *tableRow:
		out = f.rowReads(n.t, func(b *branches, row int) bool { return b.enter(n.t, row) })
	case *tableColumn:
		out = f.cellReads(n, func(b *branches, row int) bool { return b.enter(n.t, row) })
	case *folder, *nullItem:
	default:
		panic(internalError("reads has no case for node %T", n))
	}
	out = distinct(out)
	f.memo[n] = out
	return out
}

// distinct keeps one of every read gathered by several routes, so a diamond of templates
// contributes each read once however many routes reach it.
func distinct(reads []pathRead) []pathRead {
	seen := make(map[readKey]bool, len(reads))
	out := reads[:0:0]
	for _, r := range reads {
		k := readKey{r.group, r.a.path, r.branches.pins.key(), r.branches.wholePins.key(), r.clash != nil}
		if !seen[k] {
			seen[k] = true
			out = append(out, r)
		}
	}
	return out
}

// viaEdge is what rendering e from n reads: the reference e reads, then what its leaf renders
// under the rows that reference pins.
func (f *readFold) viaEdge(n node, e renderEdge, route readRoute) []pathRead {
	if !e.readsRef() {
		return routed(f.reads(e.to), route)
	}
	a := e.read
	tr := tableReadOf(n.(*template).head(a.head), a, e.to)
	out := []pathRead{{route: route, a: a, tr: tr}}
	below := f.reads(e.to)
	if tr != nil {
		below = pinBelow(below, tr)
	}
	return append(out, routed(below, route)...)
}

// pinBelow is the reads below a table read, once its selected rows are pinned above them: a read
// under a cell those rows keep out is dropped, since that cell never renders on this route, and
// a read whose own selectors clash with them carries that clash.
func pinBelow(reads []pathRead, tr *tableRead) []pathRead {
	kept := reads[:0:0]
	for _, r := range reads {
		pins := r.branches.pins.clone()
		if err := tr.replay(&pins); err != nil {
			continue
		}
		if r.tr != nil && r.clash == nil {
			check := pins.clone()
			r.clash = r.tr.replay(&check)
		}
		r.branches.pins = pins
		kept = append(kept, r)
	}
	return kept
}

// rowReads is what the rows of t read: the cells of the columns its format renders, each tagged
// with its row, and the format's other reads, which every row shares. A format reaching its own
// cells through another category is checkOwnFamily's refusal, which runs before any survey of
// these reads.
func (f *readFold) rowReads(t *table, tag func(*branches, int) bool) []pathRead {
	var out []pathRead
	for _, e := range renderEdges(t.formatTemplate) {
		if c, isColumn := e.to.(*tableColumn); isColumn && !e.readsRef() {
			out = append(out, f.cellReads(c, tag)...)
			continue
		}
		out = append(out, f.viaEdge(t.formatTemplate, e, readRoute{})...)
	}
	return out
}

// cellReads is what the cells of column c read, each tagged with its row.
func (f *readFold) cellReads(c *tableColumn, tag func(*branches, int) bool) []pathRead {
	var out []pathRead
	for r := 0; r < c.t.rowCount(); r++ {
		cell := c.t.cellTemplate(r, c.i)
		if cell == nil {
			continue
		}
		for _, read := range f.reads(cell) {
			if tag(&read.branches, r) {
				out = append(out, read)
			}
		}
	}
	return out
}

// enter pins row r of t above a read, or reports that the rows already pinned keep it out.
func (b *branches) enter(t *table, r int) bool {
	if b.pins.clash(t, r) != nil {
		return false
	}
	b.pins = b.pins.entered(t, r)
	return true
}

// routed names the route the render's root reaches reads by, where none is named yet.
func routed(reads []pathRead, route readRoute) []pathRead {
	if route == (readRoute{}) {
		return reads
	}
	out := make([]pathRead, len(reads))
	for i, r := range reads {
		if r.route == (readRoute{}) {
			r.route = route
		}
		out[i] = r
	}
	return out
}

// spelled names the route, and the reference it reaches a read by where its root edge is not that
// reference.
func (r readRoute) spelled(ref string) string {
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
