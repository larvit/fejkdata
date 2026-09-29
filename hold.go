package fejkdata

import "strings"

// hold is what one expansion has drawn for the names it holds: the variant each was drawn as, so
// every path under it reads one variant; and the value each read produced, by its path, so the
// same read written twice reads one value.
type hold struct {
	variant map[string]node
	value   map[string]readValue
}

// groupHold is what one render's draw group has drawn for its reference paths: a hold, and the
// table rows those paths pinned.
type groupHold struct {
	hold
	pins pinSet
}

// readValue is what one read drew: its text, and whether it landed on a null.
type readValue struct {
	text string
	null bool
}

// renderTrace is a test's view of the fields a render reads; nil outside a test.
type renderTrace func(group, table string, row int, a arm)

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
		panic(internalError("a column of %s renders in a scope holding no row of it", t.segment))
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
