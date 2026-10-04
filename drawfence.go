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

// readSurvey is what one render reads, gathered at load for check to compare.
type readSurvey struct {
	reads []pathRead
	clash *pinClash
}

// pinClash is the read, lowest by how it is named, whose selectors clash with a row its route pinned.
type pinClash struct {
	route    readRoute
	spelling string
	err      error
}

func (c *pinClash) named() string { return c.route.spelled(c.spelling) }

func (c *pinClash) before(o *pinClash) bool {
	if c.named() != o.named() {
		return c.named() < o.named()
	}
	return c.err.Error() < o.err.Error()
}

// survey is the fold's result as a render in group: every read not inside a nested draw group
// draws in it, and one read gathered by several routes is kept once.
func survey(reads []pathRead, group string) *readSurvey {
	s := &readSurvey{}
	for _, r := range reads {
		if r.clash != nil {
			if c := (&pinClash{r.route, r.a.spelling, r.clash}); s.clash == nil || c.before(s.clash) {
				s.clash = c
			}
			continue
		}
		if r.group == "" {
			r.group = group
		}
		s.reads = append(s.reads, r)
	}
	s.reads = distinct(s.reads)
	return s
}

// checkDrawGroup refuses a draw group that splits nothing: one whose render reads every reference
// path inside a repeat or a nested draw group, which draw apart from it whatever it names.
func (f *drawFence) checkDrawGroup(path string, n node) error {
	t, ok := n.(*template)
	if !ok || t.drawGroup == "" {
		return nil
	}
	for _, r := range f.fold.rootReads(t) {
		if r.group == "" && r.draws() {
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
	reads := f.fold.rootReads(t)
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
	reads := f.fold.columnReads(t, sortedNames(t.fields))
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
// pinned another of, and any two reads that render together yet conflict.
func (s *readSurvey) check() error {
	if c := s.clash; c != nil {
		return fmt.Errorf("%s: %w; read the family from a template beside it, or add the value as a column", c.named(), c.err)
	}
	sort.SliceStable(s.reads, func(i, j int) bool {
		if s.reads[i].group != s.reads[j].group {
			return s.reads[i].group < s.reads[j].group
		}
		return s.reads[i].a.path < s.reads[j].a.path
	})
	for i, a := range s.reads {
		for _, b := range s.reads[i+1:] {
			if !coRender(a, b) {
				continue
			}
			if err := conflict(a, b); err != nil {
				return err
			}
		}
	}
	return nil
}

// coRender reports whether two reads can render together: in one draw group, and under no two
// rows of one table, since one read renders one row of it.
func coRender(a, b pathRead) bool {
	return a.group == b.group && !a.branches.pins.differs(&b.branches.pins) && !a.branches.wholePins.differs(&b.branches.wholePins)
}

// conflict refuses two reads rendering together, a before b by path, that one draw per reference
// path cannot answer for. Pairs find every conflict a replay of all the reads would: a row clashes
// with one pinned table, and the read that pinned it holds that pin itself, with its ancestors.
func conflict(a, b pathRead) error {
	for _, pair := range [][2]pathRead{{a, b}, {b, a}} {
		if err := drawsApart(pair[0], pair[1]); err != nil {
			return err
		}
	}
	if a.tr == nil || b.tr == nil {
		return nil
	}
	d := a.tr.pins.clone()
	if err := b.tr.replay(&d); err != nil {
		return fmt.Errorf("%s: %w; select the same rows in every path into the family, or draw them apart with a drawGroup", b.route.spelled(b.a.spelling), err)
	}
	return nil
}

// drawsApart refuses x drawing apart from what y reads: a level beside a path y takes into it, a
// table x reads whole beside a path y takes into its family, and a table x draws that y pins, since
// which token renders first would then decide the row. A level landing on a row renders the row its
// selectors pin, the one a path into it reads too.
// docs/decisions.md#a-bare-reference-draws-each-time-a-reference-path-is-held
func drawsApart(x, y pathRead) error {
	if strings.HasPrefix(y.a.path, x.a.path+".") && !(x.tr != nil && x.tr.landsRow) {
		return overlapError(x.route, x.a.spelling, y)
	}
	if x.tr == nil || y.tr == nil || x.tr.headTable.familyRoot() != y.tr.headTable.familyRoot() {
		return nil
	}
	if len(x.a.tail) == 0 && len(y.a.tail) > 0 {
		return overlapError(x.route, x.a.spelling, y)
	}
	drawn := x.tr.drawnOf(&y.tr.pins)
	if drawn == nil {
		return nil
	}
	if s, ok := y.tr.selected(x.tr.headTable); ok && len(x.tr.sels) == 0 {
		tail := x.a.tail
		if s.t != x.tr.headTable {
			tail = append([]string{x.tr.headTable.segment}, tail...)
		}
		return fmt.Errorf("%s draws %s, which %s selects a row of; write {%s}, or draw them apart with a drawGroup",
			x.route.spelled(x.a.spelling), drawn.segment, y.route.spelled(y.a.spelling), joinSegments(append([]string{s.spelling}, tail...)))
	}
	return fmt.Errorf("%s draws %s, which %s selects a row of; select that row in both, or draw them apart with a drawGroup",
		x.route.spelled(x.a.spelling), drawn.segment, y.route.spelled(y.a.spelling))
}

func overlapError(route readRoute, ref string, into pathRead) error {
	apart := ""
	switch {
	case route.spelling == into.route.spelling:
	case isRef(into.route.label):
		apart = fmt.Sprintf(", or move {%s} into a field with a drawGroup", into.route.label)
	default:
		apart = fmt.Sprintf(", or give %s a drawGroup", into.route.spelling)
	}
	return fmt.Errorf("%s renders its own draw of what %s reads a path through; name the fields you want instead%s", route.spelled(ref), into.route.spelled(into.a.spelling), apart)
}
