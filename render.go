package fejkdata

import (
	"fmt"
	"sort"
	"strings"
)

// rng is the randomness the renderer draws from. Passing it in keeps the render
// functions a pure core over an explicit effect; *rand.Rand satisfies it.
type rng interface {
	IntN(n int) int
	Float64() float64
}

// Fake generates a value for a dot path. Each segment descends one level: folder
// names and the category (JSON file) come first, then named fields within it,
// e.g. "sv_SE.address" or "sv_SE.address.street". Choices along the way are
// resolved at random. A path naming a folder (no value of its own) is an error.
func (f *Generator) Fake(path string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	n, err := descend(f.rand, &group{children: f.categories}, strings.Split(path, "."))
	if err != nil {
		return "", fmt.Errorf("fejkdata: %s: %w", path, err)
	}
	if _, ok := n.(*group); ok {
		return "", fmt.Errorf("fejkdata: %s names a folder, not a value", path)
	}
	return render(f.rand, n, nil), nil
}

// descend walks named fields to the node a path names. It is the one render-side
// step that can fail, because the path comes from the caller and may name a field
// that does not exist. A choice consumes no segment, so the rest of the path must
// be one every variant carries before a variant is picked — a path that resolves
// at all resolves on every call.
func descend(s *session, root node, segments []string) (node, error) {
	var found node
	err := walkPath(root, segments, pathWalk{
		choice: func(c *choice, rest []string) ([]node, error) {
			if err := carriedByAll(c, rest); err != nil {
				return nil, err
			}
			return []node{pick(s, c)}, nil
		},
		leaf: func(n node) error { found = n; return nil },
	})
	return found, err
}

// render evaluates a compiled node to a string. compile validates every node up
// front, so rendering a compiled tree cannot fail. refScope carries the draws a
// reference shares beyond its own expansion; nil keeps every reference local.
func render(s *session, n node, refScope *draws) string {
	switch n := n.(type) {
	case *choice:
		return render(s, pick(s, n), refScope)
	case *null:
		return ""
	case *template:
		if n.repeat == 1 {
			if n.fixed {
				return n.lit
			}
			return expand(s, n, refScope)
		}
		var b strings.Builder
		b.Grow(n.repeat * (n.grow + len(n.separator)))
		for i := 0; i < n.repeat; i++ {
			if i > 0 {
				b.WriteString(n.separator)
			}
			b.WriteString(expand(s, n, refScope))
		}
		return b.String()
	default:
		panic(fmt.Sprintf("fejkdata: uncompiled node %T", n))
	}
}

// pick selects one item. Uniform choices are O(1); weighted choices are an
// O(log n) search over precomputed cumulative weights. compile guarantees a
// non-empty choice and a finite positive total, so the index is always in range.
func pick(r rng, c *choice) node {
	if c.cum == nil {
		return c.items[r.IntN(len(c.items))]
	}
	x := r.Float64() * c.cum[len(c.cum)-1]
	i := sort.Search(len(c.cum), func(i int) bool { return c.cum[i] > x })
	return c.items[i]
}

// expand renders a template's compiled ops. compile validated every token, so this
// cannot fail.
func expand(s *session, t *template, refScope *draws) string {
	var b strings.Builder
	b.Grow(t.grow)
	// One draw per held name, for this expansion only: a nested template and each
	// repeat iteration get their own, since each is its own expansion. A reference
	// reads the caller's scope instead, whenever one was supplied.
	var held *draws
	if len(t.held) > 0 {
		held = &draws{
			variant: make(map[string]node, len(t.held)),
			value:   make(map[string]string, len(t.held)),
		}
	}
	for i := range t.ops {
		o := &t.ops[i]
		switch o.kind {
		case 'l':
			b.WriteString(o.lit)
		case 'f':
			b.WriteString(readField(s, t, held, refScope, o.arms[s.IntN(len(o.arms))]))
		case 'b':
			// Read before the call, so the value a calc computes is the value the
			// format showed. calcVars fixed the order op.operands holds.
			var operands []string
			if len(o.operands) > 0 {
				operands = make([]string, len(o.operands))
				for j, a := range o.operands {
					operands[j] = readField(s, t, held, refScope, a)
				}
			}
			b.WriteString(o.call(s, b.String(), operands)) // b.String() is the output so far
		}
	}
	return b.String()
}
