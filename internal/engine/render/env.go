package render

import (
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
)

// readValue is what one read drew: its text, and whether it landed on a null.
type readValue struct {
	text string
	null bool
}

// frameStack is what lasts a whole render: the frames of the name scopes rendering, innermost
// last, and the fakes its {reword()} calls drew. A render's env points
// at it, and a repeat iteration's shares it: holding the frames in a renderEnv would move every
// render's env to the heap.
type frameStack struct {
	frames []*pickFrame
	fakes  map[fakeKey]string // each word a {reword()} call met in the render, and its fake
}

func (st *frameStack) push(f *pickFrame) int {
	mark := len(st.frames)
	st.frames = append(st.frames, f)
	return mark
}

func (st *frameStack) pop(mark int) { st.frames = st.frames[:mark] }

// renderEnv is where a render reads: the frames of the name scopes rendering; and row, the row
// of the table rendering, which its columns read. It passes by value, and what is read from it
// reaches a map key, an interface or a func value only as a copy; else env, and with it every
// render's frames, moves to the heap.
type renderEnv struct {
	frames *frameStack
	row    Row
	// base is the depth of the frame stack when the read rendering the category entered it,
	// below which no frame is the category's.
	base int
	// pick is the named pick being rendered, nil outside one; pickAt is the key in it of what renders.
	pick   *namedPick
	pickAt tree.PickKey
}

// Row is the row of Table whose columns a render reads; a zero Row is no row.
type Row struct {
	Table *tree.Table
	Index int
}

// at returns the env n renders in, n being the leaf of a path whose rows sit in pins; where n
// is a table's row or column, the env carries the row pins holds for that table.
func (env renderEnv) at(n tree.Node, pins *tree.PinSet) renderEnv {
	switch n := n.(type) {
	case *tree.TableRow:
		env.row = Row{Table: n.Table(), Index: pins.MustRow(n.Table().Rows())}
	case *tree.TableColumn:
		env.row = Row{Table: n.Table(), Index: pins.MustRow(n.Table().Rows())}
	}
	return env
}

func (env renderEnv) rowOf(t *tree.Table) int {
	if env.row.Table != t {
		panic(invariant.Broken("a column of %s renders in an env holding no row of it", t.Rows().Segment()))
	}
	return env.row.Index
}

// enter starts a read landing on n: the read sees no frame opened before it, and gets a frame for
// each name scope around n. A frame comes from memo where memo is non-nil, so reads sharing memo
// read one pick of each name. It returns env.base, the frame count before it pushed any: popping
// the frames to that mark closes them all.
func (env renderEnv) enter(n tree.Node, memo *drawMemo) (renderEnv, int) {
	env = env.entering()
	for scope := scopeAround(n); scope != nil; scope = scope.Up() {
		if len(scope.Order()) > 0 {
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
func (env renderEnv) renderFrame(t *tree.Template) int {
	own := t.OwnScope()
	if own == nil || env.frameOf(own) != nil {
		return -1
	}
	return env.frames.push(newPickFrame(own))
}

// frameOf is the frame of scope rendering since the read entering the category, nil where none is.
func (env renderEnv) frameOf(scope *tree.NameScope) *pickFrame {
	for i := len(env.frames.frames) - 1; i >= env.base; i-- {
		if f := env.frames.frames[i]; f.scope == scope {
			return f
		}
	}
	return nil
}

// drawRowOf draws the row a render of t reads: inside the pick's rows where t renders as part of one.
func (env renderEnv) drawRowOf(s *drawstate.State, t *tree.Table) int {
	if env.pick != nil {
		return t.Rows().DrawIn(s, &env.pick.pins)
	}
	return t.Rows().Draw(s)
}
