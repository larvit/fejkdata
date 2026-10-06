package fejkdata

import (
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/invariant"
)

// pickKey is a level's path of segments from a root keyed "": for a named read, the node its
// name's target starts at; for a fresh read, the template it sits in. under moves a fresh read's
// key into the pick it renders in.
type pickKey string

// under is rel, a key from a fresh read's levels, moved under k, the key of what renders. It never
// returns k itself: on a render k is env.pickAt, and a memo keeping it would move every render's env to the heap.
func (k pickKey) under(rel pickKey) pickKey {
	if k == "" {
		return rel
	}
	return k + "." + rel
}

// namedPick is one draw of a name: the variant drawn at each level a read of it addresses, the
// value each read there produced, by its pickKey, and the table rows they pinned.
type namedPick struct {
	named *nameBinding
	memo  drawMemo
	pins  pinSet
}

// drawMemo is what a named pick has drawn: the variant each level was drawn as, so every path
// under it reads one variant; the value each read produced, by its pickKey, so the same read written
// twice reads one value; the frames a read opens for the name scopes around where it lands
// (enter); and the rows drawn by each step down after a "..", by level.
type drawMemo struct {
	variant       map[pickKey]node
	value         map[pickKey]readValue
	enteredFrames map[*nameScope]*pickFrame
	steppedDown   map[pickKey]*pinSet
}

// stepDownPins is the pins a path steps down into after stepping up to t: kept in m where there
// is one, so every path stepping down there reads one draw.
func (m *drawMemo) stepDownPins(pins *pinSet, t *rowsTable, levels []pickKey, at int) *pinSet {
	if m == nil {
		return pins.Above(t)
	}
	p, ok := m.steppedDown[levels[at]]
	if !ok {
		p = pins.Above(t)
		if m.steppedDown == nil {
			m.steppedDown = map[pickKey]*pinSet{}
		}
		m.steppedDown[levels[at]] = p
	}
	return p
}

// pickFrame is one render of a name scope: a pick per binding, each drawn on its first read.
type pickFrame struct {
	scope *nameScope
	picks []namedPick
}

func newPickFrame(scope *nameScope) *pickFrame {
	f := &pickFrame{scope: scope, picks: make([]namedPick, len(scope.order))}
	for i, b := range scope.order {
		f.picks[i].named = b
	}
	return f
}

// frameStack is the frames of the name scopes rendering, innermost last. A render's env points
// at it, and a repeat iteration's shares it: holding the frames in a renderEnv would move every
// render's env to the heap.
type frameStack struct {
	frames []*pickFrame
}

func (st *frameStack) push(f *pickFrame) int {
	mark := len(st.frames)
	st.frames = append(st.frames, f)
	return mark
}

func (st *frameStack) pop(mark int) { st.frames = st.frames[:mark] }

// scopeAround is the innermost name scope a render of n reads names in, short of the frames n
// renders itself, one per iteration of a repeat.
func scopeAround(n node) *nameScope {
	switch n := n.(type) {
	case *template:
		if n.repeat > 1 && n.nameScope != nil && n.nameScope.owner == n {
			return n.nameScope.up
		}
		return n.nameScope
	case *choice:
		return n.nameScope
	}
	return nil
}

func (m *drawMemo) enteredFrame(scope *nameScope) *pickFrame {
	if m == nil {
		return newPickFrame(scope)
	}
	f, ok := m.enteredFrames[scope]
	if !ok {
		f = newPickFrame(scope)
		if m.enteredFrames == nil {
			m.enteredFrames = map[*nameScope]*pickFrame{}
		}
		m.enteredFrames[scope] = f
	}
	return f
}

// readName reads the arm a from the pick of its name, held in the frame of the name's scope.
// When the pick renders another category, enter takes that category's frames from the pick's
// memo, and renderFrame finds them and opens no new one.
// So every read through one pick sees one pick of each name the other category binds:
// {a.street} and {a.postal-code} in data/sv_SE/address.json read one pick of the name l
// that data/geo/SE/address.json binds.
func readName(s *drawstate.State, env renderEnv, a arm) readValue {
	f := env.frameOf(a.named.scope)
	if f == nil {
		panic(invariant.Broken("name %q is read where no frame of its scope renders", a.named.name))
	}
	p := &f.picks[a.named.index]
	if r, done := p.memo.value[a.key()]; done {
		return r
	}
	leaf, pins := p.draw(s, a.named.start, a.steps, a.levels)
	if a.named.bindsField() {
		return p.renderAt(s, leaf, pins, a.key(), env)
	}
	env, mark := env.enter(leaf, &p.memo)
	r := p.renderAt(s, leaf, pins, a.key(), env)
	env.frames.pop(mark)
	return r
}

// readUnder reads a, an arm of t, once per pick: t renders in env.pick at env.pickAt, and a read of
// the name addresses the level a starts at. It keys the value by a's key under env.pickAt.
func readUnder(s *drawstate.State, t *template, env renderEnv, a arm) readValue {
	p, key := env.pick, env.pickAt.under(a.key())
	if r, done := p.memo.value[key]; done {
		return r
	}
	levels := make([]pickKey, len(a.levels))
	for i, l := range a.levels[:len(levels)-1] {
		levels[i] = env.pickAt.under(l)
	}
	levels[len(levels)-1] = key
	leaf, pins := p.draw(s, t.startOf(a.head), a.steps, levels)
	return p.renderAt(s, leaf, pins, key, env)
}

// draw draws the path from start under p, keeping the variant drawn at each of levels, whose last
// is the leaf's key, and returns the leaf and the pins its row is in.
func (p *namedPick) draw(s *drawstate.State, start node, steps []pathStep, levels []pickKey) (node, *pinSet) {
	leaf, pins := drawSteps(s, start, steps, &p.pins, &p.memo, levels)
	if c, isChoice := leaf.(*choice); isChoice {
		leaf = p.memo.variantOf(s, c, levels[len(levels)-1])
	}
	return leaf, pins
}

// renderAt renders leaf, drawn at key into pins, as part of p, and keeps what it rendered.
func (p *namedPick) renderAt(s *drawstate.State, leaf node, pins *pinSet, key pickKey, env renderEnv) readValue {
	env = env.at(leaf, pins)
	env.pick, env.pickAt = p, key
	r := renderLeaf(s, leaf, env)
	if p.memo.value == nil {
		p.memo.value = map[pickKey]readValue{}
	}
	p.memo.value[key] = r
	return r
}

// variantOf is the variant of c drawn at level, drawn now where none was.
func (m *drawMemo) variantOf(s *drawstate.State, c *choice, level pickKey) node {
	n, drew := m.variant[level]
	if !drew {
		n = drawThroughChoices(s, c)
		if m.variant == nil {
			m.variant = map[pickKey]node{}
		}
		m.variant[level] = n
	}
	return n
}
