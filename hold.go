package fejkdata

import (
	"fmt"
	"sort"
	"strings"
)

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
	_ = walkPath(n, tail, pathWalk{
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

// arm is one alternative of a {a|b} token or one operand, split into the key
// naming the node in a template's fields (a sibling field, or the head a
// reference is bound under) and the tail of a dotted path into it. A non-empty
// tail is what makes the arm a bound draw: its head is drawn once per expansion
// (see compileOps).
type arm struct {
	name  string // as written, and the key a bound draw's value is held under
	key   string
	tail  []string
	steps []string // key per level passed through; the head and leaf hold their own
}

// splitArm splits one name into key and tail. refs maps a reference to what
// linkRefs bound it to; before linking, a reference is whole.
func splitArm(name string, refs map[string]refBinding) arm {
	if isRef(name) {
		b, bound := refs[name]
		if !bound || len(b.tail) == 0 {
			key := name
			if bound {
				key = b.key
			}
			return arm{name: name, key: key}
		}
		return pathArm(name, b.key, b.tail)
	}
	head, tail, dotted := strings.Cut(name, ".")
	if !dotted {
		return arm{name: name, key: name}
	}
	return pathArm(name, head, strings.Split(tail, "."))
}

func pathArm(name, key string, segs []string) arm {
	var steps []string
	for i := 0; i < len(segs)-1; i++ { // every level except the leaf's own
		steps = append(steps, key+"."+strings.Join(segs[:i+1], "."))
	}
	return arm{name: name, key: key, tail: segs, steps: steps}
}

// checkNoOverlap rejects a format that both renders a level and reads a path into
// it — {p} beside {p.first}, or {p.addr} beside {p.addr.city}. The path reads the
// level's held draw while rendering the level expands it afresh, so their values
// would disagree. Names are compared in sorted order, so which pair is reported
// does not depend on where the tokens sit.
func checkNoOverlap(format string, bound map[string]string, refs map[string]refBinding) error {
	names := boundReaders(format, bound, refs)
	// Stable over one format-order scan, so two readers of one name (a token and a
	// calc operand both naming "p") are reported as the format writes them.
	sort.SliceStable(names, func(i, j int) bool { return names[i].name < names[j].name })
	for i, level := range names {
		for _, path := range names[i+1:] {
			if strings.HasPrefix(path.name, level.name+".") {
				return fmt.Errorf("%s renders a level that {%s} reads a path into; name the fields you want instead", level.label, path.name)
			}
		}
	}
	return nil
}

// reader is one way a format reaches a bound field, and how to name that spelling.
type reader struct{ name, label string }

// boundReaders lists every way a format reaches a bound field, in the order the
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
				if _, isBound := bound[a.key]; isBound {
					names = append(names, reader{a.name, fmt.Sprintf("%s operand %q", fn, operand)})
				}
			}
			return nil
		}
		for _, a := range splitArms(t.body, refs) {
			if _, isBound := bound[a.key]; isBound {
				names = append(names, reader{a.name, "token {" + a.name + "}"})
			}
		}
		return nil
	})
	return names
}

// splitArms splits a token body's '|' alternatives.
func splitArms(body string, refs map[string]refBinding) []arm {
	parts := strings.Split(body, "|")
	arms := make([]arm, len(parts))
	for i, p := range parts {
		arms[i] = splitArm(p, refs)
	}
	return arms
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

// draws is what an expansion has already drawn for its held names: the variant each
// was drawn as, so every path under it reads one row, and the value each read, so
// the same name read twice reads one value.
type draws struct {
	variant map[string]node
	value   map[string]string
}

// readField renders one arm of a token. An arm's key is a sibling field or a
// reference linkRefs bound into fields. A name the expansion holds — a level some
// token addresses by dotted path, or a field an operand reads — is drawn once and
// kept, so {place.postal-code} and {place.locality} read one row, either read twice
// gives one value, and a shown operand is the operand computed. Every other name is
// drawn afresh, so {word} {word} still draws twice. checkTokens, checkPath and
// linkRefs prove every step, so the walk cannot fail.
func readField(s *session, t *template, held *draws, a arm) string {
	if !t.held[a.key] {
		return render(s, t.fields[a.key])
	}
	if v, read := held.value[a.name]; read {
		return v
	}
	var v string
	_ = walkPath(t.fields[a.key], a.tail, pathWalk{
		// Hold the draw at every level passed through, so two paths sharing a
		// prefix share it.
		choice: func(c *choice, rest []string) ([]node, error) {
			key := a.key
			if consumed := len(a.tail) - len(rest); consumed > 0 {
				key = a.steps[consumed-1]
			}
			n, drew := held.variant[key]
			if !drew {
				n = drawn(s, c)
				held.variant[key] = n
			}
			return []node{n}, nil
		},
		leaf: func(n node) error { v = render(s, n); return nil },
	})
	held.value[a.name] = v
	return v
}

// drawn resolves a choice to one variant, so a bound head is a concrete node the
// rest of the expansion shares. Nested choices unwrap too: a draw is one value, not
// another set to pick from.
func drawn(s *session, n node) node {
	for c, ok := n.(*choice); ok; c, ok = n.(*choice) {
		n = pick(s, c)
	}
	return n
}
