package fejkdata

import (
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// childNamed is what seg names under n: a folder's entry or a template's field.
func childNamed(n node, seg string) (node, error) {
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
		return nil, noEntry{seg}
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
		panic(invariant.Broken("childNamed has no case for node %T", n))
	}
	return nil, fmt.Errorf("no field %q", seg)
}

// pathStep is one step of a compiled path, taken at the node the steps before it
// reached, a choice there drawn first; at indexes the tail where it is taken, and
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

// compiledPath walks a path and proves it reaches a node whichever way the draws go: every
// variant of a choice carries the rest of it, and is walked, a selector names a row inside the
// rows selected before it, and no level read carries a repeat. Once run, it holds the steps a
// draw takes, and every leaf the path may render; level names the head in its errors.
type compiledPath struct {
	pins   pinSet
	level  string
	tail   []string
	steps  []pathStep
	leaves []node
}

func (w *compiledPath) run(n node) (node, error) {
	return w.walk(n, w.tail)
}

func (w *compiledPath) walk(n node, tail []string) (node, error) {
	at := pathPos{n: n, tail: tail}
	for at.more() {
		if c, ok := at.n.(*choice); ok {
			return w.walkEvery(c, at.tail)
		}
		var err error
		if w.steps, err = compileStep(&at, w.tail, w.level, w.steps, &w.pins); err != nil {
			return nil, err
		}
	}
	w.leaves = append(w.leaves, at.n)
	return at.n, nil
}

func repeatLevelError(level string) error {
	return fmt.Errorf("the level %q carries a repeat, so a path cannot read one draw of it; read %q whole", level, level)
}

// pathPos is where a walk stands: the node reached, the tail left to walk from it, and fromRow,
// whether the walk entered n, a table, from a row of a table linked to it.
type pathPos struct {
	n       node
	tail    []string
	fromRow bool
}

func (p pathPos) more() bool { return len(p.tail) > 0 || p.fromRow }

// compileStep takes the first step of at.tail from at.n, which must not be a choice, moves at past
// it, and appends the step to steps: a field, or the route through a table. whole is the full
// path; level is the prefix an error puts before it.
func compileStep(at *pathPos, whole []string, level string, steps []pathStep, pins *pinSet) ([]pathStep, error) {
	i := len(whole) - len(at.tail)
	switch x := at.n.(type) {
	case *table:
		r, err := x.route(at.tail, at.fromRow)
		if err != nil {
			return steps, err
		}
		*at = pathPos{n: r.next, tail: r.rest, fromRow: r.nextLinked}
		return routeSteps(steps, pins, x, r, i)
	case *template:
		if x.repeat > 1 && !grammar.IsSelector(at.tail[0]) {
			return steps, repeatLevelError(join(level, strings.Join(whole[:i], ".")))
		}
	}
	next, err := childNamed(at.n, at.tail[0])
	if err != nil {
		return steps, err
	}
	steps = append(steps, pathStep{kind: stepField, at: i, name: at.tail[0]})
	*at = pathPos{n: next, tail: at.tail[1:]}
	return steps, nil
}

// walkEvery walks every variant and keeps the first one's steps: every variant carries
// the rest of the path, so each compiles to the same steps past the choice.
func (w *compiledPath) walkEvery(c *choice, tail []string) (node, error) {
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

// callerPathSteps proves a path reaches a node without drawing, appending to steps the steps a
// draw takes. Where compiledPath walks every variant of a choice, it walks the first,
// which carriedByAll lets stand for all.
func callerPathSteps(n node, tail []string, steps []pathStep) ([]pathStep, error) {
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
		if steps, err = compileStep(&at, tail, "", steps, &pins); err != nil {
			return nil, err
		}
	}
	return steps, nil
}

// drawSteps draws the rows and variants a path's steps read from n, pinning the rows in pins;
// pins is nil for a sibling path, which never crosses a table, since a table is only a category.
// For a read through a name, memo keeps the variant drawn at each of levels, so paths sharing a
// prefix share it. It returns the leaf and the pins its row is in.
func drawSteps(s *drawstate.State, n node, steps []pathStep, pins *pinSet, memo *drawMemo, levels []pickKey) (node, *pinSet) {
	// climbed is the table a ".." stepped up to. The next step down from it draws the child's row
	// afresh, into pins holding only the rows from climbed up, which memo shares with every path
	// of the pick stepping down there.
	var climbed *rowsTable
	for _, st := range steps {
		if c, ok := n.(*choice); ok {
			n = drawVariant(s, c, memo, levels, st.at)
		}
		if st.kind == stepField {
			var err error
			if n, err = childNamed(n, st.name); err != nil {
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
		n = t.followStep(s, st, pins)
	}
	return n, pins
}

func drawVariant(s *drawstate.State, c *choice, memo *drawMemo, levels []pickKey, at int) node {
	if memo == nil {
		return drawThroughChoices(s, c)
	}
	return memo.variantOf(s, c, levels[at])
}

// carriedByAll is the choice rule a path that must reach a node on every call obeys:
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

func provePath(n node, tail []string, level string) error {
	_, err := (&compiledPath{level: level, tail: tail}).run(n)
	return err
}

// compilePath compiles a path provePath proved.
func compilePath(n node, tail []string) compiledPath {
	w := compiledPath{tail: tail}
	if _, err := w.run(n); err != nil {
		panic(invariant.Broken("%s: %v; the path should have been proved before it was compiled", grammar.JoinSegments(tail), err))
	}
	return w
}
