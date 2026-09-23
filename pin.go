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

// pinSet is the rows pinned: a render's, per draw group, and the draw fences' as they walk
// what a render draws.
type pinSet struct {
	pins  [8]tablePin // inline so a render over a country's five-deep geo tree stays off the heap
	npins int
	more  map[*table]int // the rows pinned past the inline eight
}

// pinned is the row pinned for t, if any.
func (p *pinSet) pinned(t *table) (int, bool) {
	for _, pin := range p.pins[:p.npins] {
		if pin.t == t {
			return pin.row, true
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

// pin pins row r of t, and the rows of t's ancestors it links to.
func (p *pinSet) pin(t *table, r int) {
	for p.pinOne(t, r) && t.parentT != nil {
		t, r = t.parentT, t.parentRow(r)
	}
}

// pinOne pins row r of t alone, and reports whether t was unpinned before.
func (p *pinSet) pinOne(t *table, r int) bool {
	if _, done := p.pinned(t); done {
		return false
	}
	if p.npins < len(p.pins) {
		p.pins[p.npins] = tablePin{t, r}
		p.npins++
		return true
	}
	if p.more == nil {
		p.more = map[*table]int{}
	}
	p.more[t] = r
	return true
}

// each calls fn for every pinned row, in pin order, the spilled ones by table name.
func (p *pinSet) each(fn func(t *table, r int)) {
	for _, pin := range p.pins[:p.npins] {
		fn(pin.t, pin.row)
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

// refusal is why row r of t cannot render with the rows pinned: another row of t is,
// or an ancestor's row it sits outside; nil where it agrees with them.
func (p *pinSet) refusal(t *table, r int) error {
	if pr, ok := p.pinned(t); ok && pr != r {
		return fmt.Errorf("%s and %s are two rows of %s", t.selectorSpelling(pr), t.selectorSpelling(r), t.path)
	}
	for a := t.parentT; a != nil; a = a.parentT {
		if pa, ok := p.pinned(a); ok && !t.under(r, a, pa) {
			return fmt.Errorf("%s is not inside %s", t.selectorSpelling(r), a.selectorSpelling(pa))
		}
	}
	return nil
}

// pinRow pins row r of t where it agrees with the rows pinned before it.
func (p *pinSet) pinRow(t *table, r int) error {
	if err := p.refusal(t, r); err != nil {
		return err
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

// entered is a copy of p with row r of t pinned, with its ancestors where the render pins
// the row rather than drawing it; a fence's walk branches, so the set it walked from stays.
func (p pinSet) entered(t *table, r int, ancestors bool) pinSet {
	p.more = maps.Clone(p.more)
	if ancestors {
		p.pin(t, r)
	} else {
		p.pinOne(t, r)
	}
	return p
}

// key spells the set for a map, the same whichever order pinned it.
func (p *pinSet) key() string {
	pins := make([]tablePin, 0, p.npins+len(p.more))
	p.each(func(t *table, r int) { pins = append(pins, tablePin{t, r}) })
	sort.Slice(pins, func(i, j int) bool { return pins[i].t.path < pins[j].t.path })
	var b strings.Builder
	for _, pin := range pins {
		fmt.Fprintf(&b, "%p[%d]", pin.t, pin.row)
	}
	return b.String()
}
