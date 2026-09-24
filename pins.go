package fejkdata

import (
	"fmt"
	"maps"
	"sort"
	"strings"
)

// tablePin is one table's pinned row.
type tablePin struct {
	t   *table
	row int
}

// pinSet is the rows of tables fixed so far: by a render, by a path read replayed at load, and by
// a fence walk over the rows a render can enter, which also holds a row rendered whole, alone.
type pinSet struct {
	inline [8]tablePin // sized so a render over a country's five-deep geo tree stays off the heap
	n      int
	spill  map[*table]int
}

// pinned is the row pinned for t, if any.
func (p *pinSet) pinned(t *table) (int, bool) {
	for _, q := range p.inline[:p.n] {
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
		panic(fmt.Sprintf("fejkdata: a column of %s is rendered with no row pinned", t.category))
	}
	return r
}

// add pins row r of t alone.
func (p *pinSet) add(t *table, r int) {
	if p.n < len(p.inline) {
		p.inline[p.n] = tablePin{t, r}
		p.n++
		return
	}
	if p.spill == nil {
		p.spill = map[*table]int{}
	}
	p.spill[t] = r
}

// pin pins row r of t, and the rows of t's ancestors it links to.
func (p *pinSet) pin(t *table, r int) {
	for {
		if _, done := p.pinned(t); done {
			return
		}
		p.add(t, r)
		if t.parentT == nil {
			return
		}
		t, r = t.parentT, t.parentRow(r)
	}
}

// each calls fn for every pinned row, in pin order, the spilled ones by path.
func (p *pinSet) each(fn func(t *table, r int)) {
	for _, q := range p.inline[:p.n] {
		fn(q.t, q.row)
	}
	spilled := make([]*table, 0, len(p.spill))
	for t := range p.spill {
		spilled = append(spilled, t)
	}
	sort.Slice(spilled, func(i, j int) bool { return spilled[i].path < spilled[j].path })
	for _, t := range spilled {
		fn(t, p.spill[t])
	}
}

// rowOf is the row of t: the one pinned, else one s draws inside the nearest pinned ancestor — its
// parent drawn inside that first where the ancestor is further up — or over the whole table, and
// pinned with its ancestors.
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

// clone is a copy of p that pins apart from it; a plain copy shares the spill map.
func (p pinSet) clone() pinSet {
	p.spill = maps.Clone(p.spill)
	return p
}

// entered is p with row r of t: pinned where the render pins it, alone where the render draws it
// to render whole. p is left as it was, since a fence walk branches.
func (p pinSet) entered(t *table, r int, pinsAncestors bool) pinSet {
	p = p.clone()
	switch _, in := p.pinned(t); {
	case pinsAncestors:
		p.pin(t, r)
	case !in:
		p.add(t, r)
	}
	return p
}

// key spells the set for a map, by the tables' identities in path order.
func (p *pinSet) key() string {
	var pins []tablePin
	p.each(func(t *table, r int) { pins = append(pins, tablePin{t, r}) })
	sort.Slice(pins, func(i, j int) bool { return pins[i].t.path < pins[j].t.path })
	var b strings.Builder
	for _, q := range pins {
		fmt.Fprintf(&b, "%p[%d]", q.t, q.row)
	}
	return b.String()
}
