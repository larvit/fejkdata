package rows

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/invariant"
)

type pin[P any] struct {
	t   *Table[P]
	row int
}

// Pins is the table rows fixed so far.
type Pins[P any] struct {
	inline [8]pin[P] // sized so a render over a country's five-deep geo tree stays off the heap
	used   int
	spill  map[*Table[P]]int
}

func (p *Pins[P]) Pinned(t *Table[P]) (int, bool) {
	for _, q := range p.inline[:p.used] {
		if q.t == t {
			return q.row, true
		}
	}
	r, ok := p.spill[t]
	return r, ok
}

// MustRow is the row pinned for t, which the walk reaching a column pinned.
func (p *Pins[P]) MustRow(t *Table[P]) int {
	r, ok := p.Pinned(t)
	if !ok {
		panic(invariant.Broken("a column of %s is rendered with no row pinned", t.segment))
	}
	return r
}

func (p *Pins[P]) add(t *Table[P], r int) {
	if p.used < len(p.inline) {
		p.inline[p.used] = pin[P]{t, r}
		p.used++
		return
	}
	if p.spill == nil {
		p.spill = map[*Table[P]]int{}
	}
	p.spill[t] = r
}

// pin pins row r of t, and the rows of t's ancestors it links to. It checks nothing, so r must
// agree with the row p pins of nearestPinned(t): ask clash first, as PinRow does, or draw r inside
// it, as DrawIn does.
func (p *Pins[P]) pin(t *Table[P], r int) {
	for stop := p.nearestPinned(t); t != stop; t = t.parent {
		p.add(t, r)
		if t.parent != stop {
			r = t.parentRow(r)
		}
	}
}

// nearestPinned is the first of t and its ancestors p pins, where pinning a row of t stops; nil
// where none is.
func (p *Pins[P]) nearestPinned(t *Table[P]) *Table[P] {
	for ; t != nil; t = t.parent {
		if _, ok := p.Pinned(t); ok {
			return t
		}
	}
	return nil
}

// clash is the table whose pinned row keeps row r of t out: t itself pinned to another row, or the
// nearest ancestor pinned to a row r is not inside; nil where none does.
func (p *Pins[P]) clash(t *Table[P], r int) *Table[P] {
	if pr, ok := p.Pinned(t); ok && pr != r {
		return t
	}
	for a := t.parent; a != nil; a = a.parent {
		if pa, ok := p.Pinned(a); ok && !t.under(r, a, pa) {
			return a
		}
	}
	return nil
}

// PinRow pins row r of t where it agrees with the rows pinned before it.
func (p *Pins[P]) PinRow(t *Table[P], r int) error {
	switch a := p.clash(t, r); {
	case a == t:
		pr, _ := p.Pinned(t)
		return fmt.Errorf("%s and %s are two rows of %s", t.selectorSpelling(pr), t.selectorSpelling(r), t.path)
	case a != nil:
		pa, _ := p.Pinned(a)
		return fmt.Errorf("%s is not inside %s", t.selectorSpelling(r), a.selectorSpelling(pa))
	}
	p.pin(t, r)
	return nil
}

// Select pins the row of t that sel names, and returns it.
func (p *Pins[P]) Select(t *Table[P], sel string) (int, error) {
	r, err := t.find(sel, p)
	if err != nil {
		return 0, err
	}
	return r, p.PinRow(t, r)
}

// inside keeps the rows of t that sit inside every pinned ancestor.
func (p *Pins[P]) inside(t *Table[P], rows []int) []int {
	for a := t.parent; a != nil; a = a.parent {
		pa, ok := p.Pinned(a)
		if !ok {
			continue
		}
		var kept []int
		for _, r := range rows {
			if t.under(r, a, pa) {
				kept = append(kept, r)
			}
		}
		rows = kept
	}
	return rows
}

// Above is a pin set holding the rows p pins of t and its ancestors, and none below them.
func (p *Pins[P]) Above(t *Table[P]) *Pins[P] {
	q := new(Pins[P])
	for ; t != nil; t = t.parent {
		if r, ok := p.Pinned(t); ok {
			q.add(t, r)
		}
	}
	return q
}
