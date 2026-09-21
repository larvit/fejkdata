package fejkdata

import (
	"fmt"
	"sort"
)

// tablePin is one table's pinned row.
type tablePin struct {
	t   *table
	row int
}

// draws is what has already been drawn for held names: the variant each was drawn as,
// so every path under it reads one row, and the draw each read made, by its one
// spelling, so the same read written twice reads one draw. An expansion keeps one for
// its sibling names, and a render one per group for its reference paths.
type draws struct {
	variant map[string]node
	value   map[string]draw
	pins    [8]tablePin // the rows pinned, inline so a render over a country's five-deep geo tree stays off the heap
	npins   int
	more    map[*table]int // the rows pinned past the inline eight
	s       *session       // what draws a row; nil where a walk only proves selectors
}

// draw is what one read drew: its text, and whether it landed on a null.
type draw struct {
	text string
	null bool
}

// pinned is the row the render pinned for t, if any.
func (d *draws) pinned(t *table) (int, bool) {
	for _, p := range d.pins[:d.npins] {
		if p.t == t {
			return p.row, true
		}
	}
	r, ok := d.more[t]
	return r, ok
}

// mustRow is the row pinned for t, which the walk reaching a column pinned.
func (d *draws) mustRow(t *table) int {
	r, ok := d.pinned(t)
	if !ok {
		panic(fmt.Sprintf("fejkdata: a column of %s is rendered with no row pinned", t.category))
	}
	return r
}

// pin pins row r of t, and the rows of t's ancestors it links to.
func (d *draws) pin(t *table, r int) {
	for {
		if _, done := d.pinned(t); done {
			return
		}
		if d.npins < len(d.pins) {
			d.pins[d.npins] = tablePin{t, r}
			d.npins++
		} else {
			if d.more == nil {
				d.more = map[*table]int{}
			}
			d.more[t] = r
		}
		if t.parentT == nil {
			return
		}
		t, r = t.parentT, t.parentRow(r)
	}
}

// each calls fn for every pinned row, in pin order, the spilled ones by table name.
func (d *draws) each(fn func(t *table, r int)) {
	for _, p := range d.pins[:d.npins] {
		fn(p.t, p.row)
	}
	spilled := make([]*table, 0, len(d.more))
	for t := range d.more {
		spilled = append(spilled, t)
	}
	sort.Slice(spilled, func(i, j int) bool { return spilled[i].category < spilled[j].category })
	for _, t := range spilled {
		fn(t, d.more[t])
	}
}

// rowOf is the render's row of t: the one pinned, else one drawn inside the
// nearest pinned ancestor — its parent drawn inside that first where the ancestor
// is further up — or over the whole table, and pinned with its ancestors.
func (d *draws) rowOf(t *table) int {
	if r, ok := d.pinned(t); ok {
		return r
	}
	r := -1
	for a := t.parentT; a != nil && r < 0; a = a.parentT {
		if _, ok := d.pinned(a); !ok {
			continue
		}
		if t.parentT != a {
			d.rowOf(t.parentT)
		}
		pr, _ := d.pinned(t.parentT)
		r = t.drawUnder(d.s, pr)
	}
	if r < 0 {
		r = t.draw(d.s)
	}
	d.pin(t, r)
	return r
}

// pinRow pins row r of t where it agrees with the rows pinned before it.
func (d *draws) pinRow(t *table, r int) error {
	if pr, ok := d.pinned(t); ok && pr != r {
		return fmt.Errorf("%s and %s are two rows of %s", t.selectorSpelling(pr), t.selectorSpelling(r), t.path)
	}
	for a := t.parentT; a != nil; a = a.parentT {
		if pa, ok := d.pinned(a); ok && !t.under(r, a, pa) {
			return fmt.Errorf("%s is not inside %s", t.selectorSpelling(r), a.selectorSpelling(pa))
		}
	}
	d.pin(t, r)
	return nil
}

// selectRow pins the row a selector names.
func (d *draws) selectRow(t *table, sel string) error {
	r, err := t.find(sel, d)
	if err != nil {
		return err
	}
	return d.pinRow(t, r)
}

// inside keeps the rows of t that sit inside every pinned ancestor.
func (d *draws) inside(t *table, rows []int) []int {
	for a := t.parentT; a != nil; a = a.parentT {
		pa, ok := d.pinned(a)
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

// readField renders one arm of a token. An arm's key is a sibling field or a
// reference linkRefs bound into fields. A name the expansion holds — a level some
// token addresses by dotted path, or a field an operand reads — is drawn once and
// kept, so {place.postal-code} and {place.locality} read one row, either read twice
// gives one value, and a shown operand is the operand computed. Every other name is
// drawn afresh, so {word} {word} still draws twice. checkTokens, checkPath and
// linkRefs prove every step, so the walk cannot fail.
func readField(s *session, t *template, held *draws, sc drawScope, a arm) draw {
	if !t.held[a.key] {
		if len(a.tail) > 0 {
			panic(fmt.Sprintf("fejkdata: %q reads a path into %q, which the expansion does not hold", a.name, a.key))
		}
		return draw{text: render(s, t.fields[a.key], sc)}
	}
	d := readScope(s, held, sc, a)
	if r, done := d.value[a.path]; done {
		return r
	}
	leaf, err := walkPath(t.fields[a.key], a.tail, pathWalk{
		// Hold the draw at every level passed through, so two paths sharing a
		// prefix share it.
		choice: func(c *choice, rest []string) ([]node, error) {
			key := a.key
			if consumed := len(a.tail) - len(rest); consumed > 0 {
				key = a.steps[consumed-1]
			}
			n, drew := d.variant[key]
			if !drew {
				n = drawn(s, c)
				if d.variant == nil {
					d.variant = map[string]node{}
				}
				d.variant[key] = n
			}
			return []node{n}, nil
		},
		pins: d,
	})
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

// readScope is the draws a held read keeps its draw in: for a reference that reads a path,
// the render's draws for its group, so its draw spans the render; for a sibling, or a
// reference read whole, held.
func readScope(s *session, held *draws, sc drawScope, a arm) *draws {
	if isRef(a.key) && len(a.tail) > 0 {
		return sc.draws(s)
	}
	return held
}

// renderLeaf draws and renders what a read lands on: null on a null item, or on a column of one
// reference alone whose read drew null.
func renderLeaf(s *session, n node, sc drawScope) draw {
	n = drawn(s, n)
	if _, isNull := n.(*null); isNull {
		return draw{null: true}
	}
	r := draw{text: render(s, n, sc)}
	if t, _ := n.(*template); t != nil && t.readsColumn != nil {
		r.null = sc.in(t).draws(s).value[t.readsColumn.a.path].null
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
