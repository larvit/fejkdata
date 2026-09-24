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

// names is the segments of a path that are names, its selectors left out.
func names(segs []string) []string {
	out := segs[:0:0]
	for _, s := range segs {
		if !isSelector(s) {
			out = append(out, s)
		}
	}
	return out
}

// pathWalk is what one walk of a dotted path does at each kind of level: choice
// returns the variants to continue into (none stops the walk), and where it is nil
// the walk draws one with s, takes the first where s is nil, or stops where pins is
// nil too; level runs at each template a segment descends
// into; leaf runs where the tail ends; pins is where the walk pins the table rows
// it selects and draws, nil skipping tables. A nil action is skipped. A render's
// walk sets only pins, so it allocates nothing.
type pathWalk struct {
	choice func(c *choice, rest []string) ([]node, error)
	level  func(t *template, rest []string) error
	leaf   func(n node) error
	pins   *pinSet
}

// walkPath descends tail from n and returns the node it ends at: a folder or
// template by its next segment, a choice by w.choice, which consumes no segment, a
// table by walkTable. A missing segment is an error, so no walk reaches past what
// the data holds. s draws rows and variants; where it is nil, only selectors pin.
// s is a parameter: a pathWalk field beside pins moves the hold maps to the heap.
func walkPath(s *session, n node, tail []string, w pathWalk) (node, error) {
	if len(tail) == 0 {
		return n, w.atLeaf(n)
	}
	if t, isTable := n.(*table); isTable {
		return walkTable(s, t, tail, w, false)
	}
	if isSelector(tail[0]) {
		return nil, fmt.Errorf("%s is not a table, so it has no row to select", selectorOf(tail[0]))
	}
	switch n := n.(type) {
	case *folder:
		child, ok := n.children[tail[0]]
		if !ok {
			return nil, fmt.Errorf("no entry %q", tail[0])
		}
		return walkPath(s, child, tail[1:], w)
	case *template:
		if w.level != nil {
			if err := w.level(n, tail); err != nil {
				return nil, err
			}
		}
		child, ok := n.field(tail[0])
		if !ok {
			return nil, fmt.Errorf("no field %q", tail[0])
		}
		return walkPath(s, child, tail[1:], w)
	case *choice:
		return walkChoice(s, n, tail, w)
	case *column:
		return nil, fmt.Errorf("no field %q: %q is a column, and a cell holds no fields", tail[0], n.t.columns[n.i])
	}
	return nil, fmt.Errorf("no field %q", tail[0])
}

func (w pathWalk) atLeaf(n node) error {
	if w.leaf != nil {
		return w.leaf(n)
	}
	return nil
}

// walkChoice continues a walk into the variants w.choice returns, or into one drawn
// by s where the walk names no choice action.
func walkChoice(s *session, c *choice, tail []string, w pathWalk) (node, error) {
	if w.choice == nil {
		if w.pins == nil {
			return nil, nil
		}
		if err := carriedByAll(c, tail); err != nil {
			return nil, err
		}
		if s == nil { // a probe: every variant carries the tail, so any one proves it
			return walkPath(s, c.items[0], tail, w)
		}
		return walkPath(s, pick(s, c), tail, w)
	}
	next, err := w.choice(c, tail)
	if err != nil {
		return nil, err
	}
	var last node
	for _, item := range next {
		if last, err = walkPath(s, item, tail, w); err != nil {
			return nil, err
		}
	}
	return last, nil
}

// walkTable descends tail from a table: a selector first, then a column or a
// linked table by name. The walk reads a row wherever a segment follows or the
// table was reached from another; a table reached whole, with no selector, is left
// to a render's own draw.
func walkTable(s *session, t *table, tail []string, w pathWalk, descended bool) (node, error) {
	sel, tail, err := t.selector(tail)
	if err != nil {
		return nil, err
	}
	column, child, err := t.step(tail)
	if err != nil {
		return nil, err
	}
	// A selector further down pins this table by ancestry, so the walk draws only
	// where none follows; drawing first could pick a row the selector is not inside.
	if w.pins != nil {
		if err := readRow(s, w.pins, t, sel, (descended || len(tail) > 0) && !hasSelector(tail)); err != nil {
			return nil, err
		}
	}
	switch {
	case len(tail) == 0 && sel == "" && !descended:
		return walkPath(s, t, nil, w)
	case len(tail) == 0:
		return walkPath(s, t.whole, nil, w)
	case child != nil:
		return walkTable(s, child, tail[1:], w, true)
	}
	return walkPath(s, column, tail[1:], w)
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
		return t.fields[t.columns[i]], nil, nil
	}
	if child = t.descendant(tail[0]); child == nil {
		return nil, nil, fmt.Errorf("no column or linked table %q in %s", tail[0], t.category)
	}
	return nil, child, nil
}

// readRow pins the row a path reads of t: the one its selector names, or, where the
// walk draws, one drawn where the path reads into the table.
func readRow(s *session, pins *pinSet, t *table, sel string, draw bool) error {
	if sel != "" {
		return pins.selectRow(t, sel)
	}
	if draw && s != nil {
		pins.rowOf(s, t)
	}
	return nil
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
// what every variant does carry. It reads the precomputed set, so a failing path
// costs no more than a rendering one.
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

// checkPath proves a dotted tail resolves whichever way the draws go — a choice
// must carry the rest of the path in the set every variant shares, a selector must
// name one row inside the rows selected before it — and that no level a path reads
// carries a repeat or a drawGroup, which one draw of it could not apply. So a path
// that validates here resolves on every render, and a typo is a New-time error.
func checkPath(n node, tail []string, level string) error {
	var pins pinSet
	_, err := walkPath(nil, n, tail, pathWalk{
		choice: func(c *choice, rest []string) ([]node, error) {
			if err := carriedByAll(c, rest); err != nil {
				return nil, err
			}
			return c.items, nil
		},
		level: func(t *template, rest []string) error {
			name := join(level, strings.Join(tail[:len(tail)-len(rest)], "."))
			switch {
			case t.repeat > 1:
				return fmt.Errorf("the level %q carries a repeat, which a path reading one draw of it cannot apply", name)
			case t.drawGroup != "":
				return fmt.Errorf("the level %q carries a drawGroup, which a path reading into it cannot apply", name)
			}
			return nil
		},
		pins: &pins,
	})
	return err
}
