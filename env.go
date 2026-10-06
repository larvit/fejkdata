package fejkdata

import (
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/invariant"
)

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

// enter starts a read landing on n: the read sees no frame opened before it, and gets a frame for
// each name scope around n, from memo where there is one, so reads sharing memo read one pick of
// each name. It returns the mark that closes those frames.
func (env renderEnv) enter(n node, memo *drawMemo) (renderEnv, int) {
	env = env.entering()
	for scope := scopeAround(n); scope != nil; scope = scope.up {
		if len(scope.order) > 0 {
			env.frames.push(memo.enteredFrame(scope))
		}
	}
	return env, env.base
}

// entering is env as a read entering a category sees it: no frame opened before it, and no pick.
func (env renderEnv) entering() renderEnv {
	env.base, env.pick = len(env.frames.frames), nil
	return env
}

// renderFrame opens a fresh frame of t's scope, where t binds names and no read entering the
// category opened one, returning the mark that closes it, or -1.
func (env renderEnv) renderFrame(t *template) int {
	own := t.ownScope()
	if own == nil || env.frameOf(own) != nil {
		return -1
	}
	return env.frames.push(newPickFrame(own))
}

// frameOf is the frame of scope rendering since the read entering the category, nil where none is.
func (env renderEnv) frameOf(scope *nameScope) *pickFrame {
	for i := len(env.frames.frames) - 1; i >= env.base; i-- {
		if f := env.frames.frames[i]; f.scope == scope {
			return f
		}
	}
	return nil
}

// drawRowOf draws the row a render of t reads: inside the pick's rows where t renders as part of one.
func (env renderEnv) drawRowOf(s *drawstate.State, t *table) int {
	if env.pick != nil {
		return t.rows.DrawIn(s, &env.pick.pins)
	}
	return t.rows.Draw(s)
}

// keeps reports whether a read of env's name addresses the level a starts at.
func (env renderEnv) keeps(a arm) bool {
	_, kept := env.pick.named.addressed[env.pickAt.under(a.levels[0])]
	return kept
}
