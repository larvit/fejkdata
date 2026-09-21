package fejkdata

import (
	"strings"
)

// holdSet is one render's reference draws: the unnamed draw group's, and each named one's.
type holdSet struct {
	unnamed hold
	named   map[string]*hold
}

// renderScope is where a render reads its reference paths: a hold set, in the draw group of the
// template rendering, and the row a table's format is rendering, which its columns read.
type renderScope struct {
	set   *holdSet
	group string
	t     *table
	row   int
}

// newHoldSet makes the unnamed group's maps where the set is declared, keeping them on that frame's
// stack for a render that reads through them; a zero holdSet makes them on its first read instead.
func newHoldSet(s *session) holdSet {
	return holdSet{unnamed: hold{variant: map[string]node{}, value: map[string]draw{}, s: s}}
}

// renderOnce renders n as one render, over a hold set of its own.
func renderOnce(s *session, n node) string {
	set := holdSet{unnamed: hold{s: s}}
	return render(s, n, renderScope{set: &set})
}

// in is the scope t renders in: its draw group where it names one, else its caller's.
func (sc renderScope) in(t *template) renderScope {
	if t.drawGroupKey != "" {
		sc.group = t.drawGroupKey
	}
	return sc
}

// hold is the set's hold for sc's draw group. The session is passed in rather than
// read out of the set: copying it from there would leak the set's maps to the heap.
func (sc renderScope) hold(s *session) *hold {
	if sc.group == "" {
		return &sc.set.unnamed
	}
	d, drew := sc.set.named[sc.group]
	if !drew {
		if sc.set.named == nil {
			sc.set.named = map[string]*hold{}
		}
		d = &hold{variant: map[string]node{}, value: map[string]draw{}, s: s}
		sc.set.named[strings.Clone(sc.group)] = d // a key from sc would leak sc, and with it every render's hold set, to the heap
	}
	return d
}
