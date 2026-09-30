package fejkdata

import (
	"fmt"
	"sort"
	"strings"
)

// splitPath splits a dotted path into segments, a selector — [key or name] after a
// table's name — becoming a segment of its own, brackets kept. A dot inside a
// selector is part of the key or name.
func splitPath(path string) ([]string, error) {
	segs := make([]string, 0, strings.Count(path, ".")+2*strings.Count(path, "[")+1)
	start, open := 0, -1
	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '[':
			if err := checkOpen(path, i, start, open); err != nil {
				return nil, err
			}
			segs = append(segs, path[start:i])
			start, open = i, i
		case ']':
			if err := checkClose(path, i, open); err != nil {
				return nil, err
			}
			segs = append(segs, path[start:i+1])
			start, open = i+2, -1
			i++
		case '.':
			if open < 0 {
				segs = append(segs, path[start:i])
				start = i + 1
			}
		}
	}
	if open >= 0 {
		return nil, fmt.Errorf(`%q opens a selector with "[" and never closes it with "]"`, path)
	}
	if start <= len(path) {
		segs = append(segs, path[start:])
	}
	return segs, nil
}

// checkOpen refuses a "[" at i that opens no selector: one inside a selector, or one
// following no name.
func checkOpen(path string, i, start, open int) error {
	switch {
	case open >= 0:
		return fmt.Errorf(`%q holds a "[" inside a selector, which no key or name may`, path)
	case i == 0:
		return fmt.Errorf(`%q starts with "["; a path starts with a name, and a selector follows a table's name`, path)
	case i == start:
		return fmt.Errorf(`%q: a selector follows its table's name; write %s`, path, path[:i-1]+path[i:])
	}
	return nil
}

// checkClose refuses a "]" at i that closes no selector, closes an empty one, or is
// followed by anything but a dot or the end.
func checkClose(path string, i, open int) error {
	switch {
	case open < 0:
		return fmt.Errorf(`%q holds a "]" that closes no "["`, path)
	case i == open+1:
		return fmt.Errorf("%q holds an empty selector; a selector names a row by key or name", path)
	case i+1 < len(path) && path[i+1] != '.':
		return fmt.Errorf(`%q: a "]" ends its selector, so a dot or the end must follow it`, path)
	}
	return nil
}

// indexOutside is the first c in s outside a [selector], or -1.
func indexOutside(s string, c byte) int {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '[':
			depth++
		case s[i] == ']' && depth > 0:
			depth--
		case s[i] == c && depth == 0:
			return i
		}
	}
	return -1
}

// splitOutside splits s on c outside any [selector].
func splitOutside(s string, c byte) []string {
	var parts []string
	for {
		i := indexOutside(s, c)
		if i < 0 {
			return append(parts, s)
		}
		parts, s = append(parts, s[:i]), s[i+1:]
	}
}

func isSelector(seg string) bool { return strings.HasPrefix(seg, "[") }

func hasSelector(segs []string) bool {
	for _, s := range segs {
		if isSelector(s) {
			return true
		}
	}
	return false
}

func selectorOf(seg string) string { return seg[1 : len(seg)-1] }

// nameSegments is the segments of a path that are names, its selectors left out.
func nameSegments(segs []string) []string {
	out := segs[:0:0]
	for _, s := range segs {
		if !isSelector(s) {
			out = append(out, s)
		}
	}
	return out
}

// stepInto is what seg names under n: a folder's entry or a template's field.
func stepInto(n node, seg string) (node, error) {
	if isSelector(seg) {
		return nil, fmt.Errorf("%s is not a table, so it has no row to select", selectorOf(seg))
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
		return nil, fmt.Errorf("no field %q: %q is a column, and a cell holds no fields", seg, n.t.header[n.i])
	case *nullItem:
	default:
		panic(internalError("stepInto has no case for node %T", n))
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
	column, child, err := t.step(tail)
	if err != nil {
		return tableRoute{}, err
	}
	// A selector further down pins this table by ancestry, so the walk draws only
	// where none follows; drawing first could pick a row the selector is not inside.
	r := tableRoute{sel: sel, draw: sel == "" && (descended || len(tail) > 0) && !hasSelector(tail)}
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
// reached; at indexes the tail segment the step consumes. A step holds no pointer, so a
// caller's steps stay on its stack.
type pathStep struct {
	kind stepKind
	at   int
	name string // stepField: the field; stepSelect: the selector; stepColumn, stepChild: the column or table stepped to
	row  int    // stepSelect: the row the selector names
}

type stepKind uint8

const (
	stepField  stepKind = iota + 1 // into a field of the template at this level, a choice there resolved first
	stepSelect                     // pin the selected row of the table
	stepDraw                       // draw a row of the table inside the pins
	stepRow                        // land on the table's row node
	stepColumn                     // into a column of the table
	stepChild                      // into a table linked to the table
)

// pathCheck proves a path resolves whichever way the draws go: every variant of a
// choice carries the rest of it, and is walked, a selector names a row inside the
// rows selected before it, and no level read carries a repeat or a drawGroup. It
// compiles the steps a draw takes, every leaf the path may render, and as cover the
// first choice it passes, else the leaf; level names the head in its errors.
type pathCheck struct {
	pins   pinSet
	level  string
	tail   []string
	steps  []pathStep
	leaves []node
	cover  node
}

func (w *pathCheck) run(n node) (node, error) {
	leaf, err := w.walk(n, w.tail)
	if err == nil && w.cover == nil {
		w.cover = leaf
	}
	return leaf, err
}

func (w *pathCheck) emit(st pathStep, tail []string) {
	st.at = len(w.tail) - len(tail)
	w.steps = append(w.steps, st)
}

func (w *pathCheck) walk(n node, tail []string) (node, error) {
	for descended := false; len(tail) > 0 || descended; {
		var err error
		switch x := n.(type) {
		case *choice:
			if w.cover == nil {
				w.cover = x
			}
			return w.walkEvery(x, tail)
		case *table:
			var r tableRoute
			if r, err = x.route(tail, descended); err != nil {
				return nil, err
			}
			if err := w.tableSteps(x, r, tail); err != nil {
				return nil, err
			}
			n, tail, descended = r.next, r.rest, r.descends
			continue
		case *template:
			if isSelector(tail[0]) {
				break
			}
			if err := w.enter(x, tail); err != nil {
				return nil, err
			}
		}
		if n, err = stepInto(n, tail[0]); err != nil {
			return nil, err
		}
		w.emit(pathStep{kind: stepField, name: tail[0]}, tail)
		tail = tail[1:]
	}
	w.leaves = append(w.leaves, n)
	return n, nil
}

func (w *pathCheck) tableSteps(t *table, r tableRoute, tail []string) error {
	if r.sel != "" {
		if err := w.pins.selectRow(t, r.sel); err != nil {
			return err
		}
		row, _ := w.pins.pinned(t)
		w.emit(pathStep{kind: stepSelect, name: r.sel, row: row}, tail)
	}
	if r.draw {
		w.emit(pathStep{kind: stepDraw}, tail)
	}
	switch next := r.next.(type) {
	case *table:
		if next != t {
			w.emit(pathStep{kind: stepChild, name: next.segment}, tail)
		}
	case *tableRow:
		w.emit(pathStep{kind: stepRow}, tail)
	case *tableColumn:
		w.emit(pathStep{kind: stepColumn, name: t.header[next.i]}, tail)
	default:
		panic(internalError("%s: a route onto %T", t.segment, r.next))
	}
	return nil
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

func (w *pathCheck) enter(t *template, rest []string) error {
	name := join(w.level, strings.Join(w.tail[:len(w.tail)-len(rest)], "."))
	switch {
	case t.repeat > 1:
		return fmt.Errorf("the level %q carries a repeat, which a path reading one draw of it cannot apply", name)
	case t.drawGroup != "":
		return fmt.Errorf("the level %q carries a drawGroup, which a path reading into it cannot apply", name)
	}
	return nil
}

// pathProbe proves a path resolves without drawing. Where pathCheck walks every
// variant of a choice, it walks the first, which carriedByAll lets stand for all; at a
// table it selects as pathCheck does.
type pathProbe struct{ pins pinSet }

func (w *pathProbe) walk(n node, tail []string) (node, error) {
	for descended := false; len(tail) > 0 || descended; {
		var err error
		switch x := n.(type) {
		case *choice:
			if err := carriedByAll(x, tail); err != nil {
				return nil, err
			}
			n = x.items[0]
			continue
		case *table:
			var r tableRoute
			if r, err = x.route(tail, descended); err != nil {
				return nil, err
			}
			if r.sel != "" {
				if err := w.pins.selectRow(x, r.sel); err != nil {
					return nil, err
				}
			}
			n, tail, descended = r.next, r.rest, r.descends
			continue
		}
		if n, err = stepInto(n, tail[0]); err != nil {
			return nil, err
		}
		tail = tail[1:]
	}
	return n, nil
}

// pathDraw draws the rows and variants a proved path reads, pinning the rows in pins; pins is nil
// for a sibling path, which never crosses a table, since a table is only a category. For a memoized
// read, memo keeps the variant drawn at each level of a, so paths sharing a prefix share it.
// Where pathCheck walks every variant of a choice, it walks the one drawn; at a table it selects
// as pathCheck does and draws the row the route asks for.
type pathDraw struct {
	s    *generatorState
	pins *pinSet
	memo *drawMemo
	a    *arm
}

// drawPath walks w over a path proved first, by a probe or at load, so the walk
// cannot fail. level names n in the panic.
func drawPath(n node, tail []string, level string, w *pathDraw) node {
	leaf, err := w.walk(n, tail)
	if err != nil {
		panic(internalError("%s: %v; the path should have been proved before it was drawn", join(level, joinSegments(tail)), err))
	}
	return leaf
}

func (w *pathDraw) walk(n node, tail []string) (node, error) {
	for descended := false; len(tail) > 0 || descended; {
		var err error
		switch x := n.(type) {
		case *choice:
			n = w.variant(x, tail)
			continue
		case *table:
			var r tableRoute
			if r, err = x.route(tail, descended); err != nil {
				return nil, err
			}
			switch {
			case r.sel != "":
				if err := w.pins.selectRow(x, r.sel); err != nil {
					return nil, err
				}
			case r.draw:
				x.drawIn(w.s, w.pins)
			}
			n, tail, descended = r.next, r.rest, r.descends
			continue
		}
		if n, err = stepInto(n, tail[0]); err != nil {
			return nil, err
		}
		tail = tail[1:]
	}
	return n, nil
}

// variant is the variant of c the walk continues into: the one memo keeps for this level
// of the read, drawn once, where the walk has a memo; else one drawn afresh.
func (w *pathDraw) variant(c *choice, rest []string) node {
	if w.memo == nil {
		return pick(w.s, c)
	}
	key := w.a.levels[len(w.a.tail)-len(rest)]
	n, drew := w.memo.variant[key]
	if !drew {
		n = resolveChoice(w.s, c)
		if w.memo.variant == nil {
			w.memo.variant = map[string]node{}
		}
		w.memo.variant[key] = n
	}
	return n
}

func (t *table) selector(tail []string) (sel string, rest []string, err error) {
	if len(tail) == 0 || !isSelector(tail[0]) {
		return "", tail, nil
	}
	sel, rest = selectorOf(tail[0]), tail[1:]
	if len(rest) > 0 && isSelector(rest[0]) {
		return "", nil, fmt.Errorf("%s[%s] is selected twice; one selector names its row", t.segment, sel)
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
	if child = t.descendant(tail[0]); child == nil {
		return nil, nil, fmt.Errorf("no column or linked table %q in %s", tail[0], t.segment)
	}
	return nil, child, nil
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
	offered := make([]string, 0, len(c.shared))
	for p := range c.shared {
		offered = append(offered, p)
	}
	sort.Strings(offered)
	return fmt.Errorf("not every variant of this %d-way choice carries %q; all carry %v", len(c.items), want, offered)
}

func checkPath(n node, tail []string, level string) error {
	_, err := (&pathCheck{level: level, tail: tail}).run(n)
	return err
}

// compilePath compiles a path checkPath proved.
func compilePath(n node, tail []string) pathCheck {
	w := pathCheck{tail: tail}
	if _, err := w.run(n); err != nil {
		panic(internalError("%s: %v; the path should have been proved before it was compiled", joinSegments(tail), err))
	}
	return w
}

// joinSegments spells segments as a path: a selector attaches to the name before it.
func joinSegments(segs []string) string {
	var b strings.Builder
	for _, s := range segs {
		if b.Len() > 0 && !isSelector(s) {
			b.WriteByte('.')
		}
		b.WriteString(s)
	}
	return b.String()
}
