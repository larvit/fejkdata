package fejkdata

import (
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/invariant"
)

// namedPick is one draw of a name: the variant drawn at each level a read of it addresses, the
// value each read there produced, keyed by the path from the name, and the table rows they
// pinned.
type namedPick struct {
	named *nameBinding
	memo  drawMemo
	pins  pinSet
}

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
func (m *drawMemo) stepDownPins(pins *pinSet, t *rowsTable, levels []string, at int) *pinSet {
	if m == nil {
		return pins.Above(t)
	}
	p, ok := m.steppedDown[levels[at]]
	if !ok {
		p = pins.Above(t)
		if m.steppedDown == nil {
			m.steppedDown = map[string]*pinSet{}
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

// frameStack is the frames of the name scopes rendering, innermost last. A render's scope points
// at it, and a repeat iteration's shares it: holding the frames in a renderScope would move every
// render's scope to the heap.
type frameStack struct {
	frames []*pickFrame
}

// push opens f until pop closes it, returning the mark pop takes.
func (st *frameStack) push(f *pickFrame) int {
	mark := len(st.frames)
	st.frames = append(st.frames, f)
	return mark
}

func (st *frameStack) pop(mark int) { st.frames = st.frames[:mark] }

// enter starts a read landing on n: the read sees no frame opened before it, and gets a frame for
// each name scope around n, from memo where there is one, so reads sharing memo read one pick of
// each name. It returns the mark that closes those frames.
func (sc renderScope) enter(n node, memo *drawMemo) (renderScope, int) {
	sc = sc.entering()
	for scope := scopeAround(n); scope != nil; scope = scope.up {
		if len(scope.order) > 0 {
			sc.frames.push(memo.enteredFrame(scope))
		}
	}
	return sc, sc.base
}

// entering is sc as a read entering a category sees it: no frame opened before it, and no pick.
func (sc renderScope) entering() renderScope {
	sc.base, sc.pick = len(sc.frames.frames), nil
	return sc
}

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

// renderFrame opens a fresh frame of t's scope, where t binds names and no read entering the
// category opened one, returning the mark that closes it, or -1.
func (sc renderScope) renderFrame(t *template) int {
	own := t.ownScope()
	if own == nil || sc.frameOf(own) != nil {
		return -1
	}
	return sc.frames.push(newPickFrame(own))
}

// frameOf is the frame of scope rendering since the read entering the category, nil where none is.
func (sc renderScope) frameOf(scope *nameScope) *pickFrame {
	for i := len(sc.frames.frames) - 1; i >= sc.base; i-- {
		if f := sc.frames.frames[i]; f.scope == scope {
			return f
		}
	}
	return nil
}

// readName reads a from its name's pick in the frame rendering the name's scope. Every read of
// one pick stays on it: the pick keeps each path's value and variant in its memo and its rows in
// its pins. A read entering another category takes the frames of the name scopes there from the
// memo too (enter), and renderFrame finds them and opens none. So {a.street} and {a.postal-code} in
// data/sv_SE/address.json read one pick of the name l that data/geo/SE/address.json binds.
func readName(s *drawstate.State, sc renderScope, a arm) readValue {
	f := sc.frameOf(a.named.scope)
	if f == nil {
		panic(invariant.Broken("name %q is read where no frame of its scope renders", a.named.name))
	}
	p := &f.picks[a.named.index]
	if r, done := p.memo.value[a.path]; done {
		return r
	}
	leaf, pins := p.draw(s, a.named.head, a.steps, a.levels, a.path)
	if a.named.bindsField() {
		return p.renderAt(s, leaf, pins, a.path, sc)
	}
	sc, mark := sc.enter(leaf, &p.memo)
	r := p.renderAt(s, leaf, pins, a.path, sc)
	sc.frames.pop(mark)
	return r
}

// drawRowOf draws the row a render of t reads: inside the pick's rows where t renders as part of one.
func (sc renderScope) drawRowOf(s *drawstate.State, t *table) int {
	if sc.pick != nil {
		return t.rows.DrawIn(s, &sc.pick.pins)
	}
	return t.rows.Draw(s)
}

// keeps reports whether a read of sc's name addresses the level a starts at.
func (sc renderScope) keeps(a arm) bool {
	_, kept := sc.pick.named.addressed[underKey(sc.pickKey, a.head)]
	return kept
}

// readUnder reads a of t, which renders as part of the pick sc.pick at sc.pickKey, where a read of
// the name addresses it: once per pick, by its path from the name.
func readUnder(s *drawstate.State, t *template, sc renderScope, a arm) readValue {
	p, key := sc.pick, underKey(sc.pickKey, a.path)
	if r, done := p.memo.value[key]; done {
		return r
	}
	levels := make([]string, len(a.levels))
	for i, l := range a.levels {
		levels[i] = underKey(sc.pickKey, l)
	}
	leaf, pins := p.draw(s, t.head(a.head), a.steps, levels, key)
	return p.renderAt(s, leaf, pins, key, sc)
}

// underKey is the key of path under the pick key prefix. It never returns prefix itself: prefix is
// sc.pickKey, and a memo keeping it would move every render's scope to the heap.
func underKey(prefix, path string) string {
	if prefix == "" {
		return path
	}
	return prefix + "." + path
}

// draw draws the path key names under p, its variant at key kept too, returning the leaf and the
// pins its row is in.
func (p *namedPick) draw(s *drawstate.State, head node, steps []pathStep, levels []string, key string) (node, *pinSet) {
	leaf, pins := drawSteps(s, head, steps, &p.pins, &p.memo, levels)
	if c, isChoice := leaf.(*choice); isChoice {
		leaf = p.memo.variantOf(s, c, key)
	}
	return leaf, pins
}

// renderAt renders leaf, drawn at key into pins, as part of p, and keeps what it rendered.
func (p *namedPick) renderAt(s *drawstate.State, leaf node, pins *pinSet, key string, sc renderScope) readValue {
	sc = sc.at(leaf, pins)
	sc.pick, sc.pickKey = p, key
	r := renderLeaf(s, leaf, sc)
	if p.memo.value == nil {
		p.memo.value = map[string]readValue{}
	}
	p.memo.value[key] = r
	return r
}

// variantOf is the variant of c drawn at level, drawn now where none was.
func (m *drawMemo) variantOf(s *drawstate.State, c *choice, level string) node {
	n, drew := m.variant[level]
	if !drew {
		n = resolveChoice(s, c)
		if m.variant == nil {
			m.variant = map[string]node{}
		}
		m.variant[level] = n
	}
	return n
}
