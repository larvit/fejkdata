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

// enter opens the frames of a read landing on n: one for each name scope around n, innermost
// last. A frame comes from memo where memo is non-nil, and is made in it where it holds none, so
// reads sharing memo read one pick of each name. It returns the frame count before it pushed any:
// popping to that mark closes them all, and the read renders in env.entered(mark).
func (st *frameStack) enter(n tree.Node, memo *drawMemo) (mark int) {
	mark = len(st.frames)
	for scope := scopeAround(n); scope != nil; scope = scope.Up() {
		if len(scope.Order()) > 0 {
			st.push(memo.enteredFrame(scope))
		}
	}
	return mark
}

// openOwn opens a fresh frame of t's scope, where t binds names and no frame from base up is that
// scope's, which is where no read entering the category opened one. It returns the mark that
// closes the frame, or -1.
func (st *frameStack) openOwn(t *tree.Template, base int) (mark int) {
	own := t.OwnScope()
	if own == nil || st.frameOf(own, base) != nil {
		return -1
	}
	return st.push(newPickFrame(own))
}

// frameOf is the frame of scope at base or above, innermost first, nil where none is.
func (st *frameStack) frameOf(scope *tree.NameScope, base int) *pickFrame {
	for i := len(st.frames) - 1; i >= base; i-- {
		if f := st.frames[i]; f.scope == scope {
			return f
		}
	}
	return nil
}

// renderEnv is where a render reads: the frames of the name scopes rendering; and row, the row
// of the table rendering, which its columns read. No method of it changes the frames or a memo:
// a call that does is a method of *frameStack, written env.frames.enter, openOwn, push or pop.
// It passes by value, and what is read from it
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

// entered is env as a read sees it that entered a category when the frames stood at mark: no
// frame below mark, and no pick.
func (env renderEnv) entered(mark int) renderEnv {
	env.base, env.pick = mark, nil
	return env
}

// frameOf is the frame of scope rendering since the read entering the category, nil where none is.
func (env renderEnv) frameOf(scope *tree.NameScope) *pickFrame {
	return env.frames.frameOf(scope, env.base)
}

// drawRowOf draws the row a render of t reads: inside the pick's rows where t renders as part of one.
func (env renderEnv) drawRowOf(s *drawstate.State, t *tree.Table) int {
	if env.pick != nil {
		return t.Rows().DrawIn(s, &env.pick.pins)
	}
	return t.Rows().Draw(s)
}
