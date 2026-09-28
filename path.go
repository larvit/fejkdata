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

// isSelector reports whether a segment is a [key or name] rather than a name.
func isSelector(seg string) bool { return strings.HasPrefix(seg, "[") }

func hasSelector(segs []string) bool {
	for _, s := range segs {
		if isSelector(s) {
			return true
		}
	}
	return false
}

// selectorOf is the key or name a selector segment holds.
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

// route is how tail passes t, reached from another table where descended. The path
// reads a row wherever a segment follows or the table was reached from another; a
// table reached whole, with no selector, is left to a render's own draw.
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
		r.next = t.pinnedRow
	case child != nil:
		r.next, r.rest, r.descends = child, tail[1:], true
	default:
		r.next, r.rest = column, tail[1:]
	}
	return r, nil
}

// pathCheck proves a path resolves whichever way the draws go: every variant of a
// choice carries the rest of it, and is walked, a selector names a row inside the
// rows selected before it, and no level read carries a repeat or a drawGroup. It
// collects every leaf the path may render; level names the head in its errors.
type pathCheck struct {
	pins   pinSet
	level  string
	tail   []string
	leaves []node
}

func (w *pathCheck) run(n node) (node, error) { return w.walk(n, w.tail) }

func (w *pathCheck) walk(n node, tail []string) (node, error) {
	for descended := false; len(tail) > 0 || descended; {
		var err error
		switch x := n.(type) {
		case *choice:
			return w.walkEvery(x, tail)
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
		tail = tail[1:]
	}
	w.leaves = append(w.leaves, n)
	return n, nil
}

func (w *pathCheck) walkEvery(c *choice, tail []string) (node, error) {
	if err := carriedByAll(c, tail); err != nil {
		return nil, err
	}
	var last node
	for _, item := range c.items {
		var err error
		if last, err = w.walk(item, tail); err != nil {
			return nil, err
		}
	}
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

type pathCover struct{ into map[node]bool }

func (w *pathCover) walk(n node, tail []string) (node, error) {
	for descended := false; len(tail) > 0 || descended; {
		var err error
		switch x := n.(type) {
		case *choice:
			cover(x, w.into, false)
			return x, nil
		case *table:
			var r tableRoute
			if r, err = x.route(tail, descended); err != nil {
				return nil, err
			}
			n, tail, descended = r.next, r.rest, r.descends
			continue
		}
		if n, err = stepInto(n, tail[0]); err != nil {
			return nil, err
		}
		tail = tail[1:]
	}
	cover(n, w.into, false)
	return n, nil
}

// pathProbe proves a path resolves without drawing: carriedByAll lets one variant
// of a choice stand for all. It marks in drawn the tables a draw would pin.
type pathProbe struct {
	pins  pinSet
	drawn map[*table]bool
}

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
			if err := w.readRow(x, r); err != nil {
				return nil, err
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

func (w *pathProbe) readRow(t *table, r tableRoute) error {
	if r.sel != "" {
		return w.pins.selectRow(t, r.sel)
	}
	if !r.draw {
		return nil
	}
	for stop := w.pins.nearestPinned(t); t != stop; t = t.parentT {
		w.drawn[t] = true
	}
	return nil
}

// pathDraw draws the rows and variants a proved path reads, pinning them in pins.
// For a held read, held keeps the variant drawn at each level of a, so paths
// sharing a prefix share it.
type pathDraw struct {
	s    *session
	pins *pinSet
	held *hold
	a    *arm
}

// drawPath walks w over a path proved first, by a probe or at load, so the walk
// cannot fail. level names n in the panic.
func drawPath(n node, tail []string, level string, w *pathDraw) node {
	leaf, err := w.walk(n, tail)
	if err != nil {
		panic(fmt.Sprintf("fejkdata: %s: %v; the path should have been proved before it was drawn", join(level, joinSegments(tail)), err))
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

// variant is the variant of c the walk continues into: the one held for this level
// of the read, drawn once, where the walk holds its draws; else one drawn afresh.
func (w *pathDraw) variant(c *choice, rest []string) node {
	if w.held == nil {
		return pick(w.s, c)
	}
	key := w.a.levels[len(w.a.tail)-len(rest)]
	n, drew := w.held.variant[key]
	if !drew {
		n = resolveChoice(w.s, c)
		if w.held.variant == nil {
			w.held.variant = map[string]node{}
		}
		w.held.variant[key] = n
	}
	return n
}

// selector splits the selector a tail starts with from the rest of it.
func (t *table) selector(tail []string) (sel string, rest []string, err error) {
	if len(tail) == 0 || !isSelector(tail[0]) {
		return "", tail, nil
	}
	sel, rest = selectorOf(tail[0]), tail[1:]
	if len(rest) > 0 && isSelector(rest[0]) {
		return "", nil, fmt.Errorf("%s[%s] is selected twice; one selector names its row", t.category, sel)
	}
	return sel, rest, nil
}

// step is what a tail's first segment names in t: a column, or a table linked to it.
func (t *table) step(tail []string) (column node, child *table, err error) {
	if len(tail) == 0 {
		return nil, nil, nil
	}
	if i, ok := t.col[tail[0]]; ok {
		return t.fields[t.header[i]], nil, nil
	}
	if child = t.descendant(tail[0]); child == nil {
		return nil, nil, fmt.Errorf("no column or linked table %q in %s", tail[0], t.category)
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

// checkPath runs pathCheck over a path read at load; level names the head in its
// errors.
func checkPath(n node, tail []string, level string) error {
	_, err := (&pathCheck{level: level, tail: tail}).run(n)
	return err
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
