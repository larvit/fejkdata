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

func eagerHoldSet() holdSet {
	return holdSet{unnamed: hold{variant: map[string]node{}, value: map[string]draw{}}}
}

// renderOnce renders n as one render, over a hold set of its own.
func renderOnce(s *session, n node) string {
	var set holdSet
	return render(s, n, renderScope{set: &set})
}

// expandAnew expands one repeat iteration of t as a render of its own, in no group. Inlined into
// render's loop, its hold set would move to the heap.
//
//go:noinline
func expandAnew(s *session, t *template) string {
	var set holdSet
	return expand(s, t, renderScope{set: &set})
}

// in is the scope t renders in: its draw group where it names one, else its caller's.
func (sc renderScope) in(t *template) renderScope {
	if t.drawGroupKey != "" {
		sc.group = t.drawGroupKey
	}
	return sc
}

func (sc renderScope) hold() *hold {
	if sc.group == "" {
		return &sc.set.unnamed
	}
	d, drew := sc.set.named[sc.group]
	if !drew {
		if sc.set.named == nil {
			sc.set.named = map[string]*hold{}
		}
		d = &hold{variant: map[string]node{}, value: map[string]draw{}}
		sc.set.named[strings.Clone(sc.group)] = d // a key from sc would leak sc, and with it every render's hold set, to the heap
	}
	return d
}
