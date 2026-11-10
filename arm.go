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

// splitArm splits one spelling into head and tail. refs maps a reference to what
// resolveRefs resolved it to; before that, a reference is whole.
func splitArm(spelling string, refs map[string]resolvedRef) arm {
	if grammar.IsRef(spelling) {
		if r, resolved := refs[spelling]; resolved {
			return pathArm(spelling, r.head, r.tail)
		}
		return pathArm(spelling, spelling, nil)
	}
	segs, err := grammar.SplitPath(spelling)
	if err != nil || len(segs) == 1 {
		return pathArm(spelling, spelling, nil)
	}
	return pathArm(spelling, segs[0], segs[1:])
}

func pathArm(spelling, head string, segs []string) arm {
	return arm{spelling: spelling, head: head, tail: segs, levels: freshLevels(head, segs)}
}

// key is the key of the leaf a lands on, the one key every way of writing this read shares.
func (a arm) key() pickKey { return a.levels[len(a.levels)-1] }

func checkSegments(a arm) error {
	if len(a.tail) == 0 {
		return nil
	}
	return grammar.CheckSegments(append([]string{a.head}, a.tail...))
}

// op is one compiled unit of a format string: a literal run, a name read,
// or a builtin already prepared with its args.
type op struct {
	grammar.Token
	arms []arm // grammar.PathRead: the '|' alternatives, split into head and tail once
	call builtinfunc.Call
	// operands are the fields the builtin reads, in the order its operands func fixed; expand reads
	// them before a call that has no draw. nil for a token that reads none.
	operands []arm
	clause   []arm // the spellings the token's where clause reads, in clauseNames' order
	// draw is what a token with a where or a distinct clause, or a {reword()} call, prints in
	// place of a plain read or call; nil for any other token
	draw tokenDraw
}

// reads is every arm a render of o reads: its arms, its builtin's operands and its clause's spellings.
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

// tokenDraw is how a token draws what it prints, built at load once its paths resolve.
type tokenDraw interface {
	// draw prints one draw of the token in t; emitted is the output so far.
	draw(s *drawstate.State, t *template, env renderEnv, emitted string) string
	// mayFallBack says whether a draw may print the where clause's fallback, or a path's empty
	// default, where the strings it reads leave no value.
	mayFallBack() bool
}

// drawToken prints one draw of d. It calls draw on d's concrete type, since a call through the
// interface would move each render's frames to the heap.
func drawToken(s *drawstate.State, t *template, env renderEnv, d tokenDraw, emitted string) string {
	switch d := d.(type) {
	case *valueDraw:
		return d.draw(s, t, env, emitted)
	case *alphabetDraw:
		return d.draw(s, t, env, emitted)
	case *rewordDraw:
		return d.draw(s, t, env, emitted)
	}
	panic(invariant.Broken("no draw for %T", d))
}

func (t *template) compileArms(spellings []string, targets map[*nameBinding]nameTarget) ([]arm, error) {
	if len(spellings) == 0 {
		return nil, nil
	}
	arms := make([]arm, len(spellings))
	for i, spelling := range spellings {
		a, err := t.compileArm(spelling, targets)
		if err != nil {
			return nil, err
		}
		arms[i] = a
	}
	return arms, nil
}

// compileArm compiles one read into a path: from the head it names, or, for a read through a
// name, from what that name binds.
func (t *template) compileArm(spelling string, targets map[*nameBinding]nameTarget) (arm, error) {
	a := splitArm(spelling, t.refs.byName)
	if !t.isName(a.head) {
		start := t.startOf(a.head)
		if start == nil {
			panic(invariant.Broken("{%s} reads a head nothing resolved", a.spelling))
		}
		w := compilePath(start, a.tail)
		a.steps, a.leaves = w.steps, w.leaves
		return a, nil
	}
	b := t.nameScope.lookup(a.head)
	if b.distinct != nil {
		member, err := memberRead(b, a)
		if err != nil {
			return a, err
		}
		a.member, a.tail = member, nil
	} else if grammar.HasSelector(a.tail) {
		if b.bindsField() {
			return a, fmt.Errorf("a path through name %q may not select a row; bind the row to a name of its own", a.head)
		}
		return a, fmt.Errorf("a path through name %q may not select a row; read the row without the name, {%s}, or bind it to a name of its own", a.head, grammar.JoinSegments(append([]string{b.ref}, a.tail...)))
	}
	target := targets[b]
	full := append(target.tail[:len(target.tail):len(target.tail)], a.tail...)
	if err := provePath(target.start, full, a.head, len(target.tail)); err != nil {
		return a, err
	}
	w := compilePath(target.start, full)
	a.kind, a.named, a.steps, a.leaves = namedRead, b, w.steps, w.leaves
	a.levels = namedLevels(full)
	return a, nil
}

// memberRead is the member of b's distinct draw that a reads, counted from 1, or 0 where a reads them
// joined.
func memberRead(b *nameBinding, a arm) (int, error) {
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
