package fejkdata

import (
	"fmt"
	"slices"

	"github.com/larvit/fejkdata/internal/invariant"
)

// eachNode visits n and every node inside it once, and passes fn each node with the label an
// error uses for it: the dot path reaching it, then ", line N" for a table cell. It never
// crosses a reference edge — a {/path} reference is skipped — so a single inline node is
// walked on its own.
func eachNode(n node, label string, fn func(label string, n node) error) error {
	seen := map[node]bool{}
	var visit func(string, node) error
	visit = func(label string, m node) error {
		if m == nil || seen[m] {
			return nil
		}
		seen[m] = true
		if err := fn(label, m); err != nil {
			return err
		}
		for _, c := range contained(m) {
			if err := visit(c.labelIn(label), c.node); err != nil {
				return err
			}
		}
		return nil
	}
	return visit(label, n)
}

// containedNode is a contained child and the segment reaching it; a choice's items carry
// no segment, matching how a dot path steps over a choice.
type containedNode struct {
	name string
	node node
	line int // a table cell's line in its rows file, else 0
}

func (c containedNode) labelIn(label string) string {
	if c.line > 0 {
		return fmt.Sprintf("%s, line %d", label, c.line)
	}
	return join(label, c.name)
}

func contained(n node) []containedNode {
	switch n := n.(type) {
	case *folder:
		return containedByName(n.children)
	case *choice:
		out := make([]containedNode, len(n.items))
		for i, it := range n.items {
			out[i] = containedNode{node: it}
		}
		return out
	case *template:
		return containedByName(n.fields)
	case *table:
		return append([]containedNode{{node: n.formatTemplate}}, containedByName(n.formatTemplate.fields)...)
	case *tableColumn:
		if len(n.t.cellTemplates) == 0 {
			return nil
		}
		var out []containedNode
		for r := 0; r < n.t.rows.Len(); r++ {
			if cell := n.t.cellTemplate(r, n.i); cell != nil {
				out = append(out, containedNode{node: cell, line: r + 2})
			}
		}
		return out
	case *nullItem, *tableRow:
		return nil
	default:
		panic(invariant.Broken("contained has no case for node %T", n))
	}
}

func containedByName(m map[string]node) []containedNode {
	out := make([]containedNode, 0, len(m))
	for _, name := range sortedNames(m) {
		out = append(out, containedNode{name: name, node: m[name]})
	}
	return out
}

// renderEdge is a child a node renders into, labelled by what reaches it (a field
// name, reference, or choice index) for a readable cycle report.
type renderEdge struct {
	to    node
	label string
}

// renderEdges lists the children rendering n recurses into, mirroring render.
func renderEdges(n node) []renderEdge {
	switch n := n.(type) {
	case *choice:
		es := make([]renderEdge, len(n.items))
		for i, it := range n.items {
			es[i] = renderEdge{to: it, label: fmt.Sprintf("[%d]", i)}
		}
		return es
	case *template:
		var es []renderEdge
		for _, o := range n.compiled.ops {
			for _, a := range slices.Concat(o.operands, o.arms) {
				for _, leaf := range a.leaves {
					es = append(es, renderEdge{leaf, a.spelling})
				}
			}
		}
		return es
	case *table:
		return []renderEdge{{to: n.formatTemplate, label: "format"}}
	case *tableRow:
		return []renderEdge{{to: n.t.formatTemplate, label: "format"}}
	case *tableColumn:
		var es []renderEdge
		for _, c := range contained(n) {
			es = append(es, renderEdge{to: c.node, label: n.name()})
		}
		return es
	case *folder, *nullItem:
		return nil
	default:
		panic(invariant.Broken("renderEdges has no case for node %T", n))
	}
}

type renderCounts map[node]int

func (m renderCounts) renderCount(n node) int {
	if r, done := m[n]; done {
		return r
	}
	r := 1
	for _, e := range renderEdges(n) {
		if c := m.renderCount(e.to); c > r {
			r = c
		}
	}
	if t, ok := n.(*template); ok {
		r *= t.repeat
	}
	m[n] = r
	return r
}

// repeatCheck bounds the renders a repeat multiplies to along any root-to-leaf
// path, so nested repeats cannot build what one repeat may not.
// docs/decisions.md#the-repeat-cap-bounds-renders-not-bytes
func repeatCheck(label string, n node, mem renderCounts) error {
	if t, ok := n.(*template); ok && t.repeat > 1 && mem.renderCount(n) > MaxRepeat {
		return fmt.Errorf("%s: repeat %d multiplies to %d renders along one path, above the maximum %d", label, t.repeat, mem.renderCount(n), MaxRepeat)
	}
	return nil
}

// checkNoCycles rejects a reference cycle: a node whose rendering can reach itself
// — directly, mutually, or through a chain — never terminates, so it must fail at
// New rather than stack-overflow at render. It is a depth-first walk of the render
// graph (renderEdges). Every node is a root: a field its parent's format never
// renders is still reachable by dot path, so a cycle in one would otherwise reach
// render and be fatal there.
func checkNoCycles(s nodeSet) error {
	const (
		grey  = 1
		black = 2
	)
	inSet := map[node]bool{}
	_ = s(func(_ string, n node) error {
		inSet[n] = true
		if t, isTable := n.(*table); isTable {
			inSet[t.rowNode] = true // a selector lands on it, and contained lists it nowhere
		}
		return nil
	})
	color := map[node]int{}
	var visit func(n node, label string) error
	visit = func(n node, label string) error {
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
		for _, e := range renderEdges(n) {
			if err := visit(e.to, label+" -> "+e.label); err != nil {
				return err
			}
		}
		color[n] = black
		return nil
	}
	return s(func(label string, n node) error { return visit(n, label) })
}
