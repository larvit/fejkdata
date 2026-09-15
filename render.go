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
	return renderOnce(f.rand, n), nil
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
// front, so rendering a compiled tree cannot fail. sc holds the reference draws the
// render shares; each repeat iteration renders over draws of its own.
func render(s *session, n node, sc drawScope) string {
	switch n := n.(type) {
	case *choice:
		return render(s, pick(s, n), sc)
	case *null:
		return ""
	case *template:
		sc = sc.in(n)
		if n.repeat == 1 {
			if n.fixed {
				return n.lit
			}
			return expand(s, n, sc)
		}
		var b strings.Builder
		b.Grow(n.repeat * (n.grow + len(n.separator)))
		for i := 0; i < n.repeat; i++ {
			if i > 0 {
				b.WriteString(n.separator)
			}
			b.WriteString(expandAnew(s, n, sc.group))
		}
		return b.String()
	default:
		panic(fmt.Sprintf("fejkdata: uncompiled node %T", n))
	}
}

// expandAnew expands one repeat iteration of t as a render of its own, in group. Inlined into
// render's loop, its draw set would move to the heap.
//
//go:noinline
func expandAnew(s *session, t *template, group string) string {
	var set drawSet
	return expand(s, t, drawScope{set: &set, group: group})
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
func expand(s *session, t *template, sc drawScope) string {
	var b strings.Builder
	b.Grow(t.grow)
	// One draw per held name, for this expansion only: a nested template and each
	// repeat iteration get their own, since each is its own expansion. A reference
	// path reads the render's draws in sc instead.
	var held *draws
	if len(t.held) > 0 {
		held = &draws{
			variant: make(map[string]node, len(t.held)),
			value:   make(map[string]draw, len(t.held)),
		}
	}
	for i := range t.ops {
		o := &t.ops[i]
		switch o.kind {
		case 'l':
			b.WriteString(o.lit)
		case 'f':
			b.WriteString(readField(s, t, held, sc, o.arms[s.IntN(len(o.arms))]).text)
		case 'b':
			// Read before the call, so the value a calc computes is the value the
			// format showed. calcVars fixed the order op.operands holds.
			var operands []string
			if len(o.operands) > 0 {
				operands = make([]string, len(o.operands))
				for j, a := range o.operands {
					operands[j] = readField(s, t, held, sc, a).text
				}
			}
			b.WriteString(o.call(s, b.String(), operands)) // b.String() is the output so far
		}
	}
	return b.String()
}
