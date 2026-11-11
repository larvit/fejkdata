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
// arms, operands and clause compileFormat built. An adddays call takes the one layout its date
// operand prints in, so every template's format must be compiled first.
func newOp(t *tree.Template, parts tree.Op, l *tree.LoadFinds, words map[tree.WordsAt]*tree.WordList) (tree.Op, error) {
	tok := parts.Token()
	o := parts
	if tok.Kind == grammar.BuiltinCall {
		call := builtinfunc.Prep(tok.Fn, tok.Args)
		if tok.Fn == builtinfunc.AddDaysName {
			layout, err := addDaysLayout(o.Operands()[0], l)
			if err != nil {
				return tree.Op{}, err
			}
			call = builtinfunc.AddDays(layout, tok.Args)
		}
		o = o.WithCall(call)
	}
	var draw tree.TokenDraw
	var err error
	switch {
	case tok.Kind == grammar.BuiltinCall && tok.Fn == builtinfunc.RewordName:
		draw, err = rewordDrawOf(t, &o, l, words)
	case tok.Where == nil && tok.Distinct == nil:
	case tok.Kind == grammar.BuiltinCall:
		draw, err = newAlphabetDraw(&o)
	default:
		draw, err = valueDrawOf(t, &o, l)
	}
	if err != nil {
		return tree.Op{}, err
	}
	return o.WithDraw(draw), nil
}

func compileArms(t *tree.Template, spellings []string, l *tree.LoadFinds) ([]tree.Arm, error) {
	if len(spellings) == 0 {
		return nil, nil
	}
	arms := make([]tree.Arm, len(spellings))
	for i, spelling := range spellings {
		a, err := compileArm(t, spelling, l)
		if err != nil {
			return nil, err
		}
		arms[i] = a
	}
	return arms, nil
}

// compileArm compiles one read into a path: from the head it names, or, for a read through a
// name, from what that name binds.
func compileArm(t *tree.Template, spelling string, l *tree.LoadFinds) (tree.Arm, error) {
	tl := l.Of(t)
	r := splitRead(spelling, tl.Refs().BySpelling())
	if !isName(t, r.head, l) {
		start := t.StartOf(r.head, tl.Refs())
		if start == nil {
			panic(invariant.Broken("{%s} reads a head nothing resolved", r.text))
		}
		w := CompilePath(start, r.tail)
		return tree.NewArm(r.text, r.head, r.tail, tree.FreshLevels(r.head, r.tail), w.steps, w.Leaves), nil
	}
	b := tl.NameScope().Lookup(r.head)
	tail, member := r.tail, 0
	if b.Distinct() != nil {
		var err error
		if member, err = memberRead(b, r); err != nil {
			return tree.Arm{}, err
		}
		tail = nil
	} else if grammar.HasSelector(tail) {
		if b.BindsField() {
			return tree.Arm{}, fmt.Errorf("a path through name %q may not select a row; bind the row to a name of its own", r.head)
		}
		return tree.Arm{}, fmt.Errorf("a path through name %q may not select a row; read the row without the name, {%s}, or bind it to a name of its own", r.head, grammar.JoinSegments(append([]string{b.Ref()}, tail...)))
	}
	target := l.Binding(b).Target()
	full := append(target.Tail()[:len(target.Tail()):len(target.Tail())], tail...)
	if err := provePath(target.Start(), full, r.head, len(target.Tail())); err != nil {
		return tree.Arm{}, err
	}
	w := CompilePath(target.Start(), full)
	return tree.NewArm(r.text, r.head, tail, tree.NamedLevels(full), w.steps, w.Leaves).ThroughName(b, member), nil
}

// memberRead is the member of b's distinct draw that a reads, counted from 1, or 0 where a reads them
// joined.
func memberRead(b *tree.NameBinding, a readSpelling) (int, error) {
	n := b.Distinct().N
	switch {
	case len(a.tail) == 0 && !b.Distinct().Joins:
		return 0, fmt.Errorf("name %q holds %d distinct members: read one, as {%s[1]}, or add joined ',' after distinct %d to read them as one", b.Name(), n, b.Name(), n)
	case len(a.tail) == 0:
		return 0, nil
	case len(a.tail) > 1 || !grammar.IsSelector(a.tail[0]):
		return 0, fmt.Errorf("name %q holds distinct members, each read whole, as {%s[1]}", b.Name(), b.Name())
	}
	k, err := strconv.Atoi(grammar.SelectorOf(a.tail[0]))
	if err != nil || k < 1 || k > n {
		return 0, fmt.Errorf("name %q holds %d members, {%s[1]} to {%s[%d]}", b.Name(), n, b.Name(), b.Name(), n)
	}
	return k, nil
}
