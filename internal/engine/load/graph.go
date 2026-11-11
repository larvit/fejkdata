package load

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
)

// EachNode visits n and every node inside it once, and passes fn each node with the label an
// error uses for it: the dot path reaching it, then ", line N" for a table cell. It never
// crosses a reference edge — a {/path} reference is skipped — so a single inline node is
// walked on its own.
func EachNode(n tree.Node, label string, fn func(label string, n tree.Node) error) error {
	seen := map[tree.Node]bool{}
	var visit func(string, tree.Node) error
	visit = func(label string, m tree.Node) error {
		if m == nil || seen[m] {
			return nil
		}
		seen[m] = true
		if err := fn(label, m); err != nil {
			return err
		}
		for _, c := range Contained(m) {
			if err := visit(c.labelIn(label), c.Node); err != nil {
				return err
			}
		}
		return nil
	}
	return visit(label, n)
}

// ContainedNode is a child Contained lists, and the segment reaching it; a choice's items carry
// no segment, matching how a dot path steps over a choice.
type ContainedNode struct {
	name string
	Node tree.Node
	line int // a table cell's line in its rows file, else 0
}

func (c ContainedNode) labelIn(label string) string {
	if c.line > 0 {
		return fmt.Sprintf("%s, line %d", label, c.line)
	}
	return join(label, c.name)
}

// Contained lists the nodes n holds, each with the segment reaching it.
func Contained(n tree.Node) []ContainedNode {
	switch n := n.(type) {
	case *tree.Folder:
		return containedByName(n.Children())
	case *tree.Choice:
		out := make([]ContainedNode, len(n.Items()))
		for i, it := range n.Items() {
			out[i] = ContainedNode{Node: it}
		}
		return out
	case *tree.Template:
		return containedByName(n.Fields())
	case *tree.Table:
		return append([]ContainedNode{{Node: n.FormatTemplate()}}, containedByName(n.FormatTemplate().Fields())...)
	case *tree.TableColumn:
		if len(n.Table().CellTemplates()) == 0 {
			return nil
		}
		var out []ContainedNode
		for r := 0; r < n.Table().Rows().Len(); r++ {
			if cell := n.Table().CellTemplate(r, n.Index()); cell != nil {
				out = append(out, ContainedNode{Node: cell, line: r + 2})
			}
		}
		return out
	case *tree.NullItem, *tree.TableRow:
		return nil
	default:
		panic(invariant.Broken("Contained has no case for node %T", n))
	}
}

func containedByName(m map[string]tree.Node) []ContainedNode {
	out := make([]ContainedNode, 0, len(m))
	for _, name := range tree.SortedNames(m) {
		out = append(out, ContainedNode{name: name, Node: m[name]})
	}
	return out
}

// RenderEdge is a child a node renders into, labelled by what reaches it (a field
// name, reference, or choice index) for a readable cycle report.
type RenderEdge struct {
	To    tree.Node
	label string
}

// RenderEdges lists the children rendering n recurses into, mirroring render.
func RenderEdges(n tree.Node, l *tree.LoadFinds) []RenderEdge {
	switch n := n.(type) {
	case *tree.Choice:
		es := make([]RenderEdge, len(n.Items()))
		for i, it := range n.Items() {
			es[i] = RenderEdge{To: it, label: fmt.Sprintf("[%d]", i)}
		}
		return es
	case *tree.Template:
		var es []RenderEdge
		for _, o := range l.Of(n).Compiled().Ops() {
			for _, a := range o.Reads() {
				for _, leaf := range a.Leaves() {
					es = append(es, RenderEdge{leaf, a.Spelling()})
				}
			}
		}
		return es
	case *tree.Table:
		return []RenderEdge{{To: n.FormatTemplate(), label: "format"}}
	case *tree.TableRow:
		return []RenderEdge{{To: n.Table().FormatTemplate(), label: "format"}}
	case *tree.TableColumn:
		var es []RenderEdge
		for _, c := range Contained(n) {
			es = append(es, RenderEdge{To: c.Node, label: n.Name()})
		}
		return es
	case *tree.Folder, *tree.NullItem:
		return nil
	default:
		panic(invariant.Broken("RenderEdges has no case for node %T", n))
	}
}

// RenderCounts holds, per node, the renders its costliest path down multiplies to, its own
// repeat included; RepeatCheck fills it as it reads it.
type RenderCounts map[tree.Node]int

func (m RenderCounts) renderCount(n tree.Node) int {
	if r, done := m[n]; done {
		return r
	}
	r := 1
	for _, e := range RenderEdges(n, nil) {
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

// MaxRepeat caps a repeat, and the renders nested repeats multiply to along any one path.
const MaxRepeat = 1 << 20

// RepeatCheck bounds the renders a repeat multiplies to along any root-to-leaf
// path, so nested repeats cannot build what one repeat may not.
// docs/decisions.md#the-repeat-cap-bounds-renders-not-bytes
func RepeatCheck(label string, n tree.Node, mem RenderCounts) error {
	if t, ok := n.(*tree.Template); ok && t.Repeat() > 1 && mem.renderCount(n) > MaxRepeat {
		return fmt.Errorf("%s: repeat %d multiplies to %d renders along one path, above the maximum %d", label, t.Repeat(), mem.renderCount(n), MaxRepeat)
	}
	return nil
}

// CheckNoCycles rejects a reference cycle: a node whose rendering can reach itself
// — directly, mutually, or through a chain — never terminates, so it must fail at
// New rather than stack-overflow at render. It is a depth-first walk of the render
// graph (RenderEdges). Every node is a root: a field its parent's format never
// renders is still reachable by dot path, so a cycle in one would otherwise reach
// render and be fatal there.
func CheckNoCycles(s func(fn func(label string, n tree.Node) error) error, l *tree.LoadFinds) error {
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
		for _, e := range RenderEdges(n, l) {
			if err := visit(e.To, label+" -> "+e.label); err != nil {
				return err
			}
		}
		color[n] = black
		return nil
	}
	return s(func(label string, n tree.Node) error { return visit(n, label) })
}
