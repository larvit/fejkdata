package fejkdata

import (
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// namedLevels is the levels of a read through a name, whose path runs from its target's start: the
// key of that start, "", then the key of each longer prefix of path, shortest first.
func namedLevels(path []string) []tree.PickKey {
	levels := make([]tree.PickKey, len(path)+1)
	for i := range levels {
		levels[i] = tree.PickKey(grammar.JoinSegments(path[:i]))
	}
	return levels
}

// freshLevels is the levels of a fresh read of head and segs: the key of head, then the key of each
// longer path, shortest first. It has no "" level, since the template the read sits in is already
// rendering and head is the first node the read can draw.
func freshLevels(head string, segs []string) []tree.PickKey {
	return namedLevels(append([]string{head}, segs...))[1:]
}

// addressedKeys is every key the reads of each name land on or pass, from the name's own level,
// with the spelling of the first read reaching it.
func addressedKeys(reads map[*tree.NameBinding][]*tree.Arm, l *tree.LoadFinds) map[*tree.NameBinding]map[tree.PickKey]string {
	keys := map[*tree.NameBinding]map[tree.PickKey]string{}
	for b, rs := range reads {
		keys[b] = map[tree.PickKey]string{}
		for _, r := range rs {
			for _, key := range r.Levels()[len(l.Binding(b).Target().Tail()):] {
				if _, seen := keys[b][key]; !seen {
					keys[b][key] = r.Spelling()
				}
			}
		}
	}
	return keys
}

// landsInPick reports whether a pick keeps the value of a. a is a read in a template the pick
// renders, at is that template's key, and by is the spelling of the read of the name that addresses
// a. Load and render both call it, so they agree on what a pick keeps. A reference draws afresh and
// a read through a name reads its own name's pick, so the pick keeps neither. It keeps any other
// read where a read of the name addresses a's first level, moved under at.
//
// The two kinds of read share a key because of where their levels start. A named read starts at
// the node its name's target starts at, and may draw a variant there, so its levels start at that
// node, keyed "" (namedLevels). A fresh read starts in the template it sits in, which is already
// rendering. The first node it can draw is its head, so its levels start at the head (freshLevels).
// at is that template's path from the target's start, so moved under at, a fresh read's key spells
// the path a named read's does. Where p binds a category whose format reads {first}, the {first}
// that {p} renders at "" and the {p.first} beside it both key "first", and read one draw.
func landsInPick(addressed map[tree.PickKey]string, at tree.PickKey, a tree.Arm) (by string, lands bool) {
	if grammar.IsRef(a.Head()) || a.Kind() == tree.NamedRead {
		return "", false
	}
	by, lands = addressed[at.Under(a.Levels()[0])]
	return by, lands
}

// keptInPick is landsInPick for a read rendering in env; with no pick in env it is false.
func keptInPick(env renderEnv, a tree.Arm) bool {
	if env.pick == nil {
		return false
	}
	_, kept := landsInPick(env.pick.named.Addressed(), env.pickAt, a)
	return kept
}

// readUnder reads a, an arm of t kept in env.pick, once per pick, keyed by a's key under env.pickAt.
func readUnder(s *drawstate.State, t *tree.Template, env renderEnv, a tree.Arm) readValue {
	p, key := env.pick, env.pickAt.Under(a.Key())
	if r, done := p.memo.value[key]; done {
		return r
	}
	levels := make([]tree.PickKey, len(a.Levels()))
	for i, l := range a.Levels()[:len(levels)-1] {
		levels[i] = env.pickAt.Under(l)
	}
	levels[len(levels)-1] = key // the leaf's level, already moved; moving it again allocates
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
func (m *drawMemo) stepDownPins(pins *tree.PinSet, t *tree.RowsTable, levels []tree.PickKey, at int) *tree.PinSet {
	p, ok := m.steppedDown[levels[at]]
	if !ok {
		p = pins.Above(t)
		if m.steppedDown == nil {
			m.steppedDown = map[tree.PickKey]*tree.PinSet{}
		}
		m.steppedDown[levels[at]] = p
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

// readName reads the arm a from the pick of its name, held in the frame of the name's scope.
// When the pick renders another category, enter takes that category's frames from the pick's
// memo, and renderFrame finds them and opens no new one. enter pushes those frames above mark,
// and popping to mark after the render leaves the frames as the read found them.
// So every read through one pick sees one pick of each name the other category binds:
// {a.street} and {a.postal-code} in sv_SE.address read one pick of the name l
// that geo.SE.address binds.
func readName(s *drawstate.State, env renderEnv, a tree.Arm) readValue {
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
	env, mark := env.enter(leaf, &p.memo)
	r := p.renderAt(s, leaf, pins, a.Key(), env)
	env.frames.pop(mark)
	return r
}

// draw draws the path from start under p, keeping the variant drawn at each of levels, whose last
// is the leaf's key, and returns the leaf and the pins its row is in.
func (p *namedPick) draw(s *drawstate.State, start tree.Node, steps []tree.PathStep, levels []tree.PickKey) (tree.Node, *tree.PinSet) {
	leaf, pins := drawSteps(s, drawInPick, start, steps, &p.pins, &p.memo, levels)
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
