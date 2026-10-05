package fejkdata

// drawMemo is what a named pick has drawn: the variant each level was drawn as, so every path
// under it reads one variant; the value each read produced, by its path, so the same read written
// twice reads one value; the frames a read opens for the name scopes around where it lands
// (enter); and the rows drawn by each step down after a "..", by level.
type drawMemo struct {
	variant       map[string]node
	value         map[string]readValue
	enteredFrames map[*nameScope]*pickFrame
	steppedDown   map[string]*pinSet
}

// stepDownPins is the pins a path steps down into after stepping up to t: kept in m where there
// is one, so every path stepping down there reads one draw.
func (m *drawMemo) stepDownPins(pins *pinSet, t *table, levels []string, at int) *pinSet {
	if m == nil {
		return pins.above(t)
	}
	p, ok := m.steppedDown[levels[at]]
	if !ok {
		p = pins.above(t)
		if m.steppedDown == nil {
			m.steppedDown = map[string]*pinSet{}
		}
		m.steppedDown[levels[at]] = p
	}
	return p
}

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
