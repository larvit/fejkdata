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

// selectRow pins the row a selector names, refusing one outside the rows pinned
// before it.
func (d *draws) selectRow(t *table, sel string) error {
	r, err := t.find(sel, d)
	if err != nil {
		return err
	}
	if pr, ok := d.pinned(t); ok && pr != r {
		return fmt.Errorf("%s is not %s, the row already drawn", t.selectorSpelling(r), t.selectorSpelling(pr))
	}
	for a := t.parentT; a != nil; a = a.parentT {
		if pa, ok := d.pinned(a); ok && !t.under(r, a, pa) {
			return fmt.Errorf("%s is not inside %s", t.selectorSpelling(r), a.selectorSpelling(pa))
		}
	}
	d.pin(t, r)
	return nil
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
	if err := w.check(); err != nil {
		return fmt.Errorf("%s: %w", path, err)
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

// drawAt is where a walk stands: the draw group it draws in, and how the render's root reached it.
type drawAt struct {
	group string
	route drawRoute
}

// drawRoute is how a render reaches a draw: as its author spells it, and the root edge's label.
type drawRoute struct{ spelling, label string }

// pathRead is one reference a render reads: a path, or a bare reference with no tail.
type pathRead struct {
	at drawAt
	a  arm
	tr *tableRead // set where the reference names a table
}

// tableRead is what a reference reads of a table family: the table named, the rows
// its selectors pin along the way, and whether it lands on a row rendered whole.
type tableRead struct {
	head  *table
	sels  []tableSel
	whole bool
}

// tableSel is one selector on the way: the table it selects a row of, and the
// selector's spelling up to its closing bracket.
type tableSel struct {
	t        *table
	row      int
	spelling string
}

// tableReadOf reads what a reference path does of a table, replaying its selectors
// over a walk of its own — checkPath proved each names a row — so two paths naming
// one row by key and by name compare equal.
func tableReadOf(head node, a arm, leaf node) *tableRead {
	t, isTable := head.(*table)
	if !isTable {
		return nil
	}
	tr := &tableRead{head: t}
	var pins draws
	cur, seen := t, a.name
	for _, seg := range a.tail {
		switch {
		case isSelector(seg):
			end := strings.Index(seen, "]") + 1
			_ = pins.selectRow(cur, selectorOf(seg))
			row, _ := pins.pinned(cur)
			tr.sels = append(tr.sels, tableSel{cur, row, a.name[:len(a.name)-len(seen)+end]})
			seen = seen[end:]
		case cur.children[seg] != nil:
			cur = cur.children[seg]
		}
	}
	c, isColumn := leaf.(*column)
	tr.whole = isColumn && c.i < 0
	return tr
}

// family is the table a chain of parents ends at.
func (t *table) family() *table {
	for t.parentT != nil {
		t = t.parentT
	}
	return t
}

// selected is the selector in r that pins t, or the nearest ancestor of t it pins.
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

func (r *tableRead) selection() string {
	parts := make([]string, len(r.sels))
	for i, s := range r.sels {
		parts[i] = fmt.Sprintf("%s[%d]", s.t.category, s.row)
	}
	return strings.Join(parts, " ")
}

type drawVisit struct {
	n     node
	group string
}

type drawKey struct {
	group string
	path  string
}

func newDrawWalk() *drawWalk {
	return &drawWalk{read: map[drawKey]bool{}, seen: map[drawVisit]bool{}}
}

// walk follows what rendering n renders. A repeat renders over draws of its own, so the walk stops
// there.
func (w *drawWalk) walk(n node, at drawAt) {
	v := drawVisit{n, at.group}
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
		w.edge(n, e, at)
	}
}

func (w *drawWalk) edge(from node, e renderEdge, at drawAt) {
	if a, reads := refRead(from, e.label); reads {
		if k := (drawKey{at.group, a.path}); !w.read[k] {
			w.read[k] = true
			w.reads = append(w.reads, pathRead{at, a, tableReadOf(from.(*template).fields[a.key], a, e.to)})
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
			if into.at.group != level.at.group {
				continue
			}
			if strings.HasPrefix(into.a.path, level.a.path+".") && !(level.tr != nil && level.tr.whole) {
				return overlapError(level.at.route, level.a.name, into)
			}
			if err := checkFamily(level, into); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkFamily refuses two reads of one table family in one group that cannot read one consistent
// draw: a table rendered whole beside a path into the family, and two paths selecting different rows.
func checkFamily(a, b pathRead) error {
	if a.tr == nil || b.tr == nil || a.tr.head.family() != b.tr.head.family() {
		return nil
	}
	for _, pair := range [][2]pathRead{{a, b}, {b, a}} {
		bare, path := pair[0], pair[1]
		if len(bare.a.tail) == 0 && len(path.a.tail) > 0 {
			return overlapError(bare.at.route, bare.a.name, path)
		}
	}
	if len(a.a.tail) == 0 || len(b.a.tail) == 0 || a.tr.selection() == b.tr.selection() {
		return nil
	}
	for _, pair := range [][2]pathRead{{a, b}, {b, a}} {
		selected, plain := pair[0], pair[1]
		if len(plain.tr.sels) > 0 {
			continue
		}
		if s, ok := selected.tr.selected(plain.tr.head); ok {
			tail := plain.a.tail
			if s.t != plain.tr.head {
				tail = append([]string{plain.tr.head.category}, tail...)
			}
			return fmt.Errorf("%s reads %s without the row %s selects; write {%s.%s}, or draw them apart with a drawGroup",
				plain.at.route.spelled(plain.a.name), plain.tr.head.category, selected.at.route.spelled(selected.a.name), s.spelling, strings.Join(tail, "."))
		}
	}
	return fmt.Errorf("%s and %s select different rows of one table family; select the same rows in both, or draw them apart with a drawGroup",
		a.at.route.spelled(a.a.name), b.at.route.spelled(b.a.name))
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
