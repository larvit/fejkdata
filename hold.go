package fejkdata

import (
	"fmt"
)

// hold is what has already been drawn for held names: the variant each was drawn as,
// so every path under it reads one row, and the draw each read made, by its one
// spelling, so the same read written twice reads one draw, and the table rows its reads pinned. An
// expansion keeps one for its sibling names, and a render one per group for its reference paths.
type hold struct {
	variant map[string]node
	value   map[string]draw
	pins    pinSet
}

// draw is what one read drew: its text, and whether it landed on a null.
type draw struct {
	text string
	null bool
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
	leaf, err := walkPath(s, t.fields[a.key], a.tail, pathWalk{
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
		pins: &d.pins,
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
