package fejkdata

import (
	"fmt"
	"sort"
	"strings"
)

// pathWalk is what one walk of a dotted path does at each kind of level: choice
// returns the variants to continue into (none stops the walk); level runs at each
// template a segment descends into; leaf runs where the tail ends. A nil action
// is skipped.
type pathWalk struct {
	choice func(c *choice, rest []string) ([]node, error)
	level  func(t *template, rest []string) error
	leaf   func(n node) error
}

// walkPath descends tail from n: a group or template by its next segment, a
// choice by w.choice, which consumes no segment. A missing segment is an error,
// so no walk reaches past what the data holds. A table-shaped dispatch, one case
// per node kind, kept whole on purpose.
func walkPath(n node, tail []string, w pathWalk) error {
	if len(tail) == 0 {
		if w.leaf != nil {
			return w.leaf(n)
		}
		return nil
	}
	switch n := n.(type) {
	case *group:
		child, ok := n.children[tail[0]]
		if !ok {
			return fmt.Errorf("no entry %q", tail[0])
		}
		return walkPath(child, tail[1:], w)
	case *template:
		if w.level != nil {
			if err := w.level(n, tail); err != nil {
				return err
			}
		}
		child, ok := n.field(tail[0])
		if !ok {
			return fmt.Errorf("no field %q", tail[0])
		}
		return walkPath(child, tail[1:], w)
	case *choice:
		if w.choice == nil {
			return nil
		}
		next, err := w.choice(n, tail)
		if err != nil {
			return err
		}
		for _, item := range next {
			if err := walkPath(item, tail, w); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("no field %q", tail[0])
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
// must carry the rest of the path in the set every variant shares — and that no
// level a path reads carries a repeat or a group, which one draw of it could not
// apply. So a path that validates here resolves on every render, and a typo is a
// New-time error.
func checkPath(n node, tail []string, level string) error {
	return walkPath(n, tail, pathWalk{
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
				return fmt.Errorf("the level %q carries a group, which a path reading into it cannot apply", name)
			}
			return nil
		},
	})
}
