package fejkdata

import (
	"fmt"
	"sort"
	"strings"
)

// heldCheck rejects every route to a held sibling name except the ones that read its
// draw. An expansion holds one draw of that name; anything else that renders it draws
// again, and the two disagree.
// docs/decisions.md#the-expansion-hold-and-the-renders-draws-are-two-fences
func heldCheck(path string, n node) error {
	t, ok := n.(*template)
	if !ok || len(t.compiled.held) == 0 {
		return nil
	}
	readers := pathKeyReaders(t.compiled.ops, t.compiled.pathKeys)
	for _, name := range heldNames(t) {
		if _, isPath := t.compiled.pathKeys[name]; isPath && isRef(name) {
			continue // held for the render: drawCheck compares every read of it, by path, across the render and its groups
		}
		if err := checkNameHeld(t, name, readers); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	return nil
}

// heldNames lists a template's held names, operands' first, then paths', each
// in name order: a level read both ways is reported by the operand's fence, and
// which overlap is reported does not vary.
func heldNames(t *template) []string {
	names := make([]string, 0, len(t.compiled.held))
	for name := range t.compiled.held {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		_, pi := t.compiled.pathKeys[names[i]]
		_, pj := t.compiled.pathKeys[names[j]]
		if pi != pj {
			return !pi
		}
		return names[i] < names[j]
	})
	return names
}

// heldNodes is what the hold of name answers for: a path pins the levels it
// passes through and the leaf it lands on, an operand exactly the value its render
// produces.
func heldNodes(t *template, name string, readers []reader) map[node]bool {
	held := map[node]bool{}
	if _, isPath := t.compiled.pathKeys[name]; !isPath {
		operandDraw(t.head(name), held)
		return held
	}
	for _, r := range readers {
		if r.a.key == name {
			coverPath(t.head(name), r.a.tail, held)
		}
	}
	return held
}

// checkNameHeld rejects every route to what name's hold pins except the readers
// holding it. One seen set across the edges: a node that cannot reach the level
// cannot reach it by another route either, so it is walked once.
func checkNameHeld(t *template, name string, readers []reader) error {
	held := heldNodes(t, name, readers)
	if len(held) == 0 {
		return nil // a fixed head holds nothing to reach
	}
	reader, isPath := t.compiled.pathKeys[name]
	seen := map[node]bool{}
	for _, e := range renderEdges(t) {
		if e.read.key == name || !renders(e.to, held, seen) {
			continue
		}
		if isPath {
			return fmt.Errorf("%s renders %q, which {%s} reads a path into; name the fields you want instead", e.reached(), name, reader)
		}
		return fmt.Errorf("%s renders %q, which a {%s()} also reads; reach it one way so it is drawn once", e.reached(), name, operandReader(t, name))
	}
	return nil
}

// operandReader names the builtin whose operand holds name.
func operandReader(t *template, name string) string {
	for _, o := range t.compiled.ops {
		for _, a := range o.operands {
			if a.key == name {
				return o.fn
			}
		}
	}
	return ""
}

// coverPath collects what holding one path pins: the first choice level it passes,
// whole, else the leaf it renders; it reads no row.
func coverPath(n node, tail []string, into map[node]bool) {
	_, _ = (&pathCover{into: into}).walk(n, tail)
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
// The walk stops at a reference edge, which is where the operand's own value ends
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
		if e.readsRef() {
			continue
		}
		operandDraw(e.to, into)
	}
}

// isFixed is a string that varies nothing: fixed text with no fields to read into.
func isFixed(n node) bool {
	t, ok := n.(*template)
	if !ok {
		return false
	}
	_, fixed := t.fixedText()
	return fixed && len(t.fields) == 0
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

// checkNoOverlap rejects a format that both renders a sibling level and reads a path
// into it — {p} beside {p.first}, {p.addr} beside {p.addr.city}. The path reads the
// level's held draw while rendering the level expands it afresh, so their values would
// disagree. Reads are compared in sorted order, so which pair is reported does not
// depend on where the tokens sit.
// docs/decisions.md#a-bare-reference-draws-each-time-a-reference-path-is-held
func checkNoOverlap(ops []op, pathKeys map[string]string) error {
	names := pathKeyReaders(ops, pathKeys)
	// Stable over one format-order scan, so two readers of one name (a token and a
	// calc operand both naming "p") are reported as the format writes them.
	sort.SliceStable(names, func(i, j int) bool { return names[i].a.path < names[j].a.path })
	for i, level := range names {
		for _, path := range names[i+1:] {
			if strings.HasPrefix(path.a.path, level.a.path+".") {
				return fmt.Errorf("%s renders a level that {%s} reads a path into; name the fields you want instead", level.label, path.a.spelling)
			}
		}
	}
	return nil
}

// reader is one way a format reaches a path key, and how to name it.
type reader struct {
	a     arm
	label string
}

// pathKeyReaders lists every way a format reaches a sibling path key, in the order the
// format writes them. An operand renders its field, so it names a level exactly
// as a token does; one scan finds both, which is what puts them in one order.
func pathKeyReaders(ops []op, pathKeys map[string]string) []reader {
	var names []reader
	for _, o := range ops {
		for _, a := range o.operands {
			if _, isPathKey := pathKeys[a.key]; isPathKey && !isRef(a.key) {
				names = append(names, reader{a, fmt.Sprintf("%s operand %q", o.fn, a.spelling)})
			}
		}
		for _, a := range o.arms {
			if _, isPathKey := pathKeys[a.key]; isPathKey && !isRef(a.key) {
				names = append(names, reader{a, "token {" + a.spelling + "}"})
			}
		}
	}
	return names
}

// checkNoRepeatedRead rejects a bare token repeated on a held name: {w} {w} beside
// {uppercase(w)} would read one draw twice, where {w} {w} alone draws twice. The
// error names the single-token spelling.
// docs/decisions.md#a-bare-reference-draws-each-time-a-reference-path-is-held
func checkNoRepeatedRead(c formatOps) error {
	count := map[string]int{}
	for _, o := range c.ops {
		for _, a := range o.arms {
			if len(a.tail) > 0 || !c.held[a.key] {
				continue
			}
			if count[a.key]++; count[a.key] > 1 {
				return fmt.Errorf("token {%s} is repeated, and %s holds %q to one draw per expansion; write {%s} once", a.spelling, c.holder[a.key], a.key, a.spelling)
			}
		}
	}
	return nil
}
