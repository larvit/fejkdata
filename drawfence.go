package fejkdata

import "fmt"

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
	own := t.site.table
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
		return fmt.Errorf("%s reads %s, a table of its own family, which a bare read of %s would draw apart from the row it renders; read the family from a template beside it, or add the value as a column", e.reached(), familyTable.segment, own.segment)
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
	_, isTable := t.head(a.head).(*table)
	return isTable
}

func checkColumnDraws(t *template, columns []string) error { return surveyColumns(t, columns).check() }
