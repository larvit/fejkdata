package fejkdata

import (
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// pickKey is the key a named pick's memo keeps one level's draw under.
//
// A level is one prefix of a read's path, and a read has one for each node it passes; a choice
// and the variant drawn from it, or a table and its row, share one. A level's key is its path from
// where the read starts, and landsInPick says where that is. Load gives each read one key per
// level, in levelKeys.
//
// A name's addressed keys are the keys some read of it lands on or passes, from the level its
// binding lands on onward; load gathers them in addressedKeys. When a template a pick renders
// reads a field, the pick keeps that read's value only if the read's first-level key, put under
// the rendering template's key, is one of the name's addressed keys. landsInPick decides this,
// and readUnder reads the kept value.
//
// In geo.SE.address, {.locality as l} binds l, and {l.street.name} passes the levels "", "street"
// and "street.name", the last of which is its key. With {l.name} and {l.postal-code.code}, l's
// addressed keys are those three, "name", "postal-code" and "postal-code.code".
type pickKey string

// under puts rel beneath k: k + "." + rel, or rel alone where k is empty. So pickKey("a").under("b")
// is "a.b". Here k is the key of what renders and rel a key from a fresh read's levels.
// It never returns k itself: on a render k is env.pickAt, and a memo keeping it would move every
// render's env to the heap.
func (k pickKey) under(rel pickKey) pickKey {
	if k == "" {
		return rel
	}
	return k + "." + rel
}

// levelKeys is the key of each prefix of path holding at least from segments, shortest first.
func levelKeys(path []string, from int) []pickKey {
	levels := make([]pickKey, len(path)-from+1)
	for i := range levels {
		levels[i] = pickKey(grammar.JoinSegments(path[:from+i]))
	}
	return levels
}

// addressedKeys is every key the reads of each name in ts land on or pass, from the name's own
// level, with the spelling of the first read reaching it.
func addressedKeys(ts []templateSite, targets map[*nameBinding]nameTarget) map[*nameBinding]map[pickKey]string {
	keys := map[*nameBinding]map[pickKey]string{}
	for _, s := range ts {
		for _, r := range namedReads(s.t) {
			b := r.a.named
			if keys[b] == nil {
				keys[b] = map[pickKey]string{}
			}
			for _, key := range r.a.levels[len(targets[b].tail):] {
				if _, seen := keys[b][key]; !seen {
					keys[b][key] = r.a.spelling
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
// node, keyed "" (compileArm). A fresh read starts in the template it sits in, which is already
// rendering. The first node it can draw is its head, so its levels start at the head (pathArm).
// at is that template's path from the target's start, so moved under at, a fresh read's key spells
// the path a named read's does. Where p binds a category whose format reads {first}, the {first}
// that {p} renders at "" and the {p.first} beside it both key "first", and read one draw.
func landsInPick(addressed map[pickKey]string, at pickKey, a arm) (by string, lands bool) {
	if grammar.IsRef(a.head) || a.kind == namedRead {
		return "", false
	}
	by, lands = addressed[at.under(a.levels[0])]
	return by, lands
}

// keptInPick is landsInPick for a read rendering in env; with no pick in env it is false.
func keptInPick(env renderEnv, a arm) bool {
	if env.pick == nil {
		return false
	}
	_, kept := landsInPick(env.pick.named.addressed, env.pickAt, a)
	return kept
}

// readUnder reads a, an arm of t kept in env.pick, once per pick, keyed by a's key under env.pickAt.
func readUnder(s *drawstate.State, t *template, env renderEnv, a arm) readValue {
	p, key := env.pick, env.pickAt.under(a.key())
	if r, done := p.memo.value[key]; done {
		return r
	}
	levels := make([]pickKey, len(a.levels))
	for i, l := range a.levels[:len(levels)-1] {
		levels[i] = env.pickAt.under(l)
	}
	levels[len(levels)-1] = key // the leaf's level, already moved; moving it again allocates
	leaf, pins := p.draw(s, t.startOf(a.head), a.steps, levels)
	return p.renderAt(s, leaf, pins, key, env)
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

// scopeAround is the innermost name scope a read landing on n looks names up in. A repeated
// template that owns its scope opens that scope's frame itself, once per iteration, so for it this
// is the scope around the repeat.
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
// {a.street} and {a.postal-code} in sv_SE.address read one pick of the name l
// that geo.SE.address binds.
func readName(s *drawstate.State, env renderEnv, a arm) readValue {
	f := env.frameOf(a.named.scope)
	if f == nil {
		panic(invariant.Broken("name %q is read where no frame of its scope renders", a.named.name))
	}
	p := &f.picks[a.named.index]
	if r, done := p.memo.value[a.key()]; done {
		return r
	}
	leaf, pins := p.draw(s, a.named.target.start, a.steps, a.levels)
	if a.named.bindsField() {
		return p.renderAt(s, leaf, pins, a.key(), env)
	}
	env, mark := env.enter(leaf, &p.memo)
	r := p.renderAt(s, leaf, pins, a.key(), env)
	env.frames.pop(mark)
	return r
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
