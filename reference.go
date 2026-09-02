package fejkdata

import (
	"fmt"
	"sort"
	"strings"
)

// refPrefix marks a {..path} token: a reference to a node elsewhere in the data
// root rather than a sibling field. The path is resolved across every loaded
// directory (see linkRefs).
const refPrefix = ".."

func isRef(name string) bool { return strings.HasPrefix(name, refPrefix) }

// linkRefs resolves every {..path} reference in the assembled tree. The head of the
// path — up to the category it names — is bound into the referring template's
// fields, and the rest reads into it the way a sibling path does, so a reference
// is held like a sibling. It runs once, after all data is merged, so a reference
// sees the final (override-resolved) tree. A path that is unknown, names a folder,
// or reads a field not every variant carries fails here, keeping a bad reference a
// New-time error, never a random render-time one.
func linkRefs(root map[string]node) error {
	return walkNodes(root, func(path string, n node) error {
		t, ok := n.(*template)
		if !ok {
			return nil
		}
		names := refTokens(t.format)
		if len(names) == 0 {
			return nil
		}
		if t.fields == nil {
			t.fields = map[string]node{}
		}
		t.refs = make(map[string]string, len(names))
		for _, name := range names {
			head, target, tail, err := resolveRef(root, strings.Split(name[len(refPrefix):], "."))
			if err != nil {
				return fmt.Errorf("%s: reference {%s}: %w", path, name, err)
			}
			key := refPrefix + strings.Join(head, ".")
			if err := checkPath(target, tail, key); err != nil {
				return fmt.Errorf("%s: reference {%s}: %w", path, name, err)
			}
			t.fields[key] = target
			t.refs[name] = key
		}
		t.compileFormat()
		if err := checkNoOverlap(t.format, t.bound, t.refs); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		return nil
	})
}

// checkBoundLevelsHeld rejects every route to a held name except the ones that read
// its draw. An expansion holds one draw of that name; anything else that renders it
// draws again, and the two disagree. checkNoOverlap settles the spellings within one
// format (a token, an operand); this settles the rest — a reference, whether it
// sits in that format or in anything the format renders, however deep.
//
// It runs after checkNoCycles, whose guarantee is what lets the walk terminate.
func checkBoundLevelsHeld(root map[string]node) error {
	return walkNodes(root, func(path string, n node) error {
		t, ok := n.(*template)
		if !ok || len(t.held) == 0 {
			return nil
		}
		heads := make([]string, 0, len(t.held))
		for head := range t.held {
			heads = append(heads, head)
		}
		// Operand heads first, then paths, each in name order: a level read both
		// ways is reported by the operand's fence, and which overlap is reported
		// does not vary.
		sort.Slice(heads, func(i, j int) bool {
			_, pi := t.bound[heads[i]]
			_, pj := t.bound[heads[j]]
			if pi != pj {
				return !pi
			}
			return heads[i] < heads[j]
		})
		readers := boundReaders(t.format, t.bound, t.refs)
		for _, head := range heads {
			// What one draw answers for depends on how the draw is read: a path pins
			// the levels it passes through and the leaf it lands on, an operand
			// exactly the value its render produces.
			held := map[node]bool{}
			reader, isPath := t.bound[head]
			if isPath {
				for _, r := range readers {
					if a := splitArm(r.name, t.refs); a.key == head {
						coverPath(t.fields[head], a.tail, held)
					}
				}
			} else {
				operandDraw(t.fields[head], held)
			}
			if len(held) == 0 {
				continue // an early out: a fixed head holds nothing to reach
			}
			// One seen set across the edges: a node that cannot reach the level
			// cannot reach it by another route either, so it is walked once here.
			seen := map[node]bool{}
			for _, e := range renderEdges(t) {
				if splitArm(e.label, t.refs).key == head {
					continue // a token or operand reading this draw, the routes allowed
				}
				if renders(e.to, held, seen) {
					if isPath {
						return fmt.Errorf("%s: %s renders %q, which {%s} reads a path into; name the fields you want instead", path, e.reached(), head, reader)
					}
					return fmt.Errorf("%s: %s renders %q, which a {%s()} also reads; reach it one way so it is drawn once", path, e.reached(), head, operandReader(t, head))
				}
			}
		}
		return nil
	})
}

// operandReader names the builtin whose operand holds head.
func operandReader(t *template, head string) string {
	fn := ""
	_ = eachToken(t.format, func(tok ftoken) error {
		if tok.kind != 'b' || fn != "" {
			return nil
		}
		if name, _, isFunc := funcCall(tok.body); isFunc {
			for _, operand := range tokenOperands(tok.body) {
				if splitArm(operand, t.refs).key == head {
					fn = name
				}
			}
		}
		return nil
	})
	return fn
}

// coverPath collects what holding one path pins: every choice level the path
// passes through, whole, and the leaf it renders.
func coverPath(n node, tail []string, into map[node]bool) {
	if _, isChoice := n.(*choice); isChoice || len(tail) == 0 {
		cover(n, into, false)
		return
	}
	t, ok := n.(*template)
	if !ok {
		return
	}
	if child, ok := t.fields[tail[0]]; ok {
		coverPath(child, tail[1:], into)
	}
}

// cover collects a level and everything contained in it. A fixed string outside a
// choice is left out — it cannot disagree with itself — but inside one each
// variant carries its own, so there it counts.
func cover(n node, into map[node]bool, inChoice bool) {
	if isFixed(n) && !inChoice {
		return
	}
	into[n] = true
	_, isChoice := n.(*choice)
	for _, c := range contained(n) {
		cover(c.node, into, inChoice || isChoice)
	}
}

// operandDraw collects what one held draw of an operand answers for: the operand
// and what rendering it settles inside itself. The builtin renders its operand
// whole, so that draw fixes every value the render produced, and a second route to
// any of them disagrees with it.
//
// The walk stops at a {..path} edge, which is where the operand's own value ends
// and a shared source begins: two names referencing one category are two draws, the
// same rule {word} {word} follows.
func operandDraw(n node, into map[node]bool) {
	if isFixed(n) {
		return
	}
	if into[n] {
		return
	}
	into[n] = true
	for _, e := range renderEdges(n) {
		if isRef(e.label) {
			continue
		}
		operandDraw(e.to, into)
	}
}

// isFixed is a string that varies nothing: fixed text with no fields to read into.
func isFixed(n node) bool {
	t, ok := n.(*template)
	return ok && t.fixed && len(t.fields) == 0
}

// renders reports whether rendering n can reach anything in want, following the
// same edges expand does. seen keeps a node shared by several routes from being
// walked twice; checkNoCycles has already proved the graph is a DAG, so the walk
// ends.
func renders(n node, want, seen map[node]bool) bool {
	if want[n] {
		return true
	}
	if seen[n] {
		return false
	}
	seen[n] = true
	for _, e := range renderEdges(n) {
		if renders(e.to, want, seen) {
			return true
		}
	}
	return false
}

// walkNodes calls fn once per contained node, passing the dot path that reaches it,
// visiting keys in sorted order so which of several broken nodes gets reported does
// not depend on map iteration.
func walkNodes(root map[string]node, fn func(path string, n node) error) error {
	seen := map[node]bool{}
	var visit func(string, node) error
	visit = func(path string, n node) error {
		if n == nil || seen[n] {
			return nil
		}
		seen[n] = true
		if err := fn(path, n); err != nil {
			return err
		}
		for _, c := range contained(n) {
			if err := visit(join(path, c.name), c.node); err != nil {
				return err
			}
		}
		return nil
	}
	for _, name := range sortedNames(root) {
		if err := visit(name, root[name]); err != nil {
			return err
		}
	}
	return nil
}

// namedNode is a contained child and the segment reaching it; a choice's items carry
// no segment, matching how a dot path steps over a choice.
type namedNode struct {
	name string
	node node
}

func contained(n node) []namedNode {
	switch n := n.(type) {
	case *group:
		return named(n.children)
	case *choice:
		out := make([]namedNode, len(n.items))
		for i, it := range n.items {
			out[i] = namedNode{node: it}
		}
		return out
	case *template:
		return named(n.fields)
	default:
		return nil
	}
}

// named skips a bound {..path} key: it is a render edge, not containment, so using
// it as a path segment would report a node under a path that does not reach it. Only
// a template's fields hold bindings — loadDir skips a dot-prefixed entry, so a
// group's children never carry the prefix — so this one skip serves both.
func named(m map[string]node) []namedNode {
	out := make([]namedNode, 0, len(m))
	for _, name := range sortedNames(m) {
		if isRef(name) {
			continue
		}
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

// resolveRef walks a reference path through the folders to the category it names,
// returning that head, the node, and the tail left to read into it.
func resolveRef(root map[string]node, segments []string) (head []string, target node, tail []string, err error) {
	var n node = &group{children: root}
	i := 0
	for ; i < len(segments); i++ {
		g, ok := n.(*group)
		if !ok {
			break
		}
		child, ok := g.children[segments[i]]
		if !ok {
			return nil, nil, nil, fmt.Errorf("no entry %q", segments[i])
		}
		n = child
	}
	if _, ok := n.(*group); ok {
		return nil, nil, nil, fmt.Errorf("names a folder, not a value")
	}
	return segments[:i], n, segments[i:], nil
}

// refTokens returns the {..path} names a format reads, as tokens or as operands.
func refTokens(format string) []string {
	var refs []string
	seen := map[string]bool{}
	for _, name := range append(fieldTokens(format), operandTokens(format)...) {
		if isRef(name) && !seen[name] {
			seen[name] = true
			refs = append(refs, name)
		}
	}
	return refs
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
// group renders nothing, so it has no edges.
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
		add := func(name, operand string) {
			a := splitArm(name, n.refs)
			c, ok := n.fields[a.key]
			if !ok {
				return
			}
			for _, leaf := range pathLeaves(c, a.tail) {
				es = append(es, renderEdge{leaf, name, operand})
			}
		}
		_ = eachToken(n.format, func(t ftoken) error {
			if t.kind != 'b' {
				return nil
			}
			if fn, _, isFunc := funcCall(t.body); isFunc {
				for _, operand := range tokenOperands(t.body) {
					add(operand, fn)
				}
				return nil
			}
			for _, name := range strings.Split(t.body, "|") {
				add(name, "")
			}
			return nil
		})
		return es
	default:
		return nil
	}
}

// pathLeaves lists what a token's dotted tail renders. A path draws the levels it
// passes through but renders only what it lands on, so the leaf is the edge — a
// bare token, whose tail is empty, lands on the field itself. A choice on the way
// contributes every variant, since any of them may be the one drawn. checkPath has
// already proved the tail resolves in every variant, so the walk drops nothing.
func pathLeaves(n node, tail []string) []node {
	if len(tail) == 0 {
		return []node{n}
	}
	if c, ok := n.(*choice); ok {
		var out []node
		for _, it := range c.items {
			out = append(out, pathLeaves(it, tail)...)
		}
		return out
	}
	return pathLeaves(child(n, tail[0]), tail[1:])
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
