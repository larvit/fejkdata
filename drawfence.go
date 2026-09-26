package fejkdata

import (
	"fmt"
	"sort"
	"strings"
)

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
// docs/decisions.md#a-render-shares-one-reference-draw-per-category-per-group
// docs/decisions.md#the-expansion-hold-and-the-renders-draws-are-two-fences
type drawCheck struct {
	memo map[hasReadMemo]bool
}

type hasReadMemo struct {
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
	if !ok {
		return nil
	}
	if err := checkOwnFamily(t); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if !c.readsPath(t) {
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
// docs/decisions.md#a-category-never-references-itself-and-a-records-fences-run-at-load
func (c *drawCheck) checkRecordDraws(path string, n node) error {
	t, ok := n.(*template)
	if !ok || !t.isRecord {
		return nil
	}
	// A record-only template's format renders nothing, so weigh the columns, not the format.
	columns := sortedNames(t.fields)
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
func (c *drawCheck) readsPath(n node) bool { return c.hasRead(n, false) }

// splitsDraws reports whether rendering n reads a reference path that n's own draw group answers
// for: one outside a repeat and outside a nested draw group, which hold their own draws.
func (c *drawCheck) splitsDraws(n node) bool { return c.hasRead(n, true) }

// hasRead walks what rendering n renders for a reference path, stopping at a repeat — and at a nested
// draw group when stopAtGroup — since each holds draws of its own.
func (c *drawCheck) hasRead(n node, stopAtGroup bool) bool {
	k := hasReadMemo{n, stopAtGroup}
	if r, done := c.memo[k]; done {
		return r
	}
	r := false
	for _, e := range renderEdges(n) {
		if readsOnEdge(n, e.label) || (walksInto(e.to, stopAtGroup) && c.hasRead(e.to, stopAtGroup)) {
			r = true
			break
		}
	}
	if c.memo == nil {
		c.memo = map[hasReadMemo]bool{}
	}
	c.memo[k] = r
	return r
}

func readsOnEdge(n node, label string) bool {
	a, isRef := refRead(n, label)
	return isRef && (len(a.tail) > 0 || readsTable(n, a))
}

func walksInto(to node, stopAtGroup bool) bool {
	return !repeats(to) && !(stopAtGroup && grouped(to))
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
	_, isTable := t.head(a.key).(*table)
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
	read  map[readKey]bool
	seen  map[nodeVisit]bool
}

// drawAt is where a walk stands: the draw group it draws in, how the render's root reached it, the
// table rows it pinned, the table a whole read draws a row of, and the rows of whole draws whose
// cells the walk is in.
type drawAt struct {
	group      string
	route      drawRoute
	pins       pinSet
	wholeTable *table // left out of the visit keys: only this table's own cells compare against it, which the own-family fence keeps true
	wholePins  pinSet
}

// drawRoute is how a render reaches a draw: as its author spells it, and the root edge's label.
type drawRoute struct{ spelling, label string }

// pathRead is one reference a render reads: a path, or a bare reference with no tail.
type pathRead struct {
	at drawAt
	a  arm
	tr *tableRead // set where the reference names a table
}

type nodeVisit struct {
	n       node
	group   string
	rowsKey string
}

type readKey struct {
	group   string
	path    string
	rowsKey string
}

func newDrawWalk() *drawWalk {
	return &drawWalk{read: map[readKey]bool{}, seen: map[nodeVisit]bool{}}
}

// walk follows what rendering n renders. A repeat renders over draws of its own, so the walk stops
// there. A cell whose row the pins keep out has no case: a pinned table renders its pinned row
// alone and an unpinned one draws a row inside its nearest pinned ancestor's, so that cell never
// renders on this route.
func (w *drawWalk) walk(n node, at drawAt) {
	v := nodeVisit{n, at.group, at.rowsKey()}
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
	if t, isTable := n.(*table); isTable {
		at.wholeTable = t
	}
	for _, e := range renderEdges(n) {
		cell, isCell := e.to.(*template)
		switch {
		case !isCell || cell.cellOf == nil:
			w.edge(n, e, at)
		case cell.cellOf == at.wholeTable:
			in := at
			in.wholePins = at.wholePins.clone()
			in.wholePins.add(cell.cellOf, cell.cellRow)
			w.edge(n, e, in)
		case at.pins.clash(cell.cellOf, cell.cellRow) == nil:
			in := at
			in.pins = at.pins.entered(cell.cellOf, cell.cellRow)
			w.edge(n, e, in)
		}
	}
}

// rowsKey spells the rows the walk stands in, pinned and whole, for a map.
func (at drawAt) rowsKey() string { return at.pins.mapKey() + "|" + at.wholePins.mapKey() }

// edge records the reference an edge reads, then walks on with every row the read
// pins entered, so a selected row renders only its own cells.
func (w *drawWalk) edge(from node, e renderEdge, at drawAt) {
	if a, reads := refRead(from, e.label); reads {
		tr := tableReadOf(from.(*template).head(a.key), a, e.to)
		if k := (readKey{at.group, a.path, at.rowsKey()}); !w.read[k] {
			w.read[k] = true
			w.reads = append(w.reads, pathRead{at, a, tr})
		}
		if tr != nil {
			tr.pins.each(func(t *table, r int) { at.pins = at.pins.entered(t, r) })
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
			if strings.HasPrefix(into.a.path, level.a.path+".") && !(level.tr != nil && level.tr.landsWhole) {
				return overlapError(level.at.route, level.a.spelling, into)
			}
		}
	}
	return checkFamilies(w.reads)
}

func overlapError(route drawRoute, ref string, into pathRead) error {
	return fmt.Errorf("%s renders a level that %s reads a path into; name the fields you want instead, or draw them apart with a drawGroup", route.spelled(ref), into.at.route.spelled(into.a.spelling))
}

// spelled names the route, and the reference it reaches a draw by where its root edge is not that
// reference.
func (r drawRoute) spelled(ref string) string {
	if ref == "" || ref == r.label {
		return r.spelling
	}
	return fmt.Sprintf("%s with {%s}", r.spelling, ref)
}
