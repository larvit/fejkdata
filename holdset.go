package fejkdata

import (
	"fmt"
	"strings"
)

// holdSet is one render's reference draws: the unnamed draw group's, and each named one's; and a
// test's trace of its reads.
type holdSet struct {
	unnamed groupHold
	named   map[string]*groupHold
	trace   renderTrace
}

// renderScope is where a render reads its reference paths: a hold set, in the draw group of the
// template rendering; and row, the row of the table rendering, which its columns read. It passes
// by value, and what is read from it reaches a map key, an interface or a func value only as a
// copy; else sc, and with it every render's hold set, moves to the heap.
type renderScope struct {
	set   *holdSet
	group string
	row   tablePin
}

// renderOnce renders n as one render, over a hold set of its own.
func renderOnce(s *session, n node) string {
	var set holdSet
	return render(s, n, renderScope{set: &set})
}

// expandAnew expands one repeat iteration of t as a render of its own, in no group. Inlined into
// render's loop, its hold set would move to the heap.
//
//go:noinline
func expandAnew(s *session, t *template) string {
	var set holdSet
	return expand(s, t, renderScope{set: &set})
}

// in is the scope t renders in: its draw group where it names one, else its caller's.
func (sc renderScope) in(t *template) renderScope {
	if t.link.drawGroupKey != "" {
		sc.group = t.link.drawGroupKey
	}
	return sc
}

// at is the scope n, the leaf of a path that pinned its rows in pins, renders in: where n is a
// table's row or column, at the row pins holds for that table.
func (sc renderScope) at(n node, pins *pinSet) renderScope {
	switch n := n.(type) {
	case *tableRow:
		sc.row = tablePin{n.t, pins.mustRow(n.t)}
	case *tableColumn:
		sc.row = tablePin{n.t, pins.mustRow(n.t)}
	}
	return sc
}

// rowOf is the row of t its columns render from.
func (sc renderScope) rowOf(t *table) int {
	if sc.row.t != t {
		panic(fmt.Sprintf("fejkdata: a column of %s is rendered with no row pinned", t.category))
	}
	return sc.row.row
}

func (sc renderScope) groupHold() *groupHold {
	if sc.group == "" {
		return &sc.set.unnamed
	}
	d, drew := sc.set.named[sc.group]
	if !drew {
		if sc.set.named == nil {
			sc.set.named = map[string]*groupHold{}
		}
		d = &groupHold{}
		sc.set.named[strings.Clone(sc.group)] = d
	}
	return d
}
