package fejkdata

import (
	"strings"
)

// drawSet is one render's reference draws: the unnamed draw group's, and each named one's.
type drawSet struct {
	unnamed draws
	named   map[string]*draws
}

// drawScope is where a render reads its reference paths: a draw set, in the draw group of the
// template rendering, and the row a table's format is rendering, which its columns read.
type drawScope struct {
	set   *drawSet
	group string
	t     *table
	row   int
}

// newDrawSet makes the unnamed group's maps where the set is declared, keeping them on that frame's
// stack for a render that reads through them; a zero drawSet makes them on its first read instead.
func newDrawSet(s *session) drawSet {
	return drawSet{unnamed: draws{variant: map[string]node{}, value: map[string]draw{}, s: s}}
}

// renderOnce renders n as one render, over draws of its own.
func renderOnce(s *session, n node) string {
	set := drawSet{unnamed: draws{s: s}}
	return render(s, n, drawScope{set: &set})
}

// in is the scope t renders in: its draw group where it names one, else its caller's.
func (sc drawScope) in(t *template) drawScope {
	if t.drawGroupKey != "" {
		sc.group = t.drawGroupKey
	}
	return sc
}

// draws is the set's draws for sc's draw group. The session is passed in rather than
// read out of the set: copying it from there would leak the set's maps to the heap.
func (sc drawScope) draws(s *session) *draws {
	if sc.group == "" {
		return &sc.set.unnamed
	}
	d, drew := sc.set.named[sc.group]
	if !drew {
		if sc.set.named == nil {
			sc.set.named = map[string]*draws{}
		}
		d = &draws{variant: map[string]node{}, value: map[string]draw{}, s: s}
		sc.set.named[strings.Clone(sc.group)] = d // a key from sc would leak sc, and with it every render's draw set, to the heap
	}
	return d
}
