package fejkdata

import (
	"fmt"
	"sort"
)

// walkNodes calls fn once per contained node, passing the dot path that reaches it,
// visiting keys in sorted order so which of several broken nodes gets reported does
// not depend on map iteration.
func walkNodes(root map[string]node, fn func(path string, n node) error) error {
	for _, name := range sortedNames(root) {
		if err := eachNode(root[name], name, fn); err != nil {
			return err
		}
	}
	return nil
}

// eachNode visits n and every node contained within it once, passing the dot path
// that reaches each. It never crosses a reference edge — a bound {/path} field is
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
			if err := visit(join(path, c.name), c.node); err != nil {
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
		return append([]namedNode{{node: n.formatTemplate}}, namedNodes(n.fields)...)
	case *tableColumn:
		if len(n.t.cellTemplates) == 0 {
			return nil
		}
		var out []namedNode
		for r := 0; r < n.t.rowCount(); r++ {
			if cell := n.t.cellTemplate(r, n.i); cell != nil {
				out = append(out, namedNode{node: cell})
			}
		}
		return out
	default:
		return nil
	}
}

func namedNodes(m map[string]node) []namedNode {
	out := make([]namedNode, 0, len(m))
	for _, name := range sortedNames(m) {
		out = append(out, namedNode{name: name, node: m[name]})
	}
	return out
}

func sortedNames(m map[string]node) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// renderEdge is a child a node renders into, labelled by what reaches it (a field
// name, reference, or choice index) for a readable cycle report. operand names
// the builtin when the label is its operand rather than a token, so an error can
// name it the way the author wrote it.
type renderEdge struct {
	to      node
	label   string
	operand string
}

// reached names an edge as the author spelled it, the vocabulary boundReaders uses
// for the sibling fence.
func (e renderEdge) reached() string {
	if e.operand != "" {
		return fmt.Sprintf("%s operand %q", e.operand, e.label)
	}
	return "{" + e.label + "}"
}

// renderEdges lists the children rendering n recurses into, mirroring expand: a
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
		add := func(a arm, operand string) {
			c := n.head(a.key)
			if c == nil {
				return
			}
			for _, leaf := range pathLeaves(c, a.tail) {
				es = append(es, renderEdge{leaf, a.spelling, operand})
			}
		}
		for _, o := range n.ops {
			for _, a := range o.operands {
				add(a, o.fn)
			}
			for _, a := range o.arms {
				add(a, "")
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
	default:
		return nil
	}
}

// pathLeaves lists what a token's dotted tail renders: a choice on the way
// contributes every variant, since any of them may be the one drawn. checkPath has
// already proved the tail resolves in every variant.
func pathLeaves(n node, tail []string) []node {
	var out []node
	_, _ = walkPath(n, tail, pathWalk{mode: walkEvery, pins: &pinSet{}, leaf: func(n node) error { out = append(out, n); return nil }})
	return out
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
// graph (renderEdges); grey marks nodes on the current path so a back-edge to one
// is the cycle, while black lets a shared node (a DAG, not a cycle) be skipped.
// Every node is a root: a field its parent's format never renders is still reachable
// by dot path, so a cycle in one would otherwise reach render and be fatal there.
func checkNoCycles(root map[string]node) error {
	const (
		grey  = 1
		black = 2
	)
	color := map[node]int{}
	var visit func(n node, path string) error
	visit = func(n node, path string) error {
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
	return walkNodes(root, func(path string, n node) error { return visit(n, path) })
}
