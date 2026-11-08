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

// maxDrawValues caps the values one where or distinct clause weighs at load.
const maxDrawValues = 1 << 20

// valueDraw draws among the values a path renders, fixed at load with the odds each has in a draw of
// the path: one of those that pass a where clause, or count distinct ones where a distinct clause
// asks. A length that reads a name reads it at each draw.
type valueDraw struct {
	lengthIndex
	lo, hi  lengthEnd
	reads   []arm  // the op's operands, which an end reading a name indexes
	filters bool   // a where clause filters the values
	count   int    // how many distinct values one draw takes
	sep     string // what prints between them
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
		if !t.isName(splitArm(name, nil).head) {
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

// valueDraws builds the draw of each token in ts with a where or a distinct clause, and gives a name
// bound to one its draw.
func valueDraws(ts []templateSite, targets map[*nameBinding]nameTarget) error {
	for _, s := range ts {
		for i := range s.t.compiled.ops {
			o := &s.t.compiled.ops[i]
			if o.Where == nil && o.Distinct == nil {
				continue
			}
			v, err := s.t.valueDrawOf(o, targets)
			if err != nil {
				return fmt.Errorf("%s: token {%s}: %w", s.label, o.Body, err)
			}
			o.values = v
			if o.Kind == grammar.NameBind {
				s.t.nameScope.bindings[o.Bound].values = v
			}
		}
	}
	for _, s := range ts {
		for _, r := range namedReads(s.t) {
			if r.a.named.values != nil && len(r.a.tail) > 0 {
				return fmt.Errorf("%s: token {%s}: name %q holds a filtered draw, which is read whole: {%s}", s.label, r.o.Body, r.a.head, r.a.head)
			}
		}
	}
	return nil
}

// valueDrawOf builds o's draw over the values the path it reads renders, and proves every draw
// leaves enough to draw.
func (t *template) valueDrawOf(o *op, targets map[*nameBinding]nameTarget) (*valueDraw, error) {
	clause := "a where clause filters"
	if o.Where == nil {
		clause = "a distinct clause draws"
	}
	var start node
	var tail []string
	if o.Kind == grammar.PathRead {
		a := o.arms[0]
		if a.kind == namedRead {
			return nil, fmt.Errorf("%q is a name, drawn once; write the clause on the draw it binds: {%s … as %s}", a.head, a.named.ref, a.head)
		}
		start, tail = t.startOf(a.head), a.tail
	} else {
		target := targets[t.nameScope.bindings[o.Bound]]
		start, tail = target.start, target.tail
	}
	w := valueWalk{targets: targets, seen: map[node]bool{}}
	if err := w.walk(start, tail, 1, "", ""); err != nil {
		return nil, fmt.Errorf("%s values fixed at load, and %w", clause, err)
	}
	v := newValueDraw(w.out, o.Where, o.Distinct, o.operands)
	return v, v.proveEveryDrawPasses()
}

type weightedText struct {
	text   string
	weight float64
}

// valueWalk lists every value a path renders with its odds, refusing a path whose values are not
// fixed at load.
type valueWalk struct {
	targets map[*nameBinding]nameTarget
	seen    map[node]bool // the nodes the walk is inside
	out     []weightedText
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
		return w.walkRows(n.t, []rowPart{{column: n}}, weight, pre, suf)
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
		case o.Where != nil || o.Distinct != nil:
			return "", nil, "", fmt.Errorf("%q renders {%s}, which draws through a clause of its own", t.format, o.Body)
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
		if parts, own := ownColumnReads(t); own {
			return w.walkRows(t, parts, weight, pre, suf)
		}
		return w.walk(t.formatTemplate, nil, weight, pre, suf)
	}
	if c, isColumn := t.formatTemplate.fields[tail[0]].(*tableColumn); isColumn && len(tail) == 1 {
		return w.walk(c, nil, weight, pre, suf)
	}
	return fmt.Errorf("the path runs through a selector, a linked table or \"..\": %s of %s", grammar.JoinSegments(tail), t.rows.Segment())
}

// rowPart is one piece of a row's text: fixed text, or the row's cell in column.
type rowPart struct {
	lit    string
	column *tableColumn
}

// ownColumnReads is t's format as fixed text and plain reads of t's own columns; own is false where
// the format reads anything else.
func ownColumnReads(t *table) (parts []rowPart, own bool) {
	for _, o := range t.formatTemplate.compiled.ops {
		if o.Kind == grammar.LiteralRun {
			parts = append(parts, rowPart{lit: o.Lit})
			continue
		}
		if o.Kind != grammar.PathRead || o.Where != nil || o.Distinct != nil || len(o.arms) != 1 {
			return nil, false
		}
		a := o.arms[0]
		c, isColumn := t.formatTemplate.fields[a.head].(*tableColumn)
		if !isColumn || a.kind != freshRead || len(a.tail) > 0 {
			return nil, false
		}
		parts = append(parts, rowPart{column: c})
	}
	return parts, true
}

// walkRows lists each row of t as parts spell it, weighed by the row, as a draw over the whole table
// does.
func (w *valueWalk) walkRows(t *table, parts []rowPart, weight float64, pre, suf string) error {
	total := 0.0
	for r := 0; r < t.rows.Len(); r++ {
		total += t.rows.Weight(r)
	}
	for r := 0; r < t.rows.Len(); r++ {
		text := pre
		for _, p := range parts {
			if p.column == nil {
				text += p.lit
				continue
			}
			if cell := t.cellTemplate(r, p.column.i); cell != nil {
				return fmt.Errorf("column %s of %s holds %q, which is not fixed text", p.column.name(), t.rows.Segment(), cell.format)
			}
			text += t.rows.Cell(r, p.column.i)
		}
		if err := w.emit(text+suf, weight*t.rows.Weight(r)/total); err != nil {
			return err
		}
	}
	return nil
}

func (w *valueWalk) emit(text string, weight float64) error {
	if weight == 0 {
		return nil
	}
	if len(w.out) == maxDrawValues {
		return fmt.Errorf("the path renders more than %d values", maxDrawValues)
	}
	w.out = append(w.out, weightedText{text, weight})
	return nil
}

// newValueDraw keeps the values that pass where's fixed conditions, shortest first, for draws of
// distinct's count, or of one where distinct is nil.
func newValueDraw(values []weightedText, where *grammar.Where, distinct *grammar.Distinct, reads []arm) *valueDraw {
	v := &valueDraw{reads: reads, count: 1, lo: lengthEnd{n: 0, read: -1}, hi: lengthEnd{n: math.MaxInt, read: -1}}
	if distinct != nil {
		v.count, v.sep = distinct.N, distinct.Sep
	}
	if where == nil {
		v.lengthIndex = newLengthIndex(values)
		return v
	}
	var kept []weightedText
	for _, value := range values {
		if !slices.Contains(where.Not, value.text) && !slices.ContainsFunc(where.NotContaining, func(s string) bool { return strings.Contains(value.text, s) }) {
			kept = append(kept, value)
		}
	}
	v.lengthIndex, v.filters = newLengthIndex(kept), true
	if where.Length != nil {
		names := lengthNames(where)
		v.lo, v.hi = endOf(where.Length.Lo, names), endOf(where.Length.Hi, names)
	}
	return v
}

func endOf(l grammar.Length, names []string) lengthEnd {
	if l.Name == "" {
		return lengthEnd{n: l.N, read: -1}
	}
	return lengthEnd{read: slices.Index(names, l.Name)}
}

func newLengthIndex(values []weightedText) lengthIndex {
	type sized struct {
		weightedText
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
	return x.texts[x.drawIndex(s, from, to, nil)]
}

// drawDistinctIn draws n of the texts from index from up to to, each by its odds among those not
// drawn yet.
func (x *lengthIndex) drawDistinctIn(s *drawstate.State, from, to, n int) []string {
	drawn := make([]int, 0, n) // ascending
	out := make([]string, n)
	for k := range out {
		i := x.drawIndex(s, from, to, drawn)
		at, _ := slices.BinarySearch(drawn, i)
		drawn = slices.Insert(drawn, at, i)
		out[k] = x.texts[i]
	}
	return out
}

// drawIndex draws the index of one of the texts from index from up to to by their odds, leaving out
// the ascending indexes drawn. It costs what drawn's length costs, whatever the span's: it draws
// among the odds left, and steps past each drawn text the draw lands at or beyond.
func (x *lengthIndex) drawIndex(s *drawstate.State, from, to int, drawn []int) int {
	base, left := x.before(from), x.cum[to-1]-x.before(from)
	for _, d := range drawn {
		left -= x.cum[d] - x.before(d)
	}
	v := base + s.Float64()*left
	for _, d := range drawn {
		if x.before(d) > v {
			break
		}
		v += x.cum[d] - x.before(d)
	}
	return undrawnNear(from+sort.Search(to-from, func(i int) bool { return x.cum[from+i] > v }), from, to, drawn)
}

// before is the running sum of the odds of the texts before index i.
func (x *lengthIndex) before(i int) float64 {
	if i == 0 {
		return 0
	}
	return x.cum[i-1]
}

// undrawnNear is i, or where rounding landed i on a drawn index or on to, the first undrawn index
// after it, else the last one before it, from up to to.
func undrawnNear(i, from, to int, drawn []int) int {
	for j := i; j < to; j++ {
		if _, taken := slices.BinarySearch(drawn, j); !taken {
			return j
		}
	}
	for j := i - 1; j >= from; j-- {
		if _, taken := slices.BinarySearch(drawn, j); !taken {
			return j
		}
	}
	panic(invariant.Broken("a draw ran out of the values the load proved enough"))
}

// proveEveryDrawPasses refuses a draw some render leaves with too few values: one too few values
// pass, or one whose length reads a name that may print a length too few values have.
func (v *valueDraw) proveEveryDrawPasses() error {
	bounds := make([][2]float64, len(v.reads))
	for i, a := range v.reads {
		p := (&valueProof{}).proveArm(a)
		switch {
		case p.NotOperand != "":
			return fmt.Errorf("length %s is not always a whole number: %s", a.spelling, p.NotOperand)
		case !p.Integral:
			return fmt.Errorf("length %s is not always a whole number: it may print a fraction", a.spelling)
		}
		bounds[i] = [2]float64{p.Lo, p.Hi}
	}
	if v.lo.read >= 0 && v.lo.read == v.hi.read {
		for n := bounds[v.lo.read][0]; n <= bounds[v.lo.read][1]; n++ {
			if passing := v.passing(n, n); passing < v.count {
				return fmt.Errorf("where %s is %.0f, %s", v.reads[v.lo.read].spelling, n, v.shortfall(passing))
			}
		}
		return nil
	}
	lo, hi := float64(v.lo.n), float64(v.hi.n)
	if v.lo.read >= 0 {
		lo = bounds[v.lo.read][1]
	}
	if v.hi.read >= 0 {
		hi = bounds[v.hi.read][0]
	}
	if passing := v.passing(lo, hi); passing < v.count {
		return v.rangeShortfall(lo, hi, passing)
	}
	return nil
}

// rangeShortfall says that the length lo to hi leaves passing values, too few, naming the length's
// names at the values that narrow it to lo to hi.
func (v *valueDraw) rangeShortfall(lo, hi float64, passing int) error {
	var at []string
	if v.lo.read >= 0 {
		at = append(at, fmt.Sprintf("%s is %.0f", v.reads[v.lo.read].spelling, lo))
	}
	if v.hi.read >= 0 {
		at = append(at, fmt.Sprintf("%s is %.0f", v.reads[v.hi.read].spelling, hi))
	}
	switch {
	case len(at) == 0:
		return errors.New(v.shortfall(passing))
	case lo <= hi:
		return fmt.Errorf("where %s, %s", strings.Join(at, " and "), v.shortfall(passing))
	case v.hi.read < 0:
		return fmt.Errorf("where %s, length %.0f to %.0f holds no length; keep %s at or below %d", at[0], lo, hi, v.reads[v.lo.read].spelling, v.hi.n)
	case v.lo.read < 0:
		return fmt.Errorf("where %s, length %.0f to %.0f holds no length; keep %s at %d or above", at[0], lo, hi, v.reads[v.hi.read].spelling, v.lo.n)
	}
	return fmt.Errorf("where %s, length %.0f to %.0f holds no length; keep %s at or below %s", strings.Join(at, " and "), lo, hi, v.reads[v.lo.read].spelling, v.reads[v.hi.read].spelling)
}

// shortfall says that passing values are fewer than a draw takes.
func (v *valueDraw) shortfall(passing int) string {
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

// passing is how many values are lo to hi long.
func (v *valueDraw) passing(lo, hi float64) int {
	past := 1.0
	if len(v.lengths) > 0 {
		past += float64(v.lengths[len(v.lengths)-1])
	}
	from, to := v.span(int(min(max(lo, -1), past)), int(max(min(hi, past), -1)))
	return max(to-from, 0)
}

// draw draws count distinct passing values, reading the length's names in env.
func (v *valueDraw) draw(s *drawstate.State, env renderEnv) []string {
	from, to := v.span(v.lo.at(s, env, v.reads), v.hi.at(s, env, v.reads))
	if to-from < v.count {
		panic(invariant.Broken("a clause the load proved leaves %d values leaves %d", v.count, to-from))
	}
	return v.drawDistinctIn(s, from, to, v.count)
}

func (e lengthEnd) at(s *drawstate.State, env renderEnv, reads []arm) int {
	if e.read < 0 {
		return e.n
	}
	v, _ := strconv.ParseFloat(strings.TrimSpace(readName(s, env, reads[e.read]).text), 64)
	return int(v)
}
