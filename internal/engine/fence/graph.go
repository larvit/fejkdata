package fence

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

// RenderCounts holds, per node, the renders its costliest path down multiplies to, its own
// repeat included; RepeatCheck fills it as it reads it.
type RenderCounts map[tree.Node]int

func (m RenderCounts) renderCount(n tree.Node) int {
	if r, done := m[n]; done {
		return r
	}
	r := 1
	for _, e := range load.RenderEdges(n, nil) {
		if c := m.renderCount(e.To); c > r {
			r = c
		}
	}
	if t, ok := n.(*tree.Template); ok {
		r *= t.Repeat()
	}
	m[n] = r
	return r
}

// RepeatCheck bounds the renders a repeat multiplies to along any root-to-leaf
// path, so nested repeats cannot build what one repeat may not.
// docs/decisions.md#the-repeat-cap-bounds-renders-not-bytes
func RepeatCheck(label string, n tree.Node, mem RenderCounts) error {
	if t, ok := n.(*tree.Template); ok && t.Repeat() > 1 && mem.renderCount(n) > load.MaxRepeat {
		return fmt.Errorf("%s: repeat %d multiplies to %d renders along one path, above the maximum %d", label, t.Repeat(), mem.renderCount(n), load.MaxRepeat)
	}
	return nil
}

// CheckNoCycles rejects a reference cycle: a node whose rendering can reach itself
// — directly, mutually, or through a chain — never terminates, so it must fail at
// New rather than stack-overflow at render. It is a depth-first walk of the render
// graph (load.RenderEdges). Every node is a root: a field its parent's format never
// renders is still reachable by dot path, so a cycle in one would otherwise reach
// render and be fatal there.
func CheckNoCycles(s NodeSet, l *tree.LoadFinds) error {
	const (
		grey  = 1
		black = 2
	)
	inSet := map[tree.Node]bool{}
	_ = s(func(_ string, n tree.Node) error {
		inSet[n] = true
		if t, isTable := n.(*tree.Table); isTable {
			inSet[t.RowNode()] = true // a selector lands on it, and Contained lists it nowhere
		}
		return nil
	})
	color := map[tree.Node]int{}
	var visit func(n tree.Node, label string) error
	visit = func(n tree.Node, label string) error {
		if !inSet[n] {
			return nil // loaded before these nodes and proven then, so it never reaches back into it
		}
		switch color[n] {
		case grey:
			return fmt.Errorf("reference cycle: %s", label)
		case black:
			return nil
		}
		color[n] = grey
		for _, e := range load.RenderEdges(n, l) {
			if err := visit(e.To, label+" -> "+e.Label); err != nil {
				return err
			}
		}
		color[n] = black
		return nil
	}
	return s(func(label string, n tree.Node) error { return visit(n, label) })
}
