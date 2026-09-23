package fejkdata

import (
	"fmt"
	"maps"
	"sort"
)

// tablePin is one table's pinned row.
type tablePin struct {
	t   *table
	row int
}

// pinSet is the table rows a draw has fixed: a render's, one per draw group, and the fences', which
// replay reads into one to prove what the render would pin.
type pinSet struct {
	pins  [8]tablePin // inline, so a render over a country's five-deep geo tree stays off the heap
	npins int
	more  map[*table]int // the rows pinned past the inline eight
}

// pinned is the row pinned for t, if any.
func (p *pinSet) pinned(t *table) (int, bool) {
	for _, tp := range p.pins[:p.npins] {
		if tp.t == t {
			return tp.row, true
		}
	}
	r, ok := p.more[t]
	return r, ok
}

// mustRow is the row pinned for t, which the walk reaching a column pinned.
func (p *pinSet) mustRow(t *table) int {
	r, ok := p.pinned(t)
	if !ok {
		panic(fmt.Sprintf("fejkdata: a column of %s is rendered with no row pinned", t.category))
	}
	return r
}

func (p *pinSet) set(t *table, r int) {
	if p.npins < len(p.pins) {
		p.pins[p.npins] = tablePin{t, r}
		p.npins++
		return
	}
	if p.more == nil {
		p.more = map[*table]int{}
	}
	p.more[t] = r
}

// pin pins row r of t, and the rows of t's ancestors it links to.
func (p *pinSet) pin(t *table, r int) {
	for {
		if _, done := p.pinned(t); done {
			return
		}
		p.set(t, r)
		if t.parentT == nil {
			return
		}
		t, r = t.parentT, t.parentRow(r)
	}
}

// each calls fn for every pinned row, in pin order, the spilled ones by table name.
func (p *pinSet) each(fn func(t *table, r int)) {
	for _, tp := range p.pins[:p.npins] {
		fn(tp.t, tp.row)
	}
	spilled := make([]*table, 0, len(p.more))
	for t := range p.more {
		spilled = append(spilled, t)
	}
	sort.Slice(spilled, func(i, j int) bool { return spilled[i].category < spilled[j].category })
	for _, t := range spilled {
		fn(t, p.more[t])
	}
}

// rowOf is the render's row of t: the one pinned, else one s draws inside the nearest pinned
// ancestor — its parent drawn inside that first where the ancestor is further up — or over the whole
// table, and pinned with its ancestors.
func (p *pinSet) rowOf(s *session, t *table) int {
	if r, ok := p.pinned(t); ok {
		return r
	}
	r := -1
	for a := t.parentT; a != nil && r < 0; a = a.parentT {
		if _, ok := p.pinned(a); !ok {
			continue
		}
		if t.parentT != a {
			p.rowOf(s, t.parentT)
		}
		pr, _ := p.pinned(t.parentT)
		r = t.drawUnder(s, pr)
	}
	if r < 0 {
		r = t.draw(s)
	}
	p.pin(t, r)
	return r
}

// pinRow pins row r of t where it agrees with the rows pinned before it.
func (p *pinSet) pinRow(t *table, r int) error {
	if pr, ok := p.pinned(t); ok && pr != r {
		return fmt.Errorf("%s and %s are two rows of %s", t.selectorSpelling(pr), t.selectorSpelling(r), t.path)
	}
	for a := t.parentT; a != nil; a = a.parentT {
		if pa, ok := p.pinned(a); ok && !t.under(r, a, pa) {
			return fmt.Errorf("%s is not inside %s", t.selectorSpelling(r), a.selectorSpelling(pa))
		}
	}
	p.pin(t, r)
	return nil
}

// selectRow pins the row a selector names.
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

// enter is p with row r of t: pinned where the render pins that row, else set alone, as a row a
// render draws whole. The set is copied, since walks branch.
func (p pinSet) enter(t *table, r int, pins bool) pinSet {
	p.more = maps.Clone(p.more)
	if pins {
		p.pin(t, r)
	} else if _, in := p.pinned(t); !in {
		p.set(t, r)
	}
	return p
}
