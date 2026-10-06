package fejkdata

import "github.com/larvit/fejkdata/internal/invariant"

// readValue is what one read drew: its text, and whether it landed on a null.
type readValue struct {
	text string
	null bool
}

// renderEnv is where a render reads: the frames of the name scopes rendering; and row, the row
// of the table rendering, which its columns read. It passes by value, and what is read from it
// reaches a map key, an interface or a func value only as a copy; else env, and with it every
// render's frames, moves to the heap.
type renderEnv struct {
	frames *frameStack
	row    renderedRow
	// base is the depth of the frame stack when the read rendering the category entered it,
	// below which no frame is the category's.
	base int
	// pick is the named pick being rendered, nil outside one; pickAt is the key in it of what renders.
	pick   *namedPick
	pickAt pickKey
}

type renderedRow struct {
	t     *table
	index int
}

// at returns the env n renders in, n being the leaf of a path whose rows sit in pins; where n
// is a table's row or column, the env carries the row pins holds for that table.
func (env renderEnv) at(n node, pins *pinSet) renderEnv {
	switch n := n.(type) {
	case *tableRow:
		env.row = renderedRow{n.t, pins.MustRow(n.t.rows)}
	case *tableColumn:
		env.row = renderedRow{n.t, pins.MustRow(n.t.rows)}
	}
	return env
}

// rowOf is the row of t its columns render from.
func (env renderEnv) rowOf(t *table) int {
	if env.row.t != t {
		panic(invariant.Broken("a column of %s renders in an env holding no row of it", t.rows.Segment()))
	}
	return env.row.index
}
