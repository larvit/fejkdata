package load

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"unicode/utf8"

	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// maxDrawValues caps the values a walk of one path lists at load.
const maxDrawValues = 1 << 20

type weightedText struct {
	text   string
	weight float64
}

// valueWalk lists every value a path renders with its odds, refusing a path whose values are not
// fixed at load.
type valueWalk struct {
	finds *tree.LoadFinds
	seen  map[tree.Node]bool // the nodes the walk is inside
	out   []weightedText
}

// walk lists the values of the path tail into n, each with weight times its odds, written
// between pre and suf.
func (w *valueWalk) walk(n tree.Node, tail []string, weight float64, pre, suf string) error {
	if w.seen[n] {
		return errors.New("the path reads itself")
	}
	w.seen[n] = true
	defer delete(w.seen, n)
	switch n := n.(type) {
	case *tree.Choice:
		return w.walkChoice(n, tail, weight, pre, suf)
	case *tree.Template:
		if len(tail) > 0 {
			return w.walk(n.Fields()[tail[0]], tail[1:], weight, pre, suf)
		}
		return w.walkTemplate(n, weight, pre, suf)
	case *tree.Table:
		return w.walkTable(n, tail, weight, pre, suf)
	case *tree.TableColumn:
		return w.walkRows(n.Table(), []rowPart{{column: n}}, weight, pre, suf)
	case *tree.NullItem:
		return errors.New("the path may draw a null")
	}
	panic(invariant.Broken("a value walk reaches %T", n))
}

func (w *valueWalk) walkChoice(c *tree.Choice, tail []string, weight float64, pre, suf string) error {
	total, prev := float64(len(c.Items())), 0.0
	if c.Cum() != nil {
		total = c.Cum()[len(c.Cum())-1]
	}
	for i, item := range c.Items() {
		odds := 1.0
		if c.Cum() != nil {
			odds, prev = c.Cum()[i]-prev, c.Cum()[i]
		}
		if err := w.walk(item, tail, weight*odds/total, pre, suf); err != nil {
			return err
		}
	}
	return nil
}

func (w *valueWalk) walkTemplate(t *tree.Template, weight float64, pre, suf string) error {
	if t.Repeat() != 1 {
		return fmt.Errorf("%q repeats", t.Format())
	}
	if lit, fixed := t.FixedText(); fixed {
		return w.emit(pre+lit+suf, weight)
	}
	before, read, after, err := oneRead(t, w.finds)
	if err != nil {
		return err
	}
	if read == nil {
		return w.emit(pre+before+after+suf, weight)
	}
	for _, a := range read.Reads() {
		if err := w.walkRead(t, a, weight/float64(len(read.Reads())), pre+before, after+suf); err != nil {
			return err
		}
	}
	return nil
}

// oneRead splits t's format around its one read, refusing a format that draws anything else.
func oneRead(t *tree.Template, l *tree.LoadFinds) (before string, read *tree.Op, after string, err error) {
	ops := l.Of(t).Compiled().Ops()
	for i := range ops {
		o := &ops[i]
		switch tok := o.Token(); {
		case tok.Kind == grammar.LiteralRun && read == nil:
			before += tok.Lit
		case tok.Kind == grammar.LiteralRun:
			after += tok.Lit
		case tok.Kind == grammar.BuiltinCall:
			return "", nil, "", fmt.Errorf("%q renders {%s}, which draws anew each time", t.Format(), tok.Body)
		case tok.Where != nil || tok.Distinct != nil:
			return "", nil, "", fmt.Errorf("%q renders {%s}, which draws through a clause of its own", t.Format(), tok.Body)
		case read != nil:
			return "", nil, "", fmt.Errorf("%q reads both {%s} and {%s}", t.Format(), read.Token().Body, tok.Body)
		default:
			read = o
		}
	}
	return before, read, after, nil
}

// walkRead walks one alternative of t's read: a path from a field or a reference, or through a name t
// binds itself; that name's one read draws as a fresh read would.
func (w *valueWalk) walkRead(t *tree.Template, a tree.Read, weight float64, pre, suf string) error {
	if a.Kind() != tree.NamedRead {
		return w.walk(t.StartOf(a.Head(), w.finds.Of(t).Refs()), a.Tail(), weight, pre, suf)
	}
	if a.Named().Binder() != t {
		return fmt.Errorf("%q reads name %q, bound outside it, which keeps one draw for a whole render", t.Format(), a.Head())
	}
	target := w.finds.Binding(a.Named()).Target()
	return w.walk(target.Start(), slices.Concat(target.Tail(), a.Tail()), weight, pre, suf)
}

func (w *valueWalk) walkTable(t *tree.Table, tail []string, weight float64, pre, suf string) error {
	if len(tail) == 0 {
		if parts, own := ownColumnReads(t, w.finds); own {
			return w.walkRows(t, parts, weight, pre, suf)
		}
		return w.walk(t.FormatTemplate(), nil, weight, pre, suf)
	}
	if c, isColumn := t.FormatTemplate().Fields()[tail[0]].(*tree.TableColumn); isColumn && len(tail) == 1 {
		return w.walk(c, nil, weight, pre, suf)
	}
	return fmt.Errorf("the path runs through a selector, a linked table or \"..\": %s of %s", grammar.JoinSegments(tail), t.Rows().Segment())
}

// rowPart is one piece of a row's text: fixed text, or the row's cell in column.
type rowPart struct {
	lit    string
	column *tree.TableColumn
}

// ownColumnReads is t's format as fixed text and plain reads of t's own columns; own is false where
// the format reads anything else.
func ownColumnReads(t *tree.Table, l *tree.LoadFinds) (parts []rowPart, own bool) {
	for _, o := range l.Of(t.FormatTemplate()).Compiled().Ops() {
		tok := o.Token()
		if tok.Kind == grammar.LiteralRun {
			parts = append(parts, rowPart{lit: tok.Lit})
			continue
		}
		if tok.Kind != grammar.PathRead || tok.Where != nil || tok.Distinct != nil || len(o.Reads()) != 1 {
			return nil, false
		}
		a := o.Reads()[0]
		c, isColumn := t.FormatTemplate().Fields()[a.Head()].(*tree.TableColumn)
		if !isColumn || a.Kind() != tree.FreshRead || len(a.Tail()) > 0 {
			return nil, false
		}
		parts = append(parts, rowPart{column: c})
	}
	return parts, true
}

// walkRows lists each row of t as parts spell it, weighed by the row, as a draw over the whole table
// does.
func (w *valueWalk) walkRows(t *tree.Table, parts []rowPart, weight float64, pre, suf string) error {
	total := 0.0
	for r := 0; r < t.Rows().Len(); r++ {
		total += t.Rows().Weight(r)
	}
	for r := 0; r < t.Rows().Len(); r++ {
		text := pre
		for _, p := range parts {
			if p.column == nil {
				text += p.lit
				continue
			}
			if cell := t.CellTemplate(r, p.column.Index()); cell != nil {
				return fmt.Errorf("column %s of %s holds %q, which is not fixed text", p.column.Name(), t.Rows().Segment(), cell.Format())
			}
			text += t.Rows().Cell(r, p.column.Index())
		}
		if err := w.emit(text+suf, weight*t.Rows().Weight(r)/total); err != nil {
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

func newLengthIndex(values []weightedText) tree.LengthIndex {
	type sized struct {
		weightedText
		length int
	}
	vs := make([]sized, len(values))
	for i, v := range values {
		vs[i] = sized{v, utf8.RuneCountInString(v.text)}
	}
	sort.SliceStable(vs, func(i, j int) bool { return vs[i].length < vs[j].length })
	texts, lengths, cum := make([]string, len(vs)), make([]int, len(vs)), make([]float64, len(vs))
	total := 0.0
	for i, v := range vs {
		total += v.weight
		texts[i], lengths[i], cum[i] = v.text, v.length, total
	}
	return tree.NewLengthIndex(texts, lengths, cum)
}
