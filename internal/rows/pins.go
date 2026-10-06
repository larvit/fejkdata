package rows

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/invariant"
)

type pin[O any] struct {
	t   *Table[O]
	row int
}

// Pins is the table rows fixed so far.
type Pins[O any] struct {
	inline [8]pin[O] // sized so a render over a country's five-deep geo tree stays off the heap
	used   int
	spill  map[*Table[O]]int
}

func (p *Pins[O]) Pinned(t *Table[O]) (int, bool) {
	for _, q := range p.inline[:p.used] {
		if q.t == t {
			return q.row, true
		}
	}
	r, ok := p.spill[t]
	return r, ok
}

// MustRow is the row pinned for t; the caller must have pinned one.
func (p *Pins[O]) MustRow(t *Table[O]) int {
	r, ok := p.Pinned(t)
	if !ok {
		panic(invariant.Broken("no row of %s is pinned", t.segment))
	}
	return r
}

func (p *Pins[O]) add(t *Table[O], r int) {
	if p.used < len(p.inline) {
		p.inline[p.used] = pin[O]{t, r}
		p.used++
		return
	}
	if p.spill == nil {
		p.spill = map[*Table[O]]int{}
	}
	p.spill[t] = r
}

// pin pins row r of t, and the rows of t's ancestors it links to. It checks nothing, so r must
// agree with the row p pins of nearestPinned(t): ask clash first, as PinRow does, or draw r inside
// it, as DrawIn does.
func (p *Pins[O]) pin(t *Table[O], r int) {
	for stop := p.nearestPinned(t); t != stop; t = t.parent {
		p.add(t, r)
		if t.parent != stop {
			r = t.parentRow(r)
		}
	}
}

// nearestPinned is the first of t and its ancestors p pins, where pinning a row of t stops; nil
// where none is.
func (p *Pins[O]) nearestPinned(t *Table[O]) *Table[O] {
	for ; t != nil; t = t.parent {
		if _, ok := p.Pinned(t); ok {
			return t
		}
	}
	return nil
}

// clash is the table whose pinned row keeps row r of t out: t itself pinned to another row, or the
// nearest ancestor pinned to a row r is not inside; nil where none does.
func (p *Pins[O]) clash(t *Table[O], r int) *Table[O] {
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
func (p *Pins[O]) PinRow(t *Table[O], r int) error {
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

func (p *Pins[O]) Select(t *Table[O], sel string) (int, error) {
	r, err := t.find(sel, p)
	if err != nil {
		return 0, err
	}
	return r, p.PinRow(t, r)
}

// inside keeps the rows of t that sit inside every pinned ancestor.
func (p *Pins[O]) inside(t *Table[O], rows []int) []int {
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
func (p *Pins[O]) Above(t *Table[O]) *Pins[O] {
	q := new(Pins[O])
	for ; t != nil; t = t.parent {
		if r, ok := p.Pinned(t); ok {
			q.add(t, r)
		}
	}
	return q
}
