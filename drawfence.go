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

// drawFence fences each template of a scope as a render of its own, remembering which nodes read a
// reference path.
// docs/decisions.md#a-render-shares-one-reference-draw-per-category-per-group
// docs/decisions.md#the-expansion-hold-and-the-renders-draws-are-two-fences
type drawFence struct {
	memo map[hasReadMemo]bool
}

type hasReadMemo struct {
	n           node
	stopAtGroup bool
}

// checkDrawGroup refuses a draw group that splits nothing: one whose render reads every reference
// path inside a repeat or a nested draw group, which draw apart from it whatever it names.
func (f *drawFence) checkDrawGroup(path string, n node) error {
	if t, ok := n.(*template); ok && t.drawGroup != "" && !f.splitsDraws(t) {
		return fmt.Errorf("%s: drawGroup %q splits nothing, since nothing it renders reads a reference path outside a repeat or a nested drawGroup; drop it", path, t.drawGroup)
	}
	return nil
}

func (f *drawFence) checkDraws(path string, n node) error {
	t, ok := n.(*template)
	if !ok {
		return nil
	}
	if err := checkOwnFamily(t); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if !f.readsPath(t) {
		return nil
	}
	if err := surveyRender(t).check(); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// checkOwnFamily refuses a table's format or cell that reads, however many templates
// away and through a repeat or a draw group too, a table of its own family: a whole
// read draws its row without pinning it, so the family would draw apart from the row
// it renders, whichever draws the reaching template holds.
// docs/decisions.md#a-table-never-reaches-its-own-family-by-any-route
func checkOwnFamily(t *template) error {
	own := t.table
	if own == nil {
		own = t.cell.table
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
			if e.readsRef() {
				if head, isTable := n.(*template).head(e.read.key).(*table); isTable && head.familyRoot() == own.familyRoot() {
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
		return fmt.Errorf("%s reads %s, a table of its own family, which a bare read of %s would draw apart from the row it renders; read the family from a template beside it, or add the value as a column", e.reached(), head.category, own.category)
	}
	return nil
}

// checkRecordDraws fences the columns of a record — a template compiled at the top without a
// repeat — so a load proves the record view of it as well as the string view.
// docs/decisions.md#a-category-never-references-itself-and-a-records-fences-run-at-load
func (f *drawFence) checkRecordDraws(path string, n node) error {
	t, ok := n.(*template)
	if !ok || !t.isRecord {
		return nil
	}
	// A record-only template's format renders nothing, so weigh the columns, not the format.
	columns := sortedNames(t.fields)
	reads := false
	for _, name := range columns {
		reads = reads || f.readsPath(t.fields[name])
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
func (f *drawFence) readsPath(n node) bool { return f.hasRead(n, false) }

// splitsDraws reports whether rendering n reads a reference path that n's own draw group answers
// for: one outside a repeat and outside a nested draw group, which hold their own draws.
func (f *drawFence) splitsDraws(n node) bool { return f.hasRead(n, true) }

// hasRead walks what rendering n renders for a reference path, stopping at a repeat — and at a nested
// draw group when stopAtGroup — since each holds draws of its own.
func (f *drawFence) hasRead(n node, stopAtGroup bool) bool {
	k := hasReadMemo{n, stopAtGroup}
	if r, done := f.memo[k]; done {
		return r
	}
	r := false
	for _, e := range renderEdges(n) {
		if readsOnEdge(n, e) || (walksInto(e.to, stopAtGroup) && f.hasRead(e.to, stopAtGroup)) {
			r = true
			break
		}
	}
	if f.memo == nil {
		f.memo = map[hasReadMemo]bool{}
	}
	f.memo[k] = r
	return r
}

func readsOnEdge(n node, e renderEdge) bool {
	return e.readsRef() && (len(e.read.tail) > 0 || readsTable(n, e.read))
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
	return isTemplate && t.link.drawGroupKey != ""
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

// checkColumnDraws fences a record's columns as one render.
func checkColumnDraws(t *template, columns []string) error { return surveyColumns(t, columns).check() }

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
					tail = append([]string{x.tr.headTable.category}, tail...)
				}
				return fmt.Errorf("%s draws %s, which %s selects a row of; write {%s.%s}, or draw them apart with a drawGroup",
					x.at.route.spelled(x.a.spelling), drawn.category, y.at.route.spelled(y.a.spelling), s.spelling, joinSegments(tail))
			}
			return fmt.Errorf("%s draws %s, which %s selects a row of; select that row in both, or draw them apart with a drawGroup",
				x.at.route.spelled(x.a.spelling), drawn.category, y.at.route.spelled(y.a.spelling))
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
