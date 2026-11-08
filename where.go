package fejkdata

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// maxFilterValues caps the values one where clause weighs at load.
const maxFilterValues = 1 << 20

// whereFilter draws one of the values a path renders that pass a where clause, with the odds it
// has in a draw of the path. The values are fixed at load; a length that reads a name reads it at
// each draw.
type whereFilter struct {
	lengthIndex
	lo, hi lengthEnd
	reads  []arm // the op's operands, which an end reading a name indexes
}

// lengthIndex holds values by length, each with its odds.
type lengthIndex struct {
	texts   []string  // shortest first
	lengths []int     // each text's length in code points
	cum     []float64 // running sum of the texts' odds
}

// lengthEnd is one end of a where clause's length: n, or the name reads[read] where read is 0 or more.
type lengthEnd struct {
	n    int
	read int
}

// lengthNames is every name w's length reads, each once.
func lengthNames(w *grammar.Where) []string {
	if w == nil || w.Length == nil {
		return nil
	}
	var names []string
	for _, end := range []grammar.Length{w.Length.Lo, w.Length.Hi} {
		if end.Name != "" && !slices.Contains(names, end.Name) {
			names = append(names, end.Name)
		}
	}
	return names
}

// compileLengths compiles the names w's length reads, each a name bound around t.
func (t *template) compileLengths(w *grammar.Where, targets map[*nameBinding]nameTarget) ([]arm, error) {
	var arms []arm
	for _, name := range lengthNames(w) {
		if !t.isName(name) {
			return nil, fmt.Errorf("length %s reads a name, and %q is a field; bind it first, as {%s as n}, and write length n", name, name, name)
		}
		a, err := t.compileArm(name, targets)
		if err != nil {
			return nil, err
		}
		arms = append(arms, a)
	}
	return arms, nil
}

// filterDraws builds the filter of each token in ts with a where clause, and returns those of the
// names bound to a filtered draw.
func filterDraws(ts []templateSite, targets map[*nameBinding]nameTarget) (map[*nameBinding]*whereFilter, error) {
	binds := map[*nameBinding]*whereFilter{}
	for _, s := range ts {
		for i := range s.t.compiled.ops {
			o := &s.t.compiled.ops[i]
			if o.Where == nil {
				continue
			}
			f, err := s.t.filterOf(o, targets)
			if err != nil {
				return nil, fmt.Errorf("%s: token {%s}: %w", s.label, o.Body, err)
			}
			o.filter = f
			if o.Kind == grammar.NameBind {
				binds[s.t.nameScope.bindings[o.Bound]] = f
			}
		}
	}
	for _, s := range ts {
		for _, r := range namedReads(s.t) {
			if binds[r.a.named] != nil && len(r.a.tail) > 0 {
				return nil, fmt.Errorf("%s: token {%s}: name %q holds a filtered draw, which is read whole: {%s}", s.label, r.o.Body, r.a.head, r.a.head)
			}
		}
	}
	return binds, nil
}

// filterOf builds o's filter over the values the path it filters renders, and proves every draw
// leaves one to draw.
func (t *template) filterOf(o *op, targets map[*nameBinding]nameTarget) (*whereFilter, error) {
	var start node
	var tail []string
	if o.Kind == grammar.PathRead {
		a := o.arms[0]
		if a.kind == namedRead {
			return nil, fmt.Errorf("%q is a name, drawn once; filter the draw it binds: {%s where … as %s}", a.head, a.named.ref, a.head)
		}
		start, tail = t.startOf(a.head), a.tail
	} else {
		target := targets[t.nameScope.bindings[o.Bound]]
		start, tail = target.start, target.tail
	}
	w := valueWalk{targets: targets, seen: map[node]bool{}}
	if err := w.walk(start, tail, 1, "", ""); err != nil {
		return nil, fmt.Errorf("a where clause filters values fixed at load, and %w", err)
	}
	f := newWhereFilter(w.out, o.Where, o.operands)
	return f, f.proveEveryDrawPasses()
}

type filterValue struct {
	text   string
	weight float64
}

// valueWalk lists every value a path renders with its odds, refusing a path whose values are not
// fixed at load.
type valueWalk struct {
	targets map[*nameBinding]nameTarget
	seen    map[node]bool // the nodes the walk is inside
	out     []filterValue
}

// walk lists the values of the path tail into n, each with weight times its odds, written
// between pre and suf.
func (w *valueWalk) walk(n node, tail []string, weight float64, pre, suf string) error {
	if w.seen[n] {
		return errors.New("the path reads itself")
	}
	w.seen[n] = true
	defer delete(w.seen, n)
	switch n := n.(type) {
	case *choice:
		return w.walkChoice(n, tail, weight, pre, suf)
	case *template:
		if len(tail) > 0 {
			return w.walk(n.fields[tail[0]], tail[1:], weight, pre, suf)
		}
		return w.walkTemplate(n, weight, pre, suf)
	case *table:
		return w.walkTable(n, tail, weight, pre, suf)
	case *tableColumn:
		return w.walkCells(n, weight, pre, suf)
	case *nullItem:
		return errors.New("the path may draw a null")
	}
	panic(invariant.Broken("a where clause walks into %T", n))
}

func (w *valueWalk) walkChoice(c *choice, tail []string, weight float64, pre, suf string) error {
	total, prev := float64(len(c.items)), 0.0
	if c.cum != nil {
		total = c.cum[len(c.cum)-1]
	}
	for i, item := range c.items {
		odds := 1.0
		if c.cum != nil {
			odds, prev = c.cum[i]-prev, c.cum[i]
		}
		if err := w.walk(item, tail, weight*odds/total, pre, suf); err != nil {
			return err
		}
	}
	return nil
}

func (w *valueWalk) walkTemplate(t *template, weight float64, pre, suf string) error {
	if t.repeat != 1 {
		return fmt.Errorf("%q repeats", t.format)
	}
	if lit, fixed := t.fixedText(); fixed {
		return w.emit(pre+lit+suf, weight)
	}
	before, read, after, err := oneRead(t)
	if err != nil {
		return err
	}
	if read == nil {
		return w.emit(pre+before+after+suf, weight)
	}
	for _, a := range read.arms {
		if err := w.walkArm(t, a, weight/float64(len(read.arms)), pre+before, after+suf); err != nil {
			return err
		}
	}
	return nil
}

// oneRead splits t's format around its one read, refusing a format that draws anything else.
func oneRead(t *template) (before string, read *op, after string, err error) {
	for i := range t.compiled.ops {
		o := &t.compiled.ops[i]
		switch {
		case o.Kind == grammar.LiteralRun && read == nil:
			before += o.Lit
		case o.Kind == grammar.LiteralRun:
			after += o.Lit
		case o.Kind == grammar.BuiltinCall:
			return "", nil, "", fmt.Errorf("%q renders {%s}, which draws anew each time", t.format, o.Body)
		case o.Where != nil:
			return "", nil, "", fmt.Errorf("%q renders {%s}, which filters a draw of its own", t.format, o.Body)
		case read != nil:
			return "", nil, "", fmt.Errorf("%q reads both {%s} and {%s}", t.format, read.Body, o.Body)
		default:
			read = o
		}
	}
	return before, read, after, nil
}

// walkArm walks one arm of t's read: a path from a field or a reference, or through a name t
// binds itself, which its one read draws as a fresh read would.
func (w *valueWalk) walkArm(t *template, a arm, weight float64, pre, suf string) error {
	if a.kind != namedRead {
		return w.walk(t.startOf(a.head), a.tail, weight, pre, suf)
	}
	if a.named.binder != t {
		return fmt.Errorf("%q reads name %q, bound outside it, which keeps one draw for a whole render", t.format, a.head)
	}
	target, resolved := w.targets[a.named]
	if !resolved {
		target = a.named.target
	}
	return w.walk(target.start, slices.Concat(target.tail, a.tail), weight, pre, suf)
}

func (w *valueWalk) walkTable(t *table, tail []string, weight float64, pre, suf string) error {
	if len(tail) == 0 {
		return w.walk(t.formatTemplate, nil, weight, pre, suf)
	}
	if c, isColumn := t.formatTemplate.fields[tail[0]].(*tableColumn); isColumn && len(tail) == 1 {
		return w.walk(c, nil, weight, pre, suf)
	}
	return fmt.Errorf("follows no selector, linked table or step up: %s of %s", grammar.JoinSegments(tail), t.rows.Segment())
}

// walkCells lists a column's cells, each weighed by its row, as a draw over the whole table does.
func (w *valueWalk) walkCells(c *tableColumn, weight float64, pre, suf string) error {
	total := 0.0
	for r := 0; r < c.t.rows.Len(); r++ {
		total += c.t.rows.Weight(r)
	}
	for r := 0; r < c.t.rows.Len(); r++ {
		if cell := c.t.cellTemplate(r, c.i); cell != nil {
			return fmt.Errorf("column %s of %s holds %q, which a where clause can't weigh", c.name(), c.t.rows.Segment(), cell.format)
		}
		if err := w.emit(pre+c.t.rows.Cell(r, c.i)+suf, weight*c.t.rows.Weight(r)/total); err != nil {
			return err
		}
	}
	return nil
}

func (w *valueWalk) emit(text string, weight float64) error {
	if weight == 0 {
		return nil
	}
	if len(w.out) == maxFilterValues {
		return fmt.Errorf("the path renders more than %d values", maxFilterValues)
	}
	w.out = append(w.out, filterValue{text, weight})
	return nil
}

// newWhereFilter keeps the values that pass where's fixed conditions, shortest first.
func newWhereFilter(values []filterValue, where *grammar.Where, reads []arm) *whereFilter {
	var kept []filterValue
	for _, v := range values {
		if !slices.Contains(where.Not, v.text) && !slices.ContainsFunc(where.NotContaining, func(s string) bool { return strings.Contains(v.text, s) }) {
			kept = append(kept, v)
		}
	}
	f := &whereFilter{lengthIndex: newLengthIndex(kept), reads: reads}
	f.lo, f.hi = lengthEnd{n: 0, read: -1}, lengthEnd{n: math.MaxInt, read: -1}
	if where.Length != nil {
		names := lengthNames(where)
		f.lo, f.hi = endOf(where.Length.Lo, names), endOf(where.Length.Hi, names)
	}
	return f
}

func endOf(l grammar.Length, names []string) lengthEnd {
	if l.Name == "" {
		return lengthEnd{n: l.N, read: -1}
	}
	return lengthEnd{read: slices.Index(names, l.Name)}
}

func newLengthIndex(values []filterValue) lengthIndex {
	type sized struct {
		filterValue
		length int
	}
	vs := make([]sized, len(values))
	for i, v := range values {
		vs[i] = sized{v, utf8.RuneCountInString(v.text)}
	}
	sort.SliceStable(vs, func(i, j int) bool { return vs[i].length < vs[j].length })
	x := lengthIndex{texts: make([]string, len(vs)), lengths: make([]int, len(vs)), cum: make([]float64, len(vs))}
	total := 0.0
	for i, v := range vs {
		total += v.weight
		x.texts[i], x.lengths[i], x.cum[i] = v.text, v.length, total
	}
	return x
}

// span is the texts from index from up to to whose length is lo to hi.
func (x *lengthIndex) span(lo, hi int) (from, to int) {
	from = sort.Search(len(x.lengths), func(i int) bool { return x.lengths[i] >= lo })
	to = sort.Search(len(x.lengths), func(i int) bool { return x.lengths[i] > hi })
	return from, to
}

// drawIn draws one of the texts from index from up to to, by their odds.
func (x *lengthIndex) drawIn(s *drawstate.State, from, to int) string {
	base := 0.0
	if from > 0 {
		base = x.cum[from-1]
	}
	v := base + s.Float64()*(x.cum[to-1]-base)
	i := from + sort.Search(to-from, func(i int) bool { return x.cum[from+i] > v })
	return x.texts[min(i, to-1)]
}

// proveEveryDrawPasses refuses a filter some draw leaves with no value: one no value passes, or one
// whose length reads a name that may print a length no value has.
func (f *whereFilter) proveEveryDrawPasses() error {
	bounds := make([][2]float64, len(f.reads))
	for i, a := range f.reads {
		v := (&valueProof{}).proveUnion(a.leaves)
		switch {
		case v.NotOperand != "":
			return fmt.Errorf("length %s is not always a whole number: %s", a.spelling, v.NotOperand)
		case !v.Integral:
			return fmt.Errorf("length %s is not always a whole number: it may print a fraction", a.spelling)
		}
		bounds[i] = [2]float64{v.Lo, v.Hi}
	}
	if f.lo.read >= 0 && f.lo.read == f.hi.read {
		for n := bounds[f.lo.read][0]; n <= bounds[f.lo.read][1]; n++ {
			if !f.passes(n, n) {
				return fmt.Errorf("where %s is %.0f, no value passes the where clause", f.reads[f.lo.read].spelling, n)
			}
		}
		return nil
	}
	lo, hi := float64(f.lo.n), float64(f.hi.n)
	if f.lo.read >= 0 {
		lo = bounds[f.lo.read][1]
	}
	if f.hi.read >= 0 {
		hi = bounds[f.hi.read][0]
	}
	if !f.passes(lo, hi) {
		return errors.New("no value passes the where clause")
	}
	return nil
}

// passes reports whether some value is lo to hi long.
func (f *whereFilter) passes(lo, hi float64) bool {
	past := 1.0
	if len(f.lengths) > 0 {
		past += float64(f.lengths[len(f.lengths)-1])
	}
	from, to := f.span(int(min(max(lo, -1), past)), int(max(min(hi, past), -1)))
	return from < to
}

// draw draws one passing value, reading the length's names in env.
func (f *whereFilter) draw(s *drawstate.State, env renderEnv) string {
	from, to := f.span(f.lo.at(s, env, f.reads), f.hi.at(s, env, f.reads))
	if from == to {
		panic(invariant.Broken("a where clause the load proved leaves no value"))
	}
	return f.drawIn(s, from, to)
}

func (e lengthEnd) at(s *drawstate.State, env renderEnv, reads []arm) int {
	if e.read < 0 {
		return e.n
	}
	v, _ := strconv.ParseFloat(strings.TrimSpace(readName(s, env, reads[e.read]).text), 64)
	return int(v)
}
