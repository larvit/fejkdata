package fejkdata

import "strings"

// drawMemo is what a hold or a draw group has drawn: the variant each level was drawn as, so
// every path under it reads one variant; and the value each read produced, by its path, so the
// same read written twice reads one value.
type drawMemo struct {
	variant map[string]node
	value   map[string]readValue
}

// groupDraws is what one render's draw group has drawn for its reference paths: a memo, and the
// table rows those paths pinned.
type groupDraws struct {
	memo drawMemo
	pins pinSet
}

// readValue is what one read drew: its text, and whether it landed on a null.
type readValue struct {
	text string
	null bool
}

// renderTrace is a test's view of the fields a render reads; nil outside a test.
type renderTrace func(group string, row renderedRow, a arm)

// renderDraws is one render's reference draws: the unnamed draw group's, and each named one's;
// and a test's trace of its reads.
type renderDraws struct {
	unnamed groupDraws
	named   map[string]*groupDraws
	trace   renderTrace
}

// renderScope is where a render reads its reference paths: its reference draws, in the draw group of the
// template rendering; and row, the row of the table rendering, which its columns read. It passes
// by value, and what is read from it reaches a map key, an interface or a func value only as a
// copy; else sc, and with it every render's draws, moves to the heap.
type renderScope struct {
	draws *renderDraws
	group string
	row   renderedRow
}

type renderedRow struct {
	t     *table
	index int
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
		sc.row = renderedRow{n.t, pins.mustRow(n.t)}
	case *tableColumn:
		sc.row = renderedRow{n.t, pins.mustRow(n.t)}
	}
	return sc
}

// rowOf is the row of t its columns render from.
func (sc renderScope) rowOf(t *table) int {
	if sc.row.t != t {
		panic(internalError("a column of %s renders in a scope holding no row of it", t.segment))
	}
	return sc.row.index
}

func (sc renderScope) groupDraws() *groupDraws {
	if sc.group == "" {
		return &sc.draws.unnamed
	}
	d, drew := sc.draws.named[sc.group]
	if !drew {
		if sc.draws.named == nil {
			sc.draws.named = map[string]*groupDraws{}
		}
		d = &groupDraws{}
		sc.draws.named[strings.Clone(sc.group)] = d
	}
	return d
}
