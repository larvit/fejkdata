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

// hold is what has already been drawn for held names: the variant each was drawn as,
// so every path under it reads one row, and the draw each read made, by its one
// spelling, so the same read written twice reads one draw. An expansion keeps one for
// its sibling names, and a render one per group for its reference paths.
type hold struct {
	pinSet
	variant map[string]node
	value   map[string]draw
}

// draw is what one read drew: its text, and whether it landed on a null.
type draw struct {
	text string
	null bool
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

// rowOf is the render's row of t: the one pinned, else one s draws inside the
// nearest pinned ancestor — its parent drawn inside that first where the ancestor
// is further up — or over the whole table, and pinned with its ancestors.
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

// disagrees is why row r of t cannot render with the rows pinned: another row of t is,
// or an ancestor's row it sits outside; nil where it agrees with them.
func (p *pinSet) disagrees(t *table, r int) error {
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
	if err := p.disagrees(t, r); err != nil {
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

// readField renders one arm of a token. An arm's key is a sibling field or a
// reference linkRefs bound into fields. A name the expansion holds — a level some
// token addresses by dotted path, or a field an operand reads — is drawn once and
// kept, so {place.postal-code} and {place.locality} read one row, either read twice
// gives one value, and a shown operand is the operand computed. Every other name is
// drawn afresh, so {word} {word} still draws twice. checkTokens, checkPath and
// linkRefs prove every step, so the walk cannot fail.
func readField(s *session, t *template, held *hold, sc renderScope, a arm) draw {
	if !t.held[a.key] {
		if len(a.tail) > 0 {
			panic(fmt.Sprintf("fejkdata: %q reads a path into %q, which the expansion does not hold", a.name, a.key))
		}
		return draw{text: render(s, t.fields[a.key], sc)}
	}
	d := readHold(held, sc, a)
	if r, done := d.value[a.path]; done {
		return r
	}
	leaf, err := walkPath(t.fields[a.key], a.tail, pathWalk{pins: &d.pinSet, draws: &pathDraws{s: s, held: d, a: &a}})
	if err != nil {
		panic(fmt.Sprintf("fejkdata: %q: %v; a fence should have refused this at New", a.name, err))
	}
	r := renderLeaf(s, leaf, sc)
	if d.value == nil {
		d.value = map[string]draw{}
	}
	d.value[a.path] = r
	return r
}

// readHold is the hold a held read keeps its draw in: for a reference that reads a path,
// the render's hold for its group, so its draw spans the render; for a sibling, or a
// reference read whole, held.
func readHold(held *hold, sc renderScope, a arm) *hold {
	if isRef(a.key) && len(a.tail) > 0 {
		return sc.hold()
	}
	return held
}

// renderLeaf draws and renders what a read lands on: null on a null item, or on a column of one
// reference alone whose read drew null.
func renderLeaf(s *session, n node, sc renderScope) draw {
	n = drawn(s, n)
	if _, isNull := n.(*null); isNull {
		return draw{null: true}
	}
	r := draw{text: render(s, n, sc)}
	if t, _ := n.(*template); t != nil && t.readsColumn != nil {
		r.null = sc.in(t).hold().value[t.readsColumn.a.path].null
	}
	return r
}

// drawn resolves a choice to one variant, so a bound head is a concrete node the
// rest of the expansion shares. Nested choices unwrap too: a draw is one value, not
// another set to pick from.
func drawn(s *session, n node) node {
	for c, ok := n.(*choice); ok; c, ok = n.(*choice) {
		n = pick(s, c)
	}
	return n
}
