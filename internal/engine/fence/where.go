package fence

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

// refuseFilteredTails refuses a path through a name holding a filtered draw, which is read whole.
func refuseFilteredTails(ts []load.TemplateSite) error {
	for _, s := range ts {
		for _, r := range load.NamedReads(s.Template, nil) {
			if r.Read.Named().Values() != nil && len(r.Read.Tail()) > 0 {
				return fmt.Errorf("%s: token {%s}: name %q holds a filtered draw, which is read whole: {%s}", s.Label, r.Op.Token().Body, r.Read.Head(), r.Read.Head())
			}
		}
	}
	return nil
}

// refuseSelfReads refuses a token whose where clause reads the name the token binds, directly or
// through what the names it reads draw: drawing the name would read itself.
func refuseSelfReads(ts []load.TemplateSite) error {
	for _, s := range ts {
		for _, o := range s.Template.Compiled().Ops() {
			if o.Token().Kind != grammar.NameBind || o.Draw() == nil {
				continue
			}
			b := s.Template.NameScope().Bindings()[o.Token().Bound]
			r := selfRead{b: b, names: map[*tree.NameBinding]bool{}, nodes: map[tree.Node]bool{}}
			for _, a := range o.Clause() {
				if !r.reaches(a) {
					continue
				}
				via := ""
				if a.Named() != b {
					via = fmt.Sprintf(", and drawing %q reads %q", a.Head(), b.Name())
				}
				return fmt.Errorf("%s: token {%s}: its where clause reads %q%s, the name it binds, so the draw would read itself; bind what the clause reads to another name", s.Label, o.Token().Body, a.Head(), via)
			}
		}
	}
	return nil
}

// selfRead walks what a read draws, looking for a read of b.
type selfRead struct {
	b     *tree.NameBinding
	names map[*tree.NameBinding]bool
	nodes map[tree.Node]bool
}

// reaches says whether reading a reads b: a name holding a filtered draw reads only its clause's
// names, since its values are fixed at load; any other read renders its leaves.
func (r selfRead) reaches(a tree.Read) bool {
	if a.Kind() == tree.NamedRead {
		if a.Named() == r.b {
			return true
		}
		if r.names[a.Named()] {
			return false
		}
		r.names[a.Named()] = true
		if v := a.Named().Values(); v != nil {
			return slices.ContainsFunc(v.Clause(), r.reaches)
		}
	}
	return slices.ContainsFunc(a.Leaves(), r.reachesNode)
}

func (r selfRead) reachesNode(n tree.Node) bool {
	if r.nodes[n] {
		return false
	}
	r.nodes[n] = true
	t, ok := n.(*tree.Template)
	if !ok {
		return slices.ContainsFunc(load.RenderEdges(n, nil), func(e load.RenderEdge) bool { return r.reachesNode(e.To) })
	}
	for _, o := range t.Compiled().Ops() {
		if slices.ContainsFunc(o.AllReads(), func(a *tree.Read) bool { return r.reaches(*a) }) {
			return true
		}
	}
	return false
}

// lengthProof is what proves a draw's length: the lengths of its values, shortest first, the
// clause's ends, and how many values one draw takes.
type lengthProof struct {
	lengths []int
	lo, hi  tree.LengthEnd
	count   int
	filters bool // a where clause filters the values
}

// proveDrawLengths refuses a draw in ts that some render leaves with too few values.
func proveDrawLengths(ts []load.TemplateSite) error {
	for _, s := range ts {
		for _, o := range s.Template.Compiled().Ops() {
			p, draws := drawLengthProof(o)
			if !draws {
				continue
			}
			if err := p.proveEveryDrawPasses(); err != nil {
				return fmt.Errorf("%s: token {%s}: %w", s.Label, o.Token().Body, err)
			}
		}
	}
	return nil
}

// drawLengthProof is what proves the length of o's draw, where o draws through a where or a
// distinct clause.
func drawLengthProof(o tree.Op) (lengthProof, bool) {
	switch d := o.Draw().(type) {
	case *tree.ValueDraw:
		return lengthProof{d.Index().Lengths(), d.Lo(), d.Hi(), d.Count(), d.Filters()}, true
	case *tree.AlphabetDraw:
		p := lengthProof{lengths: []int{d.Runs().Len()}, count: 1, filters: true}
		p.lo, p.hi = load.LengthEnds(o.Token().Where, o.Clause())
		return p, true
	}
	return lengthProof{}, false
}

// proveEveryDrawPasses refuses a draw that some render leaves with too few values: one where too few
// values pass, or one whose length reads a name that may print a length too few values have.
func (v lengthProof) proveEveryDrawPasses() error {
	bounds := map[*tree.Read][2]float64{}
	for _, e := range []tree.LengthEnd{v.lo, v.hi} {
		if e.Read() == nil {
			continue
		}
		p := (&valueProof{}).proveRead(*e.Read())
		switch {
		case p.NotOperand != "":
			return fmt.Errorf("length %s is not always a whole number: %s", e.Read().Spelling(), p.NotOperand)
		case !p.Integral:
			return fmt.Errorf("length %s is not always a whole number: it may print a fraction", e.Read().Spelling())
		}
		bounds[e.Read()] = [2]float64{p.Lo, p.Hi}
	}
	if v.lo.Read() != nil && v.lo.Read() == v.hi.Read() {
		for n := bounds[v.lo.Read()][0]; n <= bounds[v.lo.Read()][1]; n++ {
			if passing := v.passing(n, n); passing < v.count {
				return fmt.Errorf("where %s is %.0f, %s", v.lo.Read().Spelling(), n, v.shortfall(passing))
			}
		}
		return nil
	}
	lo, hi := narrowest(v.lo, v.hi, bounds)
	if passing := v.passing(lo, hi); passing < v.count {
		return v.rangeShortfall(lo, hi, passing)
	}
	return nil
}

// narrowest is the narrowest length lo to hi a render can read, given the bounds of what each
// name read prints. Every render must pass, so the low end takes its name's highest value and the
// high end its lowest, which may cross.
func narrowest(lo, hi tree.LengthEnd, bounds map[*tree.Read][2]float64) (float64, float64) {
	from, to := float64(lo.N()), float64(hi.N())
	if lo.Read() != nil {
		from = bounds[lo.Read()][1]
	}
	if hi.Read() != nil {
		to = bounds[hi.Read()][0]
	}
	return from, to
}

// rangeShortfall says that the length lo to hi leaves passing values, too few, naming the length's
// names at the values that narrow it to lo to hi.
func (v lengthProof) rangeShortfall(lo, hi float64, passing int) error {
	var at []string
	if v.lo.Read() != nil {
		at = append(at, fmt.Sprintf("%s is %.0f", v.lo.Read().Spelling(), lo))
	}
	if v.hi.Read() != nil {
		at = append(at, fmt.Sprintf("%s is %.0f", v.hi.Read().Spelling(), hi))
	}
	switch {
	case len(at) == 0:
		return errors.New(v.shortfall(passing))
	case lo <= hi:
		return fmt.Errorf("where %s, %s", strings.Join(at, " and "), v.shortfall(passing))
	case v.hi.Read() == nil:
		return fmt.Errorf("where %s, length %.0f to %.0f holds no length; keep %s at or below %d", at[0], lo, hi, v.lo.Read().Spelling(), v.hi.N())
	case v.lo.Read() == nil:
		return fmt.Errorf("where %s, length %.0f to %.0f holds no length; keep %s at %d or above", at[0], lo, hi, v.hi.Read().Spelling(), v.lo.N())
	}
	return fmt.Errorf("where %s, length %.0f to %.0f holds no length; keep %s at or below %s", strings.Join(at, " and "), lo, hi, v.lo.Read().Spelling(), v.hi.Read().Spelling())
}

// shortfall says that passing values are fewer than a draw takes.
func (v lengthProof) shortfall(passing int) string {
	n := v.count
	switch {
	case !v.filters:
		return fmt.Sprintf("distinct %d draws %d values, and the path renders %d of weight above 0", n, n, passing)
	case n == 1:
		return "no value passes the where clause"
	case passing == 0:
		return fmt.Sprintf("distinct %d draws %d values, and none pass the where clause", n, n)
	case passing == 1:
		return fmt.Sprintf("distinct %d draws %d values, and only 1 passes the where clause", n, n)
	}
	return fmt.Sprintf("distinct %d draws %d values, and only %d pass the where clause", n, n, passing)
}

// passing is how many values are lo to hi long. A name's bounds may lie beyond any int, so lo and
// hi are clamped between a length shorter than any value and one longer than all before they turn
// into ints.
func (v lengthProof) passing(lo, hi float64) int {
	const shorterThanAny = -1.0
	longerThanAll := 1.0
	if len(v.lengths) > 0 {
		longerThanAll += float64(v.lengths[len(v.lengths)-1])
	}
	clamp := func(n float64) int { return int(min(max(n, shorterThanAny), longerThanAll)) }
	from, to := tree.SpanOf(v.lengths, clamp(lo), clamp(hi))
	return max(to-from, 0)
}
