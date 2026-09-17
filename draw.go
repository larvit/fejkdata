package fejkdata

import (
	"fmt"
	"sort"
	"strings"
)

// drawSet is one render's reference draws: the unnamed draw group's, and each named one's.
type drawSet struct {
	unnamed draws
	named   map[string]*draws
}

// drawScope is where a render reads its reference paths: a draw set, in the draw group of the
// template rendering, and the row a table's format is rendering, which its columns read.
type drawScope struct {
	set   *drawSet
	group string
	t     *table
	row   int
}

// newDrawSet makes the unnamed group's maps where the set is declared, keeping them on that frame's
// stack for a render that reads through them; a zero drawSet makes them on its first read instead.
func newDrawSet(s *session) drawSet {
	return drawSet{unnamed: draws{variant: map[string]node{}, value: map[string]draw{}, s: s}}
}

// renderOnce renders n as one render, over draws of its own.
func renderOnce(s *session, n node) string {
	set := drawSet{unnamed: draws{s: s}}
	return render(s, n, drawScope{set: &set})
}

// in is the scope t renders in: its draw group where it names one, else its caller's.
func (sc drawScope) in(t *template) drawScope {
	if t.drawGroupKey != "" {
		sc.group = t.drawGroupKey
	}
	return sc
}

// draws is the set's draws for sc's draw group. The session is passed in rather than
// read out of the set: copying it from there would leak the set's maps to the heap.
func (sc drawScope) draws(s *session) *draws {
	if sc.group == "" {
		return &sc.set.unnamed
	}
	d, drew := sc.set.named[sc.group]
	if !drew {
		if sc.set.named == nil {
			sc.set.named = map[string]*draws{}
		}
		d = &draws{variant: map[string]node{}, value: map[string]draw{}, s: s}
		sc.set.named[strings.Clone(sc.group)] = d // a key from sc would leak sc, and with it every render's draw set, to the heap
	}
	return d
}

// drawGroupOf reads a template's "drawGroup" (default ""), which a repeat cannot carry: each
// iteration renders in no draw group.
func drawGroupOf(m map[string]any, repeat int) (string, error) {
	v, ok := m["drawGroup"]
	if !ok {
		return "", nil
	}
	name, ok := v.(string)
	switch {
	case !ok:
		return "", fmt.Errorf("drawGroup must be a string, got %T", v)
	case name == "":
		return "", fmt.Errorf(`drawGroup "" is the default, so it has no effect; drop it`)
	case repeat > 1:
		return "", fmt.Errorf("drawGroup %q on a repeat names nothing, since each iteration is a render of its own; drop it", name)
	}
	return name, nil
}

// keyDrawGroup keys t's draw group by the category t sits in, "" for an inline template, so a name
// is local to its category.
func (t *template) keyDrawGroup(category string) {
	if t.drawGroup != "" {
		t.drawGroupKey = category + "/" + t.drawGroup
	}
}

// checkNestedDrawGroup refuses a template beneath one drawing in group that names group again,
// short of a repeat or another draw group.
func checkNestedDrawGroup(fields map[string]node, group string) error {
	if group == "" {
		return nil
	}
	var walk func(path string, n node) error
	walk = func(path string, n node) error {
		t, isTemplate := n.(*template)
		switch {
		case isTemplate && t.drawGroup == group:
			return fmt.Errorf("%q names drawGroup %q, the draw group this template draws in already; drop it", path, group)
		case isTemplate && (t.drawGroup != "" || t.repeat > 1):
			return nil
		}
		for _, c := range contained(n) {
			if err := walk(join(path, c.name), c.node); err != nil {
				return err
			}
		}
		return nil
	}
	for _, name := range sortedNames(fields) {
		if err := walk(name, fields[name]); err != nil {
			return err
		}
	}
	return nil
}

// drawCheck fences each template of a scope as a render of its own, remembering which nodes read a
// reference path.
type drawCheck struct {
	memo map[readsMemo]bool
}

type readsMemo struct {
	n           node
	stopAtGroup bool
}

// checkDrawGroup refuses a draw group that splits nothing: one whose render reads every reference
// path inside a repeat or a nested draw group, which draw apart from it whatever it names.
func (c *drawCheck) checkDrawGroup(path string, n node) error {
	if t, ok := n.(*template); ok && t.drawGroup != "" && !c.splitsDraws(t) {
		return fmt.Errorf("%s: drawGroup %q splits nothing, since nothing it renders reads a reference path outside a repeat or a nested drawGroup; drop it", path, t.drawGroup)
	}
	return nil
}

func (c *drawCheck) checkDraws(path string, n node) error {
	t, ok := n.(*template)
	if !ok || !c.readsPath(t) {
		return nil
	}
	w := newDrawWalk()
	for _, e := range renderEdges(t) {
		w.edge(t, e, drawAt{group: t.drawGroupKey, route: drawRoute{e.reached(), e.label}})
	}
	if err := w.checkOwnFamily(t); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if err := w.check(); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// checkOwnFamily refuses a table's format or cell that reads, however many templates
// away, a table of its own family: a row rendered whole draws its row without
// pinning it, so the family would draw apart from the row being rendered.
func (w *drawWalk) checkOwnFamily(t *template) error {
	own := t.table
	if own == nil {
		own = t.cellOf
	}
	if own == nil {
		return nil
	}
	for _, r := range w.reads {
		if r.tr != nil && r.tr.head.family() == own.family() {
			return fmt.Errorf("%s reads %s, a table of its own family, which a row of %s rendered whole would draw apart from; read the family from a template beside it, or add the value as a column", r.at.route.spelled(r.a.name), r.tr.head.category, own.category)
		}
	}
	return nil
}

// checkRecordDraws fences the columns of a record — a template compiled at the top without a
// repeat — so a load proves the record view of it as well as the string view.
func (c *drawCheck) checkRecordDraws(path string, n node) error {
	t, ok := n.(*template)
	if !ok || !t.record {
		return nil
	}
	// A record-only template's format renders nothing, so weigh the columns, not the format.
	columns := recordColumns(t)
	reads := false
	for _, name := range columns {
		reads = reads || c.readsPath(t.fields[name])
	}
	if !reads {
		return nil
	}
	if err := checkColumnDraws(t, columns); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// readsPath reports whether rendering n reads a reference path, short of a repeat, which renders
// over draws of its own.
func (c *drawCheck) readsPath(n node) bool { return c.reads(n, false) }

// splitsDraws reports whether rendering n reads a reference path that n's own draw group answers
// for: one outside a repeat and outside a nested draw group, which hold their own draws.
func (c *drawCheck) splitsDraws(n node) bool { return c.reads(n, true) }

// reads walks what rendering n renders for a reference path, stopping at a repeat — and at a nested
// draw group when stopAtGroup — since each holds draws of its own.
func (c *drawCheck) reads(n node, stopAtGroup bool) bool {
	k := readsMemo{n, stopAtGroup}
	if r, done := c.memo[k]; done {
		return r
	}
	r := false
	for _, e := range renderEdges(n) {
		if a, isRef := refRead(n, e.label); (isRef && (len(a.tail) > 0 || readsTable(n, a))) || (!repeats(e.to) && !(stopAtGroup && grouped(e.to)) && c.reads(e.to, stopAtGroup)) {
			r = true
			break
		}
	}
	if c.memo == nil {
		c.memo = map[readsMemo]bool{}
	}
	c.memo[k] = r
	return r
}

func repeats(n node) bool {
	t, isTemplate := n.(*template)
	return isTemplate && t.repeat > 1
}

func grouped(n node) bool {
	t, isTemplate := n.(*template)
	return isTemplate && t.drawGroupKey != ""
}

// readsTable reports whether a reference of n names a table, whose draw the render's
// group answers for even when read whole.
func readsTable(n node, a arm) bool {
	t, isTemplate := n.(*template)
	if !isTemplate {
		return false
	}
	_, isTable := t.fields[a.key].(*table)
	return isTable
}

// refRead is the reference an edge of n reads; false when the edge reads none.
func refRead(n node, label string) (arm, bool) {
	t, isTemplate := n.(*template)
	if !isTemplate {
		return arm{}, false
	}
	a := splitArm(label, t.refs)
	return a, isRef(a.key)
}

// checkColumnDraws fences a record's columns as one render.
func checkColumnDraws(t *template, columns []string) error {
	w := newDrawWalk()
	for _, name := range columns {
		w.walk(t.fields[name], drawAt{group: t.drawGroupKey, route: drawRoute{spelling: fmt.Sprintf("column %q", name)}})
	}
	return w.check()
}

// drawWalk gathers the references one render reads, by draw group, for check to compare.
type drawWalk struct {
	reads []pathRead
	read  map[drawKey]bool
	seen  map[drawVisit]bool
}

// drawAt is where a walk stands: the draw group it draws in, how the render's root reached it, and
// the table rows it entered.
type drawAt struct {
	group string
	route drawRoute
	alt   rowSet
}

// rowSet is the rows a walk entered, one per table: only one row of a table renders, so reads in
// two rows of one table never meet, while reads in one row, across its columns and whatever they
// reach, do.
type rowSet []tablePin

func (s rowSet) rowOf(t *table) (int, bool) {
	for _, p := range s {
		if p.t == t {
			return p.row, true
		}
	}
	return 0, false
}

// enter is s with row r of t, where t is not in it yet; the set is copied, since walks branch.
func (s rowSet) enter(t *table, r int) rowSet {
	if _, in := s.rowOf(t); in {
		return s
	}
	out := append(append(make(rowSet, 0, len(s)+1), s...), tablePin{t, r})
	sort.Slice(out, func(i, j int) bool { return out[i].t.category < out[j].t.category })
	return out
}

// key spells the set for a map, by the tables' identities.
func (s rowSet) key() string {
	var b strings.Builder
	for _, p := range s {
		fmt.Fprintf(&b, "%p[%d]", p.t, p.row)
	}
	return b.String()
}

// alternatives reports whether two reads sit in different rows of one table.
func alternatives(a, b drawAt) bool {
	for _, p := range a.alt {
		if r, in := b.alt.rowOf(p.t); in && r != p.row {
			return true
		}
	}
	return false
}

// drawRoute is how a render reaches a draw: as its author spells it, and the root edge's label.
type drawRoute struct{ spelling, label string }

// pathRead is one reference a render reads: a path, or a bare reference with no tail.
type pathRead struct {
	at drawAt
	a  arm
	tr *tableRead // set where the reference names a table
}

type drawVisit struct {
	n     node
	group string
	alt   string
}

type drawKey struct {
	group string
	path  string
	alt   string
}

func newDrawWalk() *drawWalk {
	return &drawWalk{read: map[drawKey]bool{}, seen: map[drawVisit]bool{}}
}

// walk follows what rendering n renders. A repeat renders over draws of its own, so the walk stops
// there.
func (w *drawWalk) walk(n node, at drawAt) {
	v := drawVisit{n, at.group, at.alt.key()}
	if w.seen[v] {
		return
	}
	w.seen[v] = true
	if repeats(n) {
		return
	}
	if t, isTemplate := n.(*template); isTemplate && t.drawGroupKey != "" {
		at.group = t.drawGroupKey
	}
	for _, e := range renderEdges(n) {
		if cell, isCell := e.to.(*template); isCell && cell.cellOf != nil {
			if at.alt.excludes(cell) {
				continue
			}
			w.edge(n, e, drawAt{at.group, at.route, at.alt.enter(cell.cellOf, cell.cellRow)})
			continue
		}
		w.edge(n, e, at)
	}
}

// excludes reports whether a cell's row cannot render with the rows entered: another
// row of its table, or a row outside an entered ancestor's.
func (s rowSet) excludes(cell *template) bool {
	for _, p := range s {
		if p.t == cell.cellOf {
			return p.row != cell.cellRow
		}
		if cell.cellOf.descends(p.t) && !cell.cellOf.under(cell.cellRow, p.t, p.row) {
			return true
		}
	}
	return false
}

// edge records the reference an edge reads, then walks on with every row the read
// pins entered, so a selected row renders only its own cells.
func (w *drawWalk) edge(from node, e renderEdge, at drawAt) {
	if a, reads := refRead(from, e.label); reads {
		tr := tableReadOf(from.(*template).fields[a.key], a, e.to)
		if k := (drawKey{at.group, a.path, at.alt.key()}); !w.read[k] {
			w.read[k] = true
			w.reads = append(w.reads, pathRead{at, a, tr})
		}
		if tr != nil {
			tr.pins.each(func(t *table, r int) { at.alt = at.alt.enter(t, r) })
		}
	}
	w.walk(e.to, at)
}

// check refuses what one draw per reference path cannot answer for: a read of a level beside a path
// another read takes into it, and two reads of one table family that select different rows. Reads
// are compared in path order, so which pair is reported does not vary.
func (w *drawWalk) check() error {
	sort.SliceStable(w.reads, func(i, j int) bool {
		if w.reads[i].at.group != w.reads[j].at.group {
			return w.reads[i].at.group < w.reads[j].at.group
		}
		return w.reads[i].a.path < w.reads[j].a.path
	})
	for i, level := range w.reads {
		for _, into := range w.reads[i+1:] {
			if into.at.group != level.at.group || alternatives(level.at, into.at) {
				continue
			}
			if strings.HasPrefix(into.a.path, level.a.path+".") && !(level.tr != nil && level.tr.whole) {
				return overlapError(level.at.route, level.a.name, into)
			}
		}
	}
	return checkFamilies(w.reads)
}

func overlapError(route drawRoute, ref string, into pathRead) error {
	return fmt.Errorf("%s renders a level that %s reads a path into; name the fields you want instead, or draw them apart with a drawGroup", route.spelled(ref), into.at.route.spelled(into.a.name))
}

// spelled names the route, and the reference it reaches a draw by where its root edge is not that
// reference.
func (r drawRoute) spelled(ref string) string {
	if ref == "" || ref == r.label {
		return r.spelling
	}
	return fmt.Sprintf("%s with {%s}", r.spelling, ref)
}
