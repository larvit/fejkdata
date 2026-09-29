package fejkdata

import (
	"fmt"
	"sort"
	"strings"
)

// drawFence fences each template of a scope as a render of its own, over one fold of what every
// node reads.
// docs/decisions.md#a-render-shares-one-reference-draw-per-category-per-group
// docs/decisions.md#the-expansion-hold-and-the-renders-draws-are-two-fences
type drawFence struct {
	fold *readFold
}

func (f *drawFence) reads() *readFold {
	if f.fold == nil {
		f.fold = newReadFold()
	}
	return f.fold
}

// checkDrawGroup refuses a draw group that splits nothing: one whose render reads every reference
// path inside a repeat or a nested draw group, which draw apart from it whatever it names.
func (f *drawFence) checkDrawGroup(path string, n node) error {
	t, ok := n.(*template)
	if !ok || t.drawGroup == "" {
		return nil
	}
	for _, r := range f.reads().rootReads(t) {
		if r.at.group == "" && r.draws() {
			return nil
		}
	}
	return fmt.Errorf("%s: drawGroup %q splits nothing, since nothing it renders reads a reference path outside a repeat or a nested drawGroup; drop it", path, t.drawGroup)
}

func (f *drawFence) checkDraws(path string, n node) error {
	t, ok := n.(*template)
	if !ok {
		return nil
	}
	reads := f.reads().rootReads(t)
	if !anyDraws(reads) {
		return nil
	}
	if err := survey(reads, t.link.drawGroupKey).check(); err != nil {
		return fmt.Errorf("%s: %w", path, err)
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
	reads := f.reads().columnReads(t, sortedNames(t.fields))
	if !anyDraws(reads) {
		return nil
	}
	if err := survey(reads, t.link.drawGroupKey).check(); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
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

// draws reports whether a read's draw is its group's to answer for: a path, or a table read whole,
// whose row the group draws.
func (r pathRead) draws() bool { return len(r.a.tail) > 0 || r.tr != nil }

func anyDraws(reads []pathRead) bool {
	for _, r := range reads {
		if r.draws() {
			return true
		}
	}
	return false
}

func repeats(n node) bool {
	t, isTemplate := n.(*template)
	return isTemplate && t.repeat > 1
}

// check refuses what one draw per reference path cannot answer for: a read selecting a row its route
// pinned another of, a read of a level beside a path another read takes into it, and two reads of one
// table family that select different rows.
func (s *readSurvey) check() error {
	if c := s.clash; c != nil {
		return fmt.Errorf("%s: %w; read the family from a template beside it, or add the value as a column", c.route.spelled(c.spelling), c.err)
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
// The reads come sorted by group and path.
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
