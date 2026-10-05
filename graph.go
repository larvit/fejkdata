package fejkdata

import (
	"fmt"
	"slices"
	"sort"

	"github.com/larvit/fejkdata/internal/invariant"
)

// eachNode visits n and every node contained within it once, passing the dot path
// that reaches each, followed by a table cell's line. It never crosses a reference edge — a {/path} reference is
// skipped — so a single inline node is walked on its own.
func eachNode(n node, path string, fn func(path string, n node) error) error {
	seen := map[node]bool{}
	var visit func(string, node) error
	visit = func(path string, m node) error {
		if m == nil || seen[m] {
			return nil
		}
		seen[m] = true
		if err := fn(path, m); err != nil {
			return err
		}
		for _, c := range contained(m) {
			if err := visit(c.labelUnder(path), c.node); err != nil {
				return err
			}
		}
		return nil
	}
	return visit(path, n)
}

// namedNode is a contained child and the segment reaching it; a choice's items carry
// no segment, matching how a dot path steps over a choice.
type namedNode struct {
	name string
	node node
	line int // a table cell's line in its rows file, else 0
}

// labelUnder is the path naming c in an error, a cell by its line.
func (c namedNode) labelUnder(path string) string {
	if c.line > 0 {
		return fmt.Sprintf("%s, line %d", path, c.line)
	}
	return join(path, c.name)
}

func contained(n node) []namedNode {
	switch n := n.(type) {
	case *folder:
		return namedNodes(n.children)
	case *choice:
		out := make([]namedNode, len(n.items))
		for i, it := range n.items {
			out[i] = namedNode{node: it}
		}
		return out
	case *template:
		return namedNodes(n.fields)
	case *table:
		return append([]namedNode{{node: n.formatTemplate}}, namedNodes(n.formatTemplate.fields)...)
	case *tableColumn:
		if len(n.t.cellTemplates) == 0 {
			return nil
		}
		var out []namedNode
		for r := 0; r < n.t.rowCount(); r++ {
			if cell := n.t.cellTemplate(r, n.i); cell != nil {
				out = append(out, namedNode{node: cell, line: r + 2})
			}
		}
		return out
	case *nullItem, *tableRow:
		return nil
	default:
		panic(invariant.Broken("contained has no case for node %T", n))
	}
}

func namedNodes(m map[string]node) []namedNode {
	out := make([]namedNode, 0, len(m))
	for _, name := range sortedNames(m) {
		out = append(out, namedNode{name: name, node: m[name]})
	}
	return out
}

func sortedNames[V any](m map[string]V) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// renderEdge is a child a node renders into, labelled by what reaches it (a field
// name, reference, or choice index) for a readable cycle report.
type renderEdge struct {
	to    node
	label string
}

// renderEdges lists the children rendering n recurses into, mirroring render: a
// choice's items, and a template's field/reference tokens plus its operands. A
// folder renders nothing, so it has no edges.
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
			es = append(es, renderEdge{to: c.node, label: n.t.header[n.i]})
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
func repeatCheck(path string, n node, mem renderCounts) error {
	if t, ok := n.(*template); ok && t.repeat > 1 && mem.renderCount(n) > MaxRepeat {
		return fmt.Errorf("%s: repeat %d multiplies to %d renders along one path, above the maximum %d", path, t.repeat, mem.renderCount(n), MaxRepeat)
	}
	return nil
}

// checkNoCycles rejects a reference cycle: a node whose rendering can reach itself
// — directly, mutually, or through a chain — never terminates, so it must fail at
// New rather than stack-overflow at render. It is a depth-first walk of the render
// graph (renderEdges). Every node is a root: a field its parent's format never
// renders is still reachable by dot path, so a cycle in one would otherwise reach
// render and be fatal there.
func checkNoCycles(s nodeScope) error {
	const (
		grey  = 1
		black = 2
	)
	inScope := map[node]bool{}
	_ = s(func(_ string, n node) error {
		inScope[n] = true
		if t, isTable := n.(*table); isTable {
			inScope[t.rowNode] = true // a selector lands on it, and contained lists it nowhere
		}
		return nil
	})
	color := map[node]int{}
	var visit func(n node, path string) error
	visit = func(n node, path string) error {
		if !inScope[n] {
			return nil // loaded before this scope and proven then, so it never reaches back into it
		}
		switch color[n] {
		case grey:
			return fmt.Errorf("reference cycle: %s", path)
		case black:
			return nil
		}
		color[n] = grey
		for _, e := range renderEdges(n) {
			if err := visit(e.to, path+" -> "+e.label); err != nil {
				return err
			}
		}
		color[n] = black
		return nil
	}
	return s(func(path string, n node) error { return visit(n, path) })
}
