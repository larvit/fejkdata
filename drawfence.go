package fejkdata

import "fmt"

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
	if !f.readsPath(t) {
		return nil
	}
	if err := surveyRender(t).check(); err != nil {
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
	// A record-only template's format renders nothing, so weigh the columns, not the format.
	columns := sortedNames(t.fields)
	reads := false
	for _, name := range columns {
		reads = reads || f.readsPath(t.fields[name])
	}
	if !reads {
		return nil
	}
	if err := surveyColumns(t, columns).check(); err != nil {
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
