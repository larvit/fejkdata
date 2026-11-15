package load

import (
	"fmt"
	"strconv"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// readSpelling is one spelling split into the head, whose node `tree.Template.StartOf` finds, and the
// tail of a dotted path into it. refs maps a reference to what resolveRefs resolved it to; before
// that, a reference is whole.
type readSpelling struct {
	text string
	head string
	tail []string
}

func splitRead(spelling string, refs map[string]tree.ResolvedRef) readSpelling {
	if grammar.IsRef(spelling) {
		if r, resolved := refs[spelling]; resolved {
			return readSpelling{spelling, r.Head(), r.Tail()}
		}
		return readSpelling{spelling, spelling, nil}
	}
	segs, err := grammar.SplitPath(spelling)
	if err != nil || len(segs) == 1 {
		return readSpelling{spelling, spelling, nil}
	}
	return readSpelling{spelling, segs[0], segs[1:]}
}

func checkSegments(r readSpelling) error {
	if len(r.tail) == 0 {
		return nil
	}
	return grammar.CheckSegments(append([]string{r.head}, r.tail...))
}

// newOp returns parts whole: with its call, and with its draw where it has one. parts is an op whose
// alternatives, operands and clause compileFormat built. An adddays call takes the one layout its date
// operand prints in, so every template's format must be compiled first.
func newOp(t *tree.Template, parts tree.Op, l *tree.LoadFinds, words map[tree.WordsAt]*tree.WordList) (tree.Op, error) {
	o, err := withCall(parts, l)
	if err != nil {
		return tree.Op{}, err
	}
	draw, err := drawOf(t, &o, l, words)
	if err != nil {
		return tree.Op{}, err
	}
	return o.WithDraw(draw), nil
}

// withCall is o with its call, where its token calls a builtin.
func withCall(o tree.Op, l *tree.LoadFinds) (tree.Op, error) {
	tok := o.Token()
	if tok.Kind != grammar.BuiltinCall {
		return o, nil
	}
	call := builtinfunc.Prep(tok.Fn, tok.Args)
	if tok.Fn == builtinfunc.AddDaysName {
		layout, err := addDaysLayout(o.Operands()[0], l)
		if err != nil {
			return tree.Op{}, err
		}
		call = builtinfunc.AddDays(layout, tok.Args)
	}
	return o.WithCall(call), nil
}

// drawOf is o's draw: none for a plain read or call, else the draw its reword call, or its where or
// distinct clause, makes.
func drawOf(t *tree.Template, o *tree.Op, l *tree.LoadFinds, words map[tree.WordsAt]*tree.WordList) (draw tree.TokenDraw, err error) {
	tok := o.Token()
	switch {
	case tok.Kind == grammar.BuiltinCall && tok.Fn == builtinfunc.RewordName:
		return rewordDrawOf(t, o, l, words)
	case tok.Where == nil && tok.Distinct == nil:
		return draw, nil
	case tok.Kind == grammar.BuiltinCall:
		return newAlphabetDraw(o)
	}
	return valueDrawOf(t, o, l)
}

func compileReads(t *tree.Template, spellings []string, l *tree.LoadFinds) ([]tree.Read, error) {
	if len(spellings) == 0 {
		return nil, nil
	}
	reads := make([]tree.Read, len(spellings))
	for i, spelling := range spellings {
		a, err := compileRead(t, spelling, l)
		if err != nil {
			return nil, err
		}
		reads[i] = a
	}
	return reads, nil
}

// compileRead compiles one read into a path: from the head it names, or, for a read through a
// name, from what that name binds.
func compileRead(t *tree.Template, spelling string, l *tree.LoadFinds) (tree.Read, error) {
	tl := l.Of(t)
	r := splitRead(spelling, tl.Refs().BySpelling())
	if isName(t, r.head, l) {
		return compileNamedRead(t, r, l)
	}
	start := t.StartOf(r.head, tl.Refs())
	if start == nil {
		panic(invariant.Broken("{%s} reads a head nothing resolved", r.text))
	}
	w := CompilePath(start, r.tail)
	return tree.NewRead(r.text, r.head, r.tail, tree.FreshLevels(r.head, r.tail), w.steps, w.Leaves), nil
}

// compileNamedRead compiles r, a read through a name, into a path from what that name binds.
func compileNamedRead(t *tree.Template, r readSpelling, l *tree.LoadFinds) (tree.Read, error) {
	b := l.Of(t).NameScope().Lookup(r.head)
	tail, member := r.tail, 0
	if b.Distinct() != nil {
		var err error
		if member, err = memberRead(b, r); err != nil {
			return tree.Read{}, err
		}
		tail = nil
	} else if grammar.HasSelector(tail) {
		if b.BindsField() {
			return tree.Read{}, fmt.Errorf("a path through name %q may not select a row; bind the row to a name of its own", r.head)
		}
		return tree.Read{}, fmt.Errorf("a path through name %q may not select a row; read the row without the name, {%s}, or bind it to a name of its own", r.head, grammar.JoinSegments(append([]string{b.Ref()}, tail...)))
	}
	target := l.Binding(b).Target()
	full := append(target.Tail()[:len(target.Tail()):len(target.Tail())], tail...)
	if err := provePath(target.Start(), full, r.head, len(target.Tail())); err != nil {
		return tree.Read{}, err
	}
	w := CompilePath(target.Start(), full)
	return tree.NewRead(r.text, r.head, tail, tree.NamedLevels(full), w.steps, w.Leaves).ThroughName(b, member), nil
}

// memberRead is the member of b's distinct draw that a reads, counted from 1, or 0 where a reads them
// joined.
func memberRead(b *tree.NameBinding, a readSpelling) (int, error) {
	n := b.Distinct().N
	if len(a.tail) == 0 {
		return 0, readsAllMembers(b)
	}
	if len(a.tail) > 1 || !grammar.IsSelector(a.tail[0]) {
		return 0, fmt.Errorf("name %q holds distinct members, each read whole, as {%s[1]}", b.Name(), b.Name())
	}
	k, err := strconv.Atoi(grammar.SelectorOf(a.tail[0]))
	if err != nil || k < 1 || k > n {
		return 0, fmt.Errorf("name %q holds %d members, {%s[1]} to {%s[%d]}", b.Name(), n, b.Name(), b.Name(), n)
	}
	return k, nil
}

// readsAllMembers refuses a read of all of b's distinct members unless they are joined.
func readsAllMembers(b *tree.NameBinding) error {
	if b.Distinct().Joins {
		return nil
	}
	n := b.Distinct().N
	return fmt.Errorf("name %q holds %d distinct members: read one, as {%s[1]}, or add joined ',' after distinct %d to read them as one", b.Name(), n, b.Name(), n)
}
