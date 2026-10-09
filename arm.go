package fejkdata

import (
	"fmt"

	"github.com/larvit/fejkdata/internal/builtinfunc"
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
}

// armKind is how expand reads an arm, fixed at compile.
type armKind uint8

const (
	freshRead armKind = iota // drawn afresh at every read
	namedRead                // read through a name, kept in its pick
)

// splitArm splits one name into head and tail. refs maps a reference to what
// resolveRefs resolved it to; before that, a reference is whole.
func splitArm(name string, refs map[string]resolvedRef) arm {
	if grammar.IsRef(name) {
		r, resolved := refs[name]
		if !resolved || len(r.tail) == 0 {
			head := name
			if resolved {
				head = r.head
			}
			return arm{spelling: name, head: head, levels: levelKeys([]string{head}, 1)}
		}
		return pathArm(name, r.head, r.tail)
	}
	segs, err := grammar.SplitPath(name)
	if err != nil || len(segs) == 1 {
		return arm{spelling: name, head: name, levels: levelKeys([]string{name}, 1)}
	}
	return pathArm(name, segs[0], segs[1:])
}

func pathArm(name, head string, segs []string) arm {
	return arm{spelling: name, head: head, tail: segs, levels: levelKeys(append([]string{head}, segs...), 1)}
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
	// operands are the fields the builtin reads, in the order its operands func
	// fixed; expand reads them before the call. nil for a builtin that reads none.
	operands []arm
}

func (t *template) compileArms(names []string, targets map[*nameBinding]nameTarget) ([]arm, error) {
	if len(names) == 0 {
		return nil, nil
	}
	arms := make([]arm, len(names))
	for i, name := range names {
		a, err := t.compileArm(name, targets)
		if err != nil {
			return nil, err
		}
		arms[i] = a
	}
	return arms, nil
}

// compileArm compiles one read into a path: from the head it names, or, for a read through a
// name, from what that name binds.
func (t *template) compileArm(name string, targets map[*nameBinding]nameTarget) (arm, error) {
	a := splitArm(name, t.refs.byName)
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
	if grammar.HasSelector(a.tail) {
		return a, fmt.Errorf("a path through name %q may not select a row; read it directly, {%s.%s}, or bind the row to a name of its own", a.head, b.ref, grammar.JoinSegments(a.tail))
	}
	target := targets[b]
	full := append(target.tail[:len(target.tail):len(target.tail)], a.tail...)
	if err := provePath(target.start, full, a.head); err != nil {
		return a, err
	}
	w := compilePath(target.start, full)
	a.kind, a.named, a.steps, a.leaves = namedRead, b, w.steps, w.leaves
	a.levels = levelKeys(full, 0)
	return a, nil
}
