package fejkdata

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/invariant"
)

type tablePin struct {
	t   *table
	row int
}

// pinSet is the table rows fixed so far.
type pinSet struct {
	inline [8]tablePin // sized so a render over a country's five-deep geo tree stays off the heap
	used   int
	spill  map[*table]int
}

func (p *pinSet) pinned(t *table) (int, bool) {
	for _, q := range p.inline[:p.used] {
		if q.t == t {
			return q.row, true
		}
	}
	r, ok := p.spill[t]
	return r, ok
}

// mustRow is the row pinned for t, which the walk reaching a column pinned.
func (p *pinSet) mustRow(t *table) int {
	r, ok := p.pinned(t)
	if !ok {
		panic(invariant.Broken("a column of %s is rendered with no row pinned", t.segment))
	}
	return r
}

func (p *pinSet) add(t *table, r int) {
	if p.used < len(p.inline) {
		p.inline[p.used] = tablePin{t, r}
		p.used++
		return
	}
	if p.spill == nil {
		p.spill = map[*table]int{}
	}
	p.spill[t] = r
}

// pin pins row r of t, and the rows of t's ancestors it links to. It checks nothing, so r must
// agree with the row p pins of nearestPinned(t): ask clash first, as pinRow does, or draw r inside
// it, as drawIn does.
func (p *pinSet) pin(t *table, r int) {
	for stop := p.nearestPinned(t); t != stop; t = t.parentT {
		p.add(t, r)
		if t.parentT != stop {
			r = t.parentRow(r)
		}
	}
}

// nearestPinned is the first of t and its ancestors p pins, where pinning a row of t stops; nil
// where none is.
func (p *pinSet) nearestPinned(t *table) *table {
	for ; t != nil; t = t.parentT {
		if _, ok := p.pinned(t); ok {
			return t
		}
	}
	return nil
}

// clash is the table whose pinned row keeps row r of t out: t itself pinned to another row, or the
// nearest ancestor pinned to a row r is not inside; nil where none does.
func (p *pinSet) clash(t *table, r int) *table {
	if pr, ok := p.pinned(t); ok && pr != r {
		return t
	}
	for a := t.parentT; a != nil; a = a.parentT {
		if pa, ok := p.pinned(a); ok && !t.under(r, a, pa) {
			return a
		}
	}
	return nil
}

// pinRow pins row r of t where it agrees with the rows pinned before it.
func (p *pinSet) pinRow(t *table, r int) error {
	switch a := p.clash(t, r); {
	case a == t:
		pr, _ := p.pinned(t)
		return fmt.Errorf("%s and %s are two rows of %s", t.selectorSpelling(pr), t.selectorSpelling(r), t.path)
	case a != nil:
		pa, _ := p.pinned(a)
		return fmt.Errorf("%s is not inside %s", t.selectorSpelling(r), a.selectorSpelling(pa))
	}
	p.pin(t, r)
	return nil
}

func (p *pinSet) selectRow(t *table, sel string) error {
	r, err := t.find(sel, p)
	if err != nil {
		return err
	}
	return p.pinRow(t, r)
}

// inside keeps the rows of t that sit inside every pinned ancestor.
func (p *pinSet) inside(t *table, rows []int) []int {
	for a := t.parentT; a != nil; a = a.parentT {
		pa, ok := p.pinned(a)
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

// above is a pin set holding the rows p pins of t and its ancestors, and none below them.
func (p *pinSet) above(t *table) *pinSet {
	q := new(pinSet)
	for ; t != nil; t = t.parentT {
		if r, ok := p.pinned(t); ok {
			q.add(t, r)
		}
	}
	return q
}
