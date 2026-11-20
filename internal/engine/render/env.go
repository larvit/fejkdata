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

// rewordFakes is the fakes a render's {reword()} calls drew, which last the whole render: each
// word a call met, and its fake. A render's env points at it, and every env copied from that one
// shares it.
type rewordFakes struct {
	byWord map[fakeKey]string
}

// renderEnv is where a render reads: frames, the frames of the name scopes rendering in the
// category; row, the row of the table rendering, which its columns read; and the render's reword
// fakes. It passes by value, so a frame a call opens in its copy closes when the call returns, and
// a read entering another category swaps the frames in its copy alone. What is read from it
// reaches the heap, a map key, an interface or a func value only as a copy; else the fakes move
// to the heap, and the allocation gate in perf_test.go fails.
type renderEnv struct {
	frames openFrames
	fakes  *rewordFakes
	row    Row
	// pick is the named pick being rendered, nil outside one; pickAt is the key in it of what renders.
	pick   *namedPick
	pickAt tree.PickKey
}

// openFrames is the frames of the name scopes rendering in one category, outermost first. No
// call changes one in place: opening a scope makes a copy one frame longer, which shares each
// outer frame's picks.
type openFrames []pickFrame

func (fs openFrames) with(scope *tree.NameScope) openFrames {
	longer := make(openFrames, len(fs)+1)
	copy(longer, fs)
	longer[len(fs)] = pickFrame{scope: scope, picks: make([]namedPick, len(scope.Order()))}
	for i, b := range scope.Order() {
		longer[len(fs)].picks[i].named = b
	}
	return longer
}

// of is the frame of scope, innermost first, nil where none is open.
func (fs openFrames) of(scope *tree.NameScope) *pickFrame {
	for i := len(fs) - 1; i >= 0; i-- {
		if fs[i].scope == scope {
			return &fs[i]
		}
	}
	return nil
}

// freshFrames is a new frame for each scope from scope out that binds a name: the frames of a
// read drawn afresh, which no later read shares.
func freshFrames(scope *tree.NameScope) openFrames {
	scope = namingScope(scope)
	if scope == nil {
		return nil
	}
	return freshFrames(scope.Up()).with(scope)
}

// namingScope is scope, or the nearest scope around it, that binds a name; nil where none does.
func namingScope(scope *tree.NameScope) *tree.NameScope {
	for scope != nil && len(scope.Order()) == 0 {
		scope = scope.Up()
	}
	return scope
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

// inScopeOf is env with a fresh frame of t's scope innermost. It is env itself where t binds no
// name, or where a frame of that scope is already open: the read entering the category opened it.
func (env renderEnv) inScopeOf(t *tree.Template) renderEnv {
	own := t.OwnScope()
	if own == nil || env.frames.of(own) != nil {
		return env
	}
	env.frames = env.frames.with(own)
	return env
}

// drawRowOf draws the row a render of t reads: inside the pick's rows where t renders as part of one.
func (env renderEnv) drawRowOf(s *drawstate.State, t *tree.Table) int {
	if env.pick != nil {
		return t.Rows().DrawIn(s, &env.pick.pins)
	}
	return t.Rows().Draw(s)
}
