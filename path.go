package fejkdata

import (
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// stepInto is what seg names under n: a folder's entry or a template's field.
func stepInto(n node, seg string) (node, error) {
	switch {
	case grammar.IsSelector(seg):
		return nil, fmt.Errorf("%s is not a table, so it has no row to select", grammar.SelectorOf(seg))
	case seg == "..":
		if c, ok := n.(*tableColumn); ok {
			return nil, fmt.Errorf(`".." steps up from a table's row, and %q is a column; put the ".." right after the row`, c.name())
		}
		return nil, fmt.Errorf(`".." steps up only from a table's row; name any other level by its path from the data root`)
	}
	switch n := n.(type) {
	case *folder:
		if child, ok := n.children[seg]; ok {
			return child, nil
		}
		return nil, fmt.Errorf("no entry %q", seg)
	case *template:
		if child, ok := n.fields[seg]; ok {
			return child, nil
		}
	case *tableColumn:
		if n.name() == n.t.rows.Options().Parent {
			return nil, fmt.Errorf("no field %q: %q is the link column, holding its parent's key; to read the row it links to, step up with ..%s.%s", seg, n.name(), n.name(), seg)
		}
		return nil, fmt.Errorf("no field %q: %q is a column, and a cell holds no fields", seg, n.name())
	case *nullItem:
	default:
		panic(invariant.Broken("stepInto has no case for node %T", n))
	}
	return nil, fmt.Errorf("no field %q", seg)
}

// tableRoute is how a path passes one table: the row it reads, by selector or
// drawn, what it goes on into, and whether that is a linked table.
type tableRoute struct {
	sel      string
	draw     bool
	next     node
	rest     []string
	descends bool
	up       bool
}

// route is how tail passes t. descended means the previous route stepped into t
// from a row of an ancestor table, so t reads a row even where tail is empty; a
// table reached otherwise, with no selector and an empty tail, is left to a render's
// own draw.
func (t *table) route(tail []string, descended bool) (tableRoute, error) {
	sel, tail, err := t.selector(tail)
	if err != nil {
		return tableRoute{}, err
	}
	if len(tail) > 0 && tail[0] == ".." {
		if err := t.rows.ProveStepUp(tail[1:]); err != nil {
			return tableRoute{}, err
		}
		return tableRoute{sel: sel, draw: sel == "", next: t.rows.Parent().Owner(), rest: tail[2:], descends: true, up: true}, nil
	}
	column, child, err := t.step(tail)
	if err != nil {
		return tableRoute{}, err
	}
	// A selector further down pins this table by ancestry, so the walk draws only
	// where none follows; drawing first could pick a row the selector is not inside.
	r := tableRoute{sel: sel, draw: sel == "" && (descended || len(tail) > 0) && !grammar.HasSelector(tail)}
	switch {
	case len(tail) == 0 && sel == "" && !descended:
		r.next = t
	case len(tail) == 0:
		r.next = t.rowNode
	case child != nil:
		r.next, r.rest, r.descends = child, tail[1:], true
	default:
		r.next, r.rest = column, tail[1:]
	}
	return r, nil
}

// pathStep is one step of a compiled path, taken at the node the steps before it
// reached, a choice there resolved first; at indexes the tail where it is taken, and
// only a table's step sits past the tail's end. A step holds no node, so a
// caller's steps stay on its stack.
type pathStep struct {
	kind stepKind
	at   int
	name string // stepField: the field; stepSelect: the selector; stepColumn, stepChild: the column or table stepped to
	row  int    // stepSelect: the row the selector names
}

type stepKind uint8

const (
	stepField  stepKind = iota + 1 // into a folder's entry or a template's field
	stepSelect                     // pin the selected row of the table
	stepDraw                       // draw a row of the table inside the pins
	stepRow                        // land on the table's row node
	stepColumn
	stepChild
	stepParent // up to the table's parent, at the row its own row links to
)

// pathCheck proves a path resolves whichever way the draws go: every variant of a
// choice carries the rest of it, and is walked, a selector names a row inside the
// rows selected before it, and no level read carries a repeat. It compiles the steps
// a draw takes, and every leaf the path may render; level names the head in its errors.
type pathCheck struct {
	pins   pinSet
	level  string
	tail   []string
	steps  []pathStep
	leaves []node
}

func (w *pathCheck) run(n node) (node, error) {
	return w.walk(n, w.tail)
}

func (w *pathCheck) walk(n node, tail []string) (node, error) {
	at := pathPos{n: n, tail: tail}
	for at.more() {
		if c, ok := at.n.(*choice); ok {
			return w.walkEvery(c, at.tail)
		}
		var err error
		if w.steps, err = takeStep(&at, w.tail, w.level, w.steps, &w.pins); err != nil {
			return nil, err
		}
	}
	w.leaves = append(w.leaves, at.n)
	return at.n, nil
}

// pathPos is where a walk stands: the node reached, the tail left to walk from it, and whether
// the step into it came from a row of an ancestor table, so a table there reads a row even
// where the tail is empty.
type pathPos struct {
	n         node
	tail      []string
	descended bool
}

func (p pathPos) more() bool { return len(p.tail) > 0 || p.descended }

// takeStep moves at past the step its tail starts with, from its node, which is no choice, and
// appends what it takes to steps: a field, or the route a table passes. whole is the path at's
// tail ends, and level names its head in errors.
func takeStep(at *pathPos, whole []string, level string, steps []pathStep, pins *pinSet) ([]pathStep, error) {
	i := len(whole) - len(at.tail)
	switch x := at.n.(type) {
	case *table:
		r, err := x.route(at.tail, at.descended)
		if err != nil {
			return steps, err
		}
		*at = pathPos{r.next, r.rest, r.descends}
		return routeSteps(steps, pins, x, r, i)
	case *template:
		if x.repeat > 1 && !grammar.IsSelector(at.tail[0]) {
			return steps, fmt.Errorf("the level %q carries a repeat, which a path reading one draw of it cannot apply", join(level, strings.Join(whole[:i], ".")))
		}
	}
	next, err := stepInto(at.n, at.tail[0])
	if err != nil {
		return steps, err
	}
	steps = append(steps, pathStep{kind: stepField, at: i, name: at.tail[0]})
	*at = pathPos{n: next, tail: at.tail[1:]}
	return steps, nil
}

// routeSteps appends the steps r takes past t, the first at at, pinning in pins the row
// it selects.
func routeSteps(steps []pathStep, pins *pinSet, t *table, r tableRoute, at int) ([]pathStep, error) {
	if r.sel != "" {
		row, err := pins.Select(t.rows, r.sel)
		if err != nil {
			return steps, err
		}
		steps = append(steps, pathStep{kind: stepSelect, at: at, name: r.sel, row: row})
		at++
	}
	if r.draw {
		steps = append(steps, pathStep{kind: stepDraw, at: at})
	}
	switch next := r.next.(type) {
	case *table:
		switch {
		case r.up:
			steps = append(steps, pathStep{kind: stepParent, at: at})
		case next != t:
			steps = append(steps, pathStep{kind: stepChild, at: at, name: next.rows.Segment()})
		}
	case *tableRow:
		steps = append(steps, pathStep{kind: stepRow, at: at})
	case *tableColumn:
		steps = append(steps, pathStep{kind: stepColumn, at: at, name: next.name()})
	default:
		panic(invariant.Broken("%s: a route onto %T", t.rows.Segment(), r.next))
	}
	return steps, nil
}

// walkEvery walks every variant and keeps the first one's steps: every variant carries
// the rest of the path, so each compiles to the same steps past the choice.
func (w *pathCheck) walkEvery(c *choice, tail []string) (node, error) {
	if err := carriedByAll(c, tail); err != nil {
		return nil, err
	}
	var last node
	before, first := w.steps[:len(w.steps):len(w.steps)], w.steps
	for i, item := range c.items {
		var err error
		w.steps = before
		if last, err = w.walk(item, tail); err != nil {
			return nil, err
		}
		if i == 0 {
			first = w.steps
		}
	}
	w.steps = first
	return last, nil
}

// probePath proves a path resolves without drawing, appending to steps the steps a
// draw takes. Where pathCheck walks every variant of a choice, it walks the first,
// which carriedByAll lets stand for all.
func probePath(n node, tail []string, steps []pathStep) ([]pathStep, error) {
	var pins pinSet
	for at := (pathPos{n: n, tail: tail}); at.more(); {
		if c, ok := at.n.(*choice); ok {
			if err := carriedByAll(c, at.tail); err != nil {
				return nil, err
			}
			at.n = c.items[0]
			continue
		}
		var err error
		if steps, err = takeStep(&at, tail, "", steps, &pins); err != nil {
			return nil, err
		}
	}
	return steps, nil
}

// drawSteps draws the rows and variants a path's steps read from n, pinning the rows in pins;
// pins is nil for a sibling path, which never crosses a table, since a table is only a category.
// For a read through a name, memo keeps the variant drawn at each of levels, so paths sharing a
// prefix share it. It returns the leaf and the pins its row is in: a step down after a step up draws
// afresh, into pins of its own, which memo shares with every path stepping down there.
func drawSteps(s *drawstate.State, n node, steps []pathStep, pins *pinSet, memo *drawMemo, levels []string) (node, *pinSet) {
	var climbed *rowsTable
	for _, st := range steps {
		if c, ok := n.(*choice); ok {
			n = drawVariant(s, c, memo, levels, st.at)
		}
		if st.kind == stepField {
			var err error
			if n, err = stepInto(n, st.name); err != nil {
				panic(invariant.Broken("step %d: %v; a path's steps should have been compiled from the node they are drawn from", st.at, err))
			}
			continue
		}
		t, ok := n.(*table)
		if !ok {
			panic(invariant.Broken("step %d: a table's step of kind %d on %T", st.at, st.kind, n))
		}
		switch {
		case st.kind == stepParent:
			climbed = t.rows.Parent()
		case st.kind == stepChild && climbed != nil:
			pins, climbed = memo.stepDownPins(pins, climbed, levels, st.at), nil
		}
		n = t.drawStep(s, st, pins)
	}
	return n, pins
}

func (t *table) drawStep(s *drawstate.State, st pathStep, pins *pinSet) node {
	switch st.kind {
	case stepSelect:
		if err := pins.PinRow(t.rows, st.row); err != nil {
			panic(invariant.Broken("%s[%s]: %v", t.rows.Segment(), st.name, err))
		}
		return t
	case stepDraw:
		t.rows.DrawIn(s, pins)
		return t
	case stepRow:
		return t.rowNode
	case stepColumn:
		return t.formatTemplate.fields[st.name]
	case stepChild:
		return t.rows.Descendant(st.name).Owner()
	case stepParent:
		return t.rows.Parent().Owner()
	}
	panic(invariant.Broken("drawStep has no case for step kind %d", st.kind))
}

func drawVariant(s *drawstate.State, c *choice, memo *drawMemo, levels []string, at int) node {
	if memo == nil {
		return resolveChoice(s, c)
	}
	return memo.variantOf(s, c, levels[at])
}

func (t *table) selector(tail []string) (sel string, rest []string, err error) {
	if len(tail) == 0 || !grammar.IsSelector(tail[0]) {
		return "", tail, nil
	}
	sel, rest = grammar.SelectorOf(tail[0]), tail[1:]
	if len(rest) > 0 && grammar.IsSelector(rest[0]) {
		return "", nil, fmt.Errorf("%s[%s] is selected twice; one selector names its row", t.rows.Segment(), sel)
	}
	return sel, rest, nil
}

// step is what a tail's first segment names in t: a column, or a table linked to it.
func (t *table) step(tail []string) (column node, child *table, err error) {
	if len(tail) == 0 {
		return nil, nil, nil
	}
	if column, ok := t.formatTemplate.fields[tail[0]]; ok {
		return column, nil, nil
	}
	d := t.rows.Descendant(tail[0])
	if d == nil {
		return nil, nil, fmt.Errorf("no column or linked table %q in %s", tail[0], t.rows.Segment())
	}
	return nil, d.Owner(), nil
}

// carriedByAll is the choice rule a path that must resolve on every call obeys:
// the rest of the tail must be one every variant carries.
func carriedByAll(c *choice, rest []string) error {
	if want := strings.Join(rest, "."); !c.shared[want] {
		return unreachableInChoice(c, want)
	}
	return nil
}

// unreachableInChoice reports that a path cannot step through this choice, listing
// what every variant does carry.
func unreachableInChoice(c *choice, want string) error {
	if len(c.shared) == 0 {
		return fmt.Errorf("no variant of this %d-way choice carries %q", len(c.items), want)
	}
	offered := sortedNames(c.shared)
	return fmt.Errorf("not every variant of this %d-way choice carries %q; all carry %v", len(c.items), want, offered)
}

func checkPathResolves(n node, tail []string, level string) error {
	_, err := (&pathCheck{level: level, tail: tail}).run(n)
	return err
}

// compilePath compiles a path checkPathResolves proved.
func compilePath(n node, tail []string) pathCheck {
	w := pathCheck{tail: tail}
	if _, err := w.run(n); err != nil {
		panic(invariant.Broken("%s: %v; the path should have been proved before it was compiled", grammar.JoinSegments(tail), err))
	}
	return w
}
