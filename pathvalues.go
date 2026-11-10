package fejkdata

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"unicode/utf8"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// maxDrawValues caps the values a walk of one path lists at load.
const maxDrawValues = 1 << 20

// lengthIndex holds values by length, each with its odds.
type lengthIndex struct {
	texts   []string  // shortest first
	lengths []int     // each text's length in code points
	cum     []float64 // running sum of the texts' odds
}

type weightedText struct {
	text   string
	weight float64
}

// valueWalk lists every value a path renders with its odds, refusing a path whose values are not
// fixed at load.
type valueWalk struct {
	links *links
	seen  map[node]bool // the nodes the walk is inside
	out   []weightedText
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
	panic(invariant.Broken("a value walk reaches %T", n))
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
	before, read, after, err := oneRead(t, w.links)
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
func oneRead(t *template, l *links) (before string, read *op, after string, err error) {
	ops := l.of(t).compiled.ops
	for i := range ops {
		o := &ops[i]
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
// binds itself; that name's one read draws as a fresh read would.
func (w *valueWalk) walkArm(t *template, a arm, weight float64, pre, suf string) error {
	if a.kind != namedRead {
		return w.walk(t.startOf(a.head, w.links), a.tail, weight, pre, suf)
	}
	if a.named.binder != t {
		return fmt.Errorf("%q reads name %q, bound outside it, which keeps one draw for a whole render", t.format, a.head)
	}
	target := w.links.binding(a.named).target
	return w.walk(target.start, slices.Concat(target.tail, a.tail), weight, pre, suf)
}

func (w *valueWalk) walkTable(t *table, tail []string, weight float64, pre, suf string) error {
	if len(tail) == 0 {
		if parts, own := ownColumnReads(t, w.links); own {
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
func ownColumnReads(t *table, l *links) (parts []rowPart, own bool) {
	for _, o := range l.of(t.formatTemplate).compiled.ops {
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

// span bounds the texts in x.texts[from:to], those whose length is lo to hi.
func (x *lengthIndex) span(lo, hi int) (from, to int) { return spanOf(x.lengths, lo, hi) }

func spanOf(lengths []int, lo, hi int) (from, to int) {
	from = sort.Search(len(lengths), func(i int) bool { return lengths[i] >= lo })
	to = sort.Search(len(lengths), func(i int) bool { return lengths[i] > hi })
	return from, to
}

// drawIn draws one of the texts in x.texts[from:to], by their odds.
func (x *lengthIndex) drawIn(s *drawstate.State, from, to int) string {
	return x.texts[x.drawIndex(s, from, to, nil)]
}

// drawIndex draws the index of one of the texts in x.texts[from:to] by their odds, leaving out
// the ascending indexes drawn. It costs drawn's length plus a binary search over the span: it draws
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
// after it, else the last one before it, within from:to.
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

// nearest is the span of the texts whose length is n or, where none is, nearest n, the shorter on
// a tie.
func (x *lengthIndex) nearest(n int) (from, to int) {
	from, to = x.span(n, n)
	if from < to {
		return from, to
	}
	l := x.lengths[len(x.lengths)-1]
	switch {
	case from == len(x.lengths):
	case from == 0 || x.lengths[from]-n < n-x.lengths[from-1]:
		l = x.lengths[from]
	default:
		l = x.lengths[from-1]
	}
	return x.span(l, l)
}
