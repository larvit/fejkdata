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
	return render(f.rand, n), nil
}

// Template is an inline template compiled, referenced and validated against a
// generator's loaded data once, ready to render many times with [Template.Fake].
// It is safe for concurrent use: Fake serializes on its generator's lock, so a
// seeded sequence is reproducible only when a generator — and its templates — are
// drawn from one goroutine.
type Template struct {
	g *Generator
	n node
}

// Fake renders the template with one draw.
func (t *Template) Fake() string {
	t.g.mu.Lock()
	defer t.g.mu.Unlock()
	return render(t.g.rand, t.n)
}

// NewTemplate compiles an inline template — a format string or a JSON value — and
// binds its references against the loaded tree, so repeated renders pay the
// compile and validation once. It shares [New]'s guarantees: a bad template errors
// here, and rendering cannot fail.
// NewTemplate runs the same fences loadData does for a category, scoped to one
// inline node with everything but checkNoCycles: a reference binds only into the
// loaded tree, which has no path into this node, so rendering it cannot reach
// itself. A new fence belongs in both places (loadData and here).
func (f *Generator) NewTemplate(input string) (*Template, error) {
	n, err := compileInput(input)
	if err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	if err := linkNodeRefs(n, f.categories); err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	if err := checkNodeRepeatReach(n); err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	if err := checkNodeBoundLevelsHeld(n); err != nil {
		return nil, fmt.Errorf("fejkdata: %w", err)
	}
	return &Template{g: f, n: n}, nil
}

// FakeTemplate compiles and renders an inline template in one call. It is
// [NewTemplate] then [Template.Fake]; to render the same template many times, hold
// the *Template and call its Fake.
func (f *Generator) FakeTemplate(input string) (string, error) {
	t, err := f.NewTemplate(input)
	if err != nil {
		return "", err
	}
	return t.Fake(), nil
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
// front, so rendering a compiled tree cannot fail.
func render(s *session, n node) string {
	switch n := n.(type) {
	case *choice:
		return render(s, pick(s, n))
	case *template:
		if n.repeat == 1 {
			if n.fixed {
				return n.lit
			}
			return expand(s, n)
		}
		var b strings.Builder
		b.Grow(n.repeat * (n.grow + len(n.separator)))
		for i := 0; i < n.repeat; i++ {
			if i > 0 {
				b.WriteString(n.separator)
			}
			b.WriteString(expand(s, n))
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
func expand(s *session, t *template) string {
	var b strings.Builder
	b.Grow(t.grow)
	// One draw per held name, for this expansion only: a nested template and each
	// repeat iteration get their own, since each is its own expansion.
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
			b.WriteString(readField(s, t, held, o.arms[s.IntN(len(o.arms))]))
		case 'b':
			// Read before the call, so the value a calc computes is the value the
			// format showed. calcVars fixed the order op.operands holds.
			var operands []string
			if len(o.operands) > 0 {
				operands = make([]string, len(o.operands))
				for j, a := range o.operands {
					operands[j] = readField(s, t, held, a)
				}
			}
			b.WriteString(o.call(s, b.String(), operands)) // b.String() is the output so far
		}
	}
	return b.String()
}
