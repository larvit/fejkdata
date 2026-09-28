package fejkdata

import (
	"fmt"
	"strings"
)

// hold is what one expansion has drawn for the names it holds: the variant each was drawn as, so
// every path under it reads one variant; and the value each read produced, by its path, so the
// same read written twice reads one value.
type hold struct {
	variant map[string]node
	value   map[string]readValue
}

// groupHold is what one render's draw group has drawn for its reference paths: a hold, and the
// table rows those paths pinned.
type groupHold struct {
	hold
	pins pinSet
}

// readValue is what one read drew: its text, and whether it landed on a null.
type readValue struct {
	text string
	null bool
}

// readField renders one arm of a token. A name the expansion holds — a level some
// token addresses by dotted path, or a field an operand reads — is drawn once and
// kept, so {place.postal-code} and {place.locality} read one row, either read twice
// gives one value, and a shown operand is the operand computed. Every other name is
// drawn afresh, so {word} {word} still draws twice. A reference path is kept in the
// render's hold for its group, so its draw spans the render; every other held name in
// held, the expansion's.
func readField(s *session, t *template, held *hold, sc renderScope, a arm) readValue {
	if sc.set.trace != nil {
		traceRead(sc.set.trace, t, sc, a)
	}
	if !t.compiled.held[a.key] {
		if len(a.tail) > 0 {
			panic(fmt.Sprintf("fejkdata: %q reads a path into %q, which the expansion does not hold", a.spelling, a.key))
		}
		return readValue{text: render(s, t.head(a.key), sc)}
	}
	w := pathDraw{s: s, held: held, a: &a}
	if isRef(a.key) && len(a.tail) > 0 {
		group := sc.groupHold()
		w.held, w.pins = &group.hold, &group.pins
	}
	if r, done := w.held.value[a.path]; done {
		return r
	}
	leaf := drawPath(t.head(a.key), a.tail, a.key, &w)
	if w.pins != nil {
		sc = sc.at(leaf, w.pins)
	}
	r := renderLeaf(s, leaf, sc)
	if w.held.value == nil {
		w.held.value = map[string]readValue{}
	}
	w.held.value[a.path] = r
	return r
}

// renderTrace is a test's view of the fields a render reads; nil outside a test.
type renderTrace func(group, table string, row int, a arm)

func traceRead(trace renderTrace, t *template, sc renderScope, a arm) {
	var table string
	var row int
	switch {
	case t.cell.table != nil:
		table, row = t.cell.table.path, t.cell.row
	case t.table != nil:
		table, row = t.table.path, sc.rowOf(t.table)
	}
	trace(strings.Clone(sc.group), strings.Clone(table), row, a)
}

// renderLeaf draws and renders what a read lands on: null on a null item, or on a column of one
// reference alone whose read drew null.
func renderLeaf(s *session, n node, sc renderScope) readValue {
	n = resolveChoice(s, n)
	switch leaf := n.(type) {
	case *nullItem:
		return readValue{null: true}
	case *template:
		if leaf.link.readsColumn != nil {
			return readField(s, leaf, nil, sc.in(leaf), leaf.link.readsColumn.a)
		}
	}
	return readValue{text: render(s, n, sc)}
}

// resolveChoice resolves a choice to one variant, so a held head is a concrete node the
// rest of the expansion shares. Nested choices unwrap too: a draw is one value, not
// another set to pick from.
func resolveChoice(s *session, n node) node {
	for c, ok := n.(*choice); ok; c, ok = n.(*choice) {
		n = pick(s, c)
	}
	return n
}
