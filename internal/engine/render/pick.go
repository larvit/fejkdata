package render

import (
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
)

// keptInPick is tree.LandsInPick for a read rendering in env; with no pick in env it is false.
func keptInPick(env renderEnv, a tree.Read) bool {
	if env.pick == nil {
		return false
	}
	_, kept := tree.LandsInPick(env.pick.named.Addressed(), env.pickAt, a)
	return kept
}

// readUnder reads a, a read of t kept in env.pick, once per pick, keyed by a's key under env.pickAt.
func readUnder(s *drawstate.State, t *tree.Template, env renderEnv, a tree.Read) readValue {
	p, key := env.pick, env.pickAt.Under(a.Key()) // the last of LevelsUnder, alone: a hit needs no other
	if r, done := p.memo.value[key]; done {
		return r
	}
	levels := tree.LevelsUnder(env.pickAt, a)
	leaf, pins := p.draw(s, t.StartOf(a.Head(), t.Refs()), a.Steps(), levels)
	return p.renderAt(s, leaf, pins, key, env)
}

// namedPick is one draw of a name: the variant drawn at each level a read of it addresses, the
// value each read there produced, by its tree.PickKey, and the table rows they pinned.
//
// Its life runs in this order:
//  1. It is created with its frame, by newPickFrame, and holds no draw yet.
//  2. It draws on the first read of the name.
//  3. Each read keeps its value by key, so a later read of that key prints the same value.
//  4. It ends when its frame pops. A frame kept in another pick's memo (enter) lasts as long as that pick.
type namedPick struct {
	named *tree.NameBinding
	memo  drawMemo
	pins  tree.PinSet
	// drawn is what a where or a distinct clause drew, nil before the first read.
	drawn []string
}

// drawMemo is what a named pick has drawn: the variant each level was drawn as, so every path
// under it reads one variant; the value each read produced, by its tree.PickKey, so the same read written
// twice reads one value; the frames a read opens for the name scopes around where it lands
// (enter); and the rows drawn by each step down after a "..", by level.
type drawMemo struct {
	variant       map[tree.PickKey]tree.Node
	value         map[tree.PickKey]readValue
	enteredFrames map[*tree.NameScope]*pickFrame
	steppedDown   map[tree.PickKey]*tree.PinSet
}

// stepDownPins is the pins a path steps down into after stepping up to t, drawn once per level
// and kept in m, so every path of the pick stepping down at that level reads one draw.
func (m *drawMemo) stepDownPins(pins *tree.PinSet, t *tree.RowsTable, level tree.PickKey) *tree.PinSet {
	p, ok := m.steppedDown[level]
	if !ok {
		p = pins.Above(t)
		if m.steppedDown == nil {
			m.steppedDown = map[tree.PickKey]*tree.PinSet{}
		}
		m.steppedDown[level] = p
	}
	return p
}

// pickFrame is one render of a name scope: a pick per binding, each drawn on its first read.
type pickFrame struct {
	scope *tree.NameScope
	picks []namedPick
}

func newPickFrame(scope *tree.NameScope) *pickFrame {
	f := &pickFrame{scope: scope, picks: make([]namedPick, len(scope.Order()))}
	for i, b := range scope.Order() {
		f.picks[i].named = b
	}
	return f
}

// scopeAround is the innermost name scope whose frame enter opens for a read landing on n. A
// repeated template that owns its scope opens that scope's frame itself, once per iteration, so
// for it this is the scope around the repeat.
func scopeAround(n tree.Node) *tree.NameScope {
	switch n := n.(type) {
	case *tree.Template:
		if n.Repeat() > 1 && n.NameScope() != nil && n.NameScope().Owner() == n {
			return n.NameScope().Up()
		}
		return n.NameScope()
	case *tree.Choice:
		return n.NameScope()
	}
	return nil
}

func (m *drawMemo) enteredFrame(scope *tree.NameScope) *pickFrame {
	if m == nil {
		return newPickFrame(scope)
	}
	f, ok := m.enteredFrames[scope]
	if !ok {
		f = newPickFrame(scope)
		if m.enteredFrames == nil {
			m.enteredFrames = map[*tree.NameScope]*pickFrame{}
		}
		m.enteredFrames[scope] = f
	}
	return f
}

// readName reads a from the pick of its name, held in the frame of the name's scope.
// When the pick renders another category, env.frames.enter pushes that category's frames above
// mark. The pick's memo keeps those frames. openOwn finds them and opens no new one. Popping to
// mark after the render leaves the frames as the read found them.
// So every read through one pick sees one pick of each name the other category binds:
// {a.street} and {a.postal-code} in sv_SE.address read one pick of the name l
// that geo.SE.address binds.
func readName(s *drawstate.State, env renderEnv, a tree.Read) readValue {
	f := env.frameOf(a.Named().Scope())
	if f == nil {
		panic(invariant.Broken("name %q is read where no frame of its scope renders", a.Named().Name()))
	}
	p := &f.picks[a.Named().Index()]
	if r, done := p.memo.value[a.Key()]; done {
		return r
	}
	if v := a.Named().Values(); v != nil {
		if p.drawn == nil {
			p.drawn = drawValues(v, s, env)
		}
		if a.Member() > 0 {
			return readValue{text: p.drawn[a.Member()-1]}
		}
		return readValue{text: strings.Join(p.drawn, v.Sep())}
	}
	leaf, pins := p.draw(s, a.Named().Target().Start(), a.Steps(), a.Levels())
	if a.Named().BindsField() { // the read stays in the category, so its frames are already open
		return p.renderAt(s, leaf, pins, a.Key(), env)
	}
	mark := env.frames.enter(leaf, &p.memo)
	r := p.renderAt(s, leaf, pins, a.Key(), env.entered(mark))
	env.frames.pop(mark)
	return r
}

// draw draws the path from start under p, keeping the variant drawn at each of levels, whose last
// is the leaf's key, and returns the leaf and the pins its row is in.
func (p *namedPick) draw(s *drawstate.State, start tree.Node, steps []tree.PathStep, levels []tree.PickKey) (tree.Node, *tree.PinSet) {
	leaf, pins := drawAcrossTables(s, start, steps, &p.pins, &keptDraws{memo: &p.memo, levels: levels})
	if c, isChoice := leaf.(*tree.Choice); isChoice {
		leaf = p.memo.variantOf(s, c, levels[len(levels)-1])
	}
	return leaf, pins
}

// renderAt renders leaf, drawn at key into pins, as part of p, and keeps what it rendered.
func (p *namedPick) renderAt(s *drawstate.State, leaf tree.Node, pins *tree.PinSet, key tree.PickKey, env renderEnv) readValue {
	env = env.at(leaf, pins)
	env.pick, env.pickAt = p, key
	r := renderLeaf(s, leaf, env)
	if p.memo.value == nil {
		p.memo.value = map[tree.PickKey]readValue{}
	}
	p.memo.value[key] = r
	return r
}

// variantOf is the variant of c drawn at level, drawn now where none was.
func (m *drawMemo) variantOf(s *drawstate.State, c *tree.Choice, level tree.PickKey) tree.Node {
	n, drew := m.variant[level]
	if !drew {
		n = drawThroughChoices(s, c)
		if m.variant == nil {
			m.variant = map[tree.PickKey]tree.Node{}
		}
		m.variant[level] = n
	}
	return n
}
