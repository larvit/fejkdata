package fejkdata

import (
	"fmt"
	"strconv"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// arm is one alternative of a {a|b} token or one operand, split into the head,
// whose node `template.startOf` finds, and the tail of a dotted path into it.
type arm struct {
	spelling string // as written, for messages
	head     string
	tail     []string
	levels   []pickKey // the key of each level the steps pass, from the head they start at to the leaf; a step's at indexes it
	steps    []pathStep
	leaves   []node // every node the path may land on, one per variant it passes
	kind     armKind
	named    *nameBinding // namedRead: the binding of the name it reads through
	member   int          // namedRead of a distinct draw: the member it reads, counted from 1; 0 reads them joined
}

// armKind is how expand reads an arm, fixed at compile.
type armKind uint8

const (
	freshRead armKind = iota // drawn afresh at every read
	namedRead                // read through a name, kept in its pick
)

// readSpelling is one spelling split into the head, whose node `template.startOf` finds, and the
// tail of a dotted path into it. refs maps a reference to what resolveRefs resolved it to; before
// that, a reference is whole.
type readSpelling struct {
	text string
	head string
	tail []string
}

func splitRead(spelling string, refs map[string]resolvedRef) readSpelling {
	if grammar.IsRef(spelling) {
		if r, resolved := refs[spelling]; resolved {
			return readSpelling{spelling, r.head, r.tail}
		}
		return readSpelling{spelling, spelling, nil}
	}
	segs, err := grammar.SplitPath(spelling)
	if err != nil || len(segs) == 1 {
		return readSpelling{spelling, spelling, nil}
	}
	return readSpelling{spelling, segs[0], segs[1:]}
}

// key is the key of the leaf a lands on, the one key every way of writing this read shares.
func (a arm) key() pickKey { return a.levels[len(a.levels)-1] }

func checkSegments(r readSpelling) error {
	if len(r.tail) == 0 {
		return nil
	}
	return grammar.CheckSegments(append([]string{r.head}, r.tail...))
}

// op is one compiled unit of a format string: a literal run, a read,
// or a builtin already prepared with its args.
type op struct {
	grammar.Token
	arms []arm // grammar.PathRead: the '|' alternatives, split into head and tail once
	call builtinfunc.Call
	// operands are the fields the builtin reads, in the order its operands func fixed; expand reads
	// them before a call that has no draw. nil for a token that reads none.
	operands []arm
	clause   []arm // the token's where clause's reads, in clauseSpellings' order
	// draw is what a token with a where or a distinct clause, or a {reword()} call, prints in
	// place of a plain read or call; nil for any other token
	draw tokenDraw
}

// newOp returns parts whole: with its call, and with its draw where it has one. parts is an op whose
// arms, operands and clause compileFormat built. An adddays call takes the one layout its date
// operand prints in, so every template's format must be compiled first.
func newOp(t *template, parts op, l *loadFinds, words map[wordsAt]*wordList) (op, error) {
	o := parts
	if o.Kind == grammar.BuiltinCall {
		o.call = builtinfunc.Prep(o.Fn, o.Args)
		if o.Fn == builtinfunc.AddDaysName {
			layout, err := addDaysLayout(o.operands[0], l)
			if err != nil {
				return op{}, err
			}
			o.call = builtinfunc.AddDays(layout, o.Args)
		}
	}
	var err error
	switch {
	case o.Kind == grammar.BuiltinCall && o.Fn == builtinfunc.RewordName:
		o.draw, err = rewordDrawOf(t, &o, l, words)
	case o.Where == nil && o.Distinct == nil:
	case o.Kind == grammar.BuiltinCall:
		o.draw, err = newAlphabetDraw(&o)
	default:
		o.draw, err = valueDrawOf(t, &o, l)
	}
	if err != nil {
		return op{}, err
	}
	return o, nil
}

// reads is every arm a render of o reads: its arms, its builtin's operands and its clause's reads.
// Each points into o's slices, so the same arm always yields the same pointer.
func (o *op) reads() []*arm {
	found := make([]*arm, 0, len(o.arms)+len(o.operands)+len(o.clause))
	for _, arms := range [][]arm{o.arms, o.operands, o.clause} {
		for i := range arms {
			found = append(found, &arms[i])
		}
	}
	return found
}

// tokenDraw is a token's draw, one of the types drawToken switches on, built at load once its
// paths resolve.
type tokenDraw interface {
	// mayFallBack says whether a draw may print the where clause's fallback, or a path's empty
	// default, where the strings it reads leave no value.
	mayFallBack() bool
}

// drawToken prints one draw of d in t; emitted is the output so far.
func drawToken(s *drawstate.State, t *template, env renderEnv, d tokenDraw, emitted string) string {
	switch d := d.(type) {
	case *valueDraw:
		return drawValue(d, s, env)
	case *alphabetDraw:
		return drawAlphabet(d, s, env, emitted)
	case *rewordDraw:
		return drawReword(d, s, t, env)
	}
	panic(invariant.Broken("no draw for %T", d))
}

func compileArms(t *template, spellings []string, l *loadFinds) ([]arm, error) {
	if len(spellings) == 0 {
		return nil, nil
	}
	arms := make([]arm, len(spellings))
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
func compileArm(t *template, spelling string, l *loadFinds) (arm, error) {
	tl := l.of(t)
	r := splitRead(spelling, tl.refs.bySpelling)
	if !isName(t, r.head, l) {
		start := t.startOf(r.head, tl.refs)
		if start == nil {
			panic(invariant.Broken("{%s} reads a head nothing resolved", r.text))
		}
		w := compilePath(start, r.tail)
		return arm{spelling: r.text, head: r.head, tail: r.tail, levels: freshLevels(r.head, r.tail), steps: w.steps, leaves: w.leaves}, nil
	}
	b := tl.nameScope.lookup(r.head)
	tail, member := r.tail, 0
	if b.distinct != nil {
		var err error
		if member, err = memberRead(b, r); err != nil {
			return arm{}, err
		}
		tail = nil
	} else if grammar.HasSelector(tail) {
		if b.bindsField() {
			return arm{}, fmt.Errorf("a path through name %q may not select a row; bind the row to a name of its own", r.head)
		}
		return arm{}, fmt.Errorf("a path through name %q may not select a row; read the row without the name, {%s}, or bind it to a name of its own", r.head, grammar.JoinSegments(append([]string{b.ref}, tail...)))
	}
	target := l.binding(b).target
	full := append(target.tail[:len(target.tail):len(target.tail)], tail...)
	if err := provePath(target.start, full, r.head, len(target.tail)); err != nil {
		return arm{}, err
	}
	w := compilePath(target.start, full)
	return arm{spelling: r.text, head: r.head, tail: tail, levels: namedLevels(full), steps: w.steps, leaves: w.leaves, kind: namedRead, named: b, member: member}, nil
}

// memberRead is the member of b's distinct draw that a reads, counted from 1, or 0 where a reads them
// joined.
func memberRead(b *nameBinding, a readSpelling) (int, error) {
	n := b.distinct.N
	switch {
	case len(a.tail) == 0 && !b.distinct.Joins:
		return 0, fmt.Errorf("name %q holds %d distinct members: read one, as {%s[1]}, or add joined ',' after distinct %d to read them as one", b.name, n, b.name, n)
	case len(a.tail) == 0:
		return 0, nil
	case len(a.tail) > 1 || !grammar.IsSelector(a.tail[0]):
		return 0, fmt.Errorf("name %q holds distinct members, each read whole, as {%s[1]}", b.name, b.name)
	}
	k, err := strconv.Atoi(grammar.SelectorOf(a.tail[0]))
	if err != nil || k < 1 || k > n {
		return 0, fmt.Errorf("name %q holds %d members, {%s[1]} to {%s[%d]}", b.name, n, b.name, b.name, n)
	}
	return k, nil
}
