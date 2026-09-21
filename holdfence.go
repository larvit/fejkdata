package fejkdata

import (
	"fmt"
	"sort"
	"strings"
)

// heldCheck rejects every route to a held sibling name except the ones that read its
// draw. An expansion holds one draw of that name; anything else that renders it draws
// again, and the two disagree.
func heldCheck(path string, n node) error {
	t, ok := n.(*template)
	if !ok || len(t.held) == 0 {
		return nil
	}
	readers := boundReaders(t.format, t.bound, t.refs)
	for _, head := range heldHeads(t) {
		if _, isPath := t.bound[head]; isPath && isRef(head) {
			continue
		}
		if err := checkHeadHeld(t, head, readers); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	return nil
}

// heldHeads lists a template's held names, operand heads first, then paths, each
// in name order: a level read both ways is reported by the operand's fence, and
// which overlap is reported does not vary.
func heldHeads(t *template) []string {
	heads := make([]string, 0, len(t.held))
	for head := range t.held {
		heads = append(heads, head)
	}
	sort.Slice(heads, func(i, j int) bool {
		_, pi := t.bound[heads[i]]
		_, pj := t.bound[heads[j]]
		if pi != pj {
			return !pi
		}
		return heads[i] < heads[j]
	})
	return heads
}

// pinned collects what the hold of head answers for: a path pins the levels it
// passes through and the leaf it lands on, an operand exactly the value its render
// produces.
func pinned(t *template, head string, readers []reader) map[node]bool {
	held := map[node]bool{}
	if _, isPath := t.bound[head]; !isPath {
		operandDraw(t.fields[head], held)
		return held
	}
	for _, r := range readers {
		if a := splitArm(r.name, t.refs); a.key == head {
			coverPath(t.fields[head], a.tail, held)
		}
	}
	return held
}

// checkHeadHeld rejects every route to what head's hold pins except the readers
// holding it. One seen set across the edges: a node that cannot reach the level
// cannot reach it by another route either, so it is walked once.
func checkHeadHeld(t *template, head string, readers []reader) error {
	held := pinned(t, head, readers)
	if len(held) == 0 {
		return nil // a fixed head holds nothing to reach
	}
	reader, isPath := t.bound[head]
	seen := map[node]bool{}
	for _, e := range renderEdges(t) {
		if splitArm(e.label, t.refs).key == head || !renders(e.to, held, seen) {
			continue
		}
		if isPath {
			return fmt.Errorf("%s renders %q, which {%s} reads a path into; name the fields you want instead", e.reached(), head, reader)
		}
		return fmt.Errorf("%s renders %q, which a {%s()} also reads; reach it one way so it is drawn once", e.reached(), head, operandReader(t, head))
	}
	return nil
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
	_, _ = walkPath(n, tail, pathWalk{
		choice: func(c *choice, _ []string) ([]node, error) { cover(c, into, false); return nil, nil },
		leaf:   func(n node) error { cover(n, into, false); return nil },
	})
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

// checkNoOverlap rejects a format that both renders a sibling level and reads a path
// into it — {p} beside {p.first}, {p.addr} beside {p.addr.city}. The path reads the
// level's held draw while rendering the level expands it afresh, so their values would
// disagree. Reads are compared in sorted order, so which pair is reported does not
// depend on where the tokens sit.
func checkNoOverlap(format string, bound map[string]string, refs map[string]refBinding) error {
	names := boundReaders(format, bound, refs)
	// Stable over one format-order scan, so two readers of one name (a token and a
	// calc operand both naming "p") are reported as the format writes them.
	sort.SliceStable(names, func(i, j int) bool { return names[i].path < names[j].path })
	for i, level := range names {
		for _, path := range names[i+1:] {
			if strings.HasPrefix(path.path, level.path+".") {
				return fmt.Errorf("%s renders a level that {%s} reads a path into; name the fields you want instead", level.label, path.name)
			}
		}
	}
	return nil
}

// reader is one way a format reaches a bound field: as written, by its one
// spelling, and how to name it.
type reader struct{ name, path, label string }

// boundReaders lists every way a format reaches a bound sibling field, in the order the
// format writes them. An operand renders its field, so it names a level exactly
// as a token does; one scan finds both, which is what puts them in one order.
func boundReaders(format string, bound map[string]string, refs map[string]refBinding) []reader {
	var names []reader
	_ = eachToken(format, func(t ftoken) error {
		if t.kind != 'b' {
			return nil
		}
		if fn, _, isFunc := funcCall(t.body); isFunc {
			for _, operand := range tokenOperands(t.body) {
				a := splitArm(operand, refs)
				if _, isBound := bound[a.key]; isBound && !isRef(a.key) {
					names = append(names, reader{a.name, a.path, fmt.Sprintf("%s operand %q", fn, operand)})
				}
			}
			return nil
		}
		for _, a := range splitArms(t.body, refs) {
			if _, isBound := bound[a.key]; isBound && !isRef(a.key) {
				names = append(names, reader{a.name, a.path, "token {" + a.name + "}"})
			}
		}
		return nil
	})
	return names
}

// checkNoRepeatedRead rejects a bare token repeated on a held name: {w} {w} beside
// {uppercase(w)} would read one draw twice, where {w} {w} alone draws twice. The
// error names the single-token spelling.
func checkNoRepeatedRead(format string, c formatOps, refs map[string]refBinding) error {
	count := map[string]int{}
	return eachToken(format, func(t ftoken) error {
		if t.kind != 'b' {
			return nil
		}
		if _, _, isFunc := funcCall(t.body); isFunc {
			return nil
		}
		for _, a := range splitArms(t.body, refs) {
			if len(a.tail) > 0 || !c.held[a.key] {
				continue
			}
			if count[a.key]++; count[a.key] > 1 {
				return fmt.Errorf("token {%s} is repeated, and %s holds %q to one draw per expansion; write {%s} once", a.name, c.holder[a.key], a.key, a.name)
			}
		}
		return nil
	})
}
