package fejkdata

import "github.com/larvit/fejkdata/internal/invariant"

// readValue is what one read drew: its text, and whether it landed on a null.
type readValue struct {
	text string
	null bool
}

// renderScope is where a render reads: the frames of the name scopes rendering; and row, the row
// of the table rendering, which its columns read. It passes by value, and what is read from it
// reaches a map key, an interface or a func value only as a copy; else sc, and with it every
// render's frames, moves to the heap.
type renderScope struct {
	frames *frameStack
	row    renderedRow
	// base is the depth of the frame stack when the read rendering the category entered it,
	// below which no frame is the category's.
	base int
	// pick is the named pick being rendered, nil outside one; pickKey is the path from the name
	// to what renders.
	pick    *namedPick
	pickKey string
}

type renderedRow struct {
	t     *table
	index int
}

// at returns the scope n renders in, n being the leaf of a path whose rows sit in pins; where n
// is a table's row or column, the scope carries the row pins holds for that table.
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
		panic(invariant.Broken("a column of %s renders in a scope holding no row of it", t.segment))
	}
	return sc.row.index
}
