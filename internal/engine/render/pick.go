package render

import (
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
)

// placedInPick is tree.LandsInPick for a read rendering in env; with no pick in env it is false.
func placedInPick(env renderEnv, a tree.Read) (tree.PlacedRead, bool) {
	if env.pick == nil {
		return tree.PlacedRead{}, false
	}
	placed, _, kept := tree.LandsInPick(env.pick.named.Addressed(), env.pickAt, a)
	return placed, kept
}

// readUnder reads a, a read of t that env.pick keeps as placed, once per pick.
func readUnder(s *drawstate.State, t *tree.Template, env renderEnv, a tree.Read, placed tree.PlacedRead) readValue {
	p, key := env.pick, placed.Key()
	if d := p.memo.levels[key]; d.read {
		return d.value
	}
	leaf, pins := p.draw(s, t.StartOf(a.Head(), t.Refs()), a.Steps(), placed.Levels())
	return p.renderAt(s, leaf, pins, key, env)
}

// namedPick is one draw of a name: the variant drawn at each level a read of it addresses, the
// value each read there produced, by its tree.PickKey, and the table rows they pinned.
//
// Its life runs in this order:
//  1. It is created with its frame, by openFrames.with, and holds no draw yet.
//  2. It draws on the first read of the name.
//  3. Each read keeps its value by key, so a later read of that key prints the same value.
//  4. It ends with its frame, when the render that opened the frame returns. A frame the pick
//     of another category's name keeps in entered lasts as long as that pick.
type namedPick struct {
	named *tree.NameBinding
	memo  drawMemo
	pins  tree.PinSet
	// entered is the frames of the other category this pick renders, by the innermost scope a
	// read lands inside: the pick makes them on the first such read, and hands them to every
	// later one.
	entered map[*tree.NameScope]openFrames
	// drawn is what a where or a distinct clause drew, nil before the first read.
	drawn []string
}

// drawMemo is what a named pick has drawn at each level, by the level's tree.PickKey.
type drawMemo struct {
	levels map[tree.PickKey]levelDraw
}

// levelDraw is what a named pick keeps of one level.
type levelDraw struct {
	// variant is the variant the level's choice was drawn as, so every path under it reads one
	// variant; nil before that draw.
	variant tree.Node
	// value is what the read with this key produced, and read whether it has, so the same read
	// written twice reads one value.
	value readValue
	read  bool
	// steppedDown is the rows drawn by the step down at this level after a "..", nil before it.
	steppedDown *tree.PinSet
}

func (m *drawMemo) keep(level tree.PickKey, d levelDraw) {
	if m.levels == nil {
		m.levels = map[tree.PickKey]levelDraw{}
	}
	m.levels[level] = d
}

// stepDownPins is the pins a path steps down into after stepping up to t, drawn once per level
// and kept in m, so every path of the pick stepping down at that level reads one draw.
func (m *drawMemo) stepDownPins(pins *tree.PinSet, t *tree.RowsTable, level tree.PickKey) *tree.PinSet {
	d := m.levels[level]
	if d.steppedDown == nil {
		d.steppedDown = pins.Above(t)
		m.keep(level, d)
	}
	return d.steppedDown
}

// pickFrame is one render of a name scope: a pick per binding, each drawn on its first read.
type pickFrame struct {
	scope *tree.NameScope
	picks []namedPick
}

// scopeAround is the innermost name scope whose frame a read landing on n opens. A
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

// framesFrom is the frames p keeps for the scopes from scope out that bind a name. p makes a
// frame on the first call reaching its scope, so reads through p read one pick of each name.
func (p *namedPick) framesFrom(scope *tree.NameScope) openFrames {
	scope = namingScope(scope)
	if scope == nil {
		return nil
	}
	fs, ok := p.entered[scope]
	if !ok {
		fs = p.framesFrom(scope.Up()).with(scope)
		if p.entered == nil {
			p.entered = map[*tree.NameScope]openFrames{}
		}
		p.entered[scope] = fs
	}
	return fs
}

// readName reads a from the pick of its name, held in the frame of the name's scope.
func readName(s *drawstate.State, env renderEnv, a tree.Read) readValue {
	f := env.frames.of(a.Named().Scope())
	if f == nil {
		panic(invariant.Broken("name %q is read where no frame of its scope renders", a.Named().Name()))
	}
	p := &f.picks[a.Named().Index()]
	if d := p.memo.levels[a.Key()]; d.read {
		return d.value
	}
	if v := a.Named().Values(); v != nil {
		return p.readValues(s, env, v, a.Member())
	}
	return p.readPath(s, env, a)
}

// readValues reads member of the values p draws once, or all of them joined where member is 0.
func (p *namedPick) readValues(s *drawstate.State, env renderEnv, v *tree.ValueDraw, member int) readValue {
	if p.drawn == nil {
		p.drawn = drawValues(v, s, env)
	}
	if member > 0 {
		return readValue{text: p.drawn[member-1]}
	}
	return readValue{text: strings.Join(p.drawn, v.Sep())}
}

// readPath draws a's path under p and renders its leaf. A read of a field stays in the category,
// and renders in the frames the reader holds. A read of another category renders in the frames p
// keeps for it, so every read through p sees one pick of each name that category binds:
// {a.street} and {a.postal-code} in sv_SE.address read one pick of the name l that
// geo.SE.address binds.
func (p *namedPick) readPath(s *drawstate.State, env renderEnv, a tree.Read) readValue {
	leaf, pins := p.draw(s, a.Named().Target().Start(), a.Steps(), a.Levels())
	if !a.Named().BindsField() {
		env.frames = p.framesFrom(scopeAround(leaf))
	}
	return p.renderAt(s, leaf, pins, a.Key(), env)
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
	d := p.memo.levels[key]
	d.value, d.read = r, true
	p.memo.keep(key, d)
	return r
}

// variantOf is the variant of c drawn at level, drawn now where none was.
func (m *drawMemo) variantOf(s *drawstate.State, c *tree.Choice, level tree.PickKey) tree.Node {
	d := m.levels[level]
	if d.variant == nil {
		d.variant = drawThroughChoices(s, c)
		m.keep(level, d)
	}
	return d.variant
}
