package fejkdata

import (
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// checkFunc validates a call at compile time: a known builtin, its args, and, for calc, the
// operands held in fields.
func checkFunc(tok grammar.Token, fields map[string]node) error {
	err := builtinfunc.Check(tok.Fn, tok.Args)
	if err == nil && tok.Fn == builtinfunc.CalcName {
		err = checkCalcFields(tok.Args, fields)
	}
	if err != nil {
		return fmt.Errorf("token {%s}: %w", tok.Body, err)
	}
	return nil
}

func parseChecked(format string, fields map[string]node) ([]grammar.Token, []unboundRead, error) {
	toks, err := grammar.ParseFormat(format)
	if err != nil {
		return nil, nil, err
	}
	unbound, err := checkTokens(toks, fields)
	return toks, unbound, err
}

// unboundRead is a token reading a head no field holds, which only a name bound around the
// template can answer; err is the refusal where none does, and body is the token's.
type unboundRead struct {
	head, body string
	err        error
}

// checkTokens proves every token's grammar, functions and field paths, and returns the reads
// whose head no field holds, for bindNames to look up among the names.
func checkTokens(toks []grammar.Token, fields map[string]node) ([]unboundRead, error) {
	var unbound []unboundRead
	for _, t := range toks {
		u, err := checkToken(t, fields)
		if err != nil {
			return nil, err
		}
		unbound = append(unbound, u...)
	}
	return unbound, nil
}

func checkToken(t grammar.Token, fields map[string]node) ([]unboundRead, error) {
	switch t.Kind {
	case grammar.BuiltinCall:
		if err := checkFunc(t, fields); err != nil {
			return nil, err
		}
		return checkReads(t, fields, true)
	case grammar.NameBind:
		return nil, checkBind(t, fields)
	case grammar.NameRead:
		unbound, err := checkReads(t, fields, false)
		if err != nil {
			return nil, err
		}
		// Last, so an arm broken on its own terms is reported as that: a repeat is
		// the consequence of such a mistake, not the mistake itself.
		return unbound, checkNoRepeatedArm(t.Body, t.Arms)
	}
	return nil, nil
}

// checkReads proves each name t reads, an arm or an operand, is a reference or a path into a
// field, returning those whose head no field holds.
func checkReads(t grammar.Token, fields map[string]node, operands bool) ([]unboundRead, error) {
	var unbound []unboundRead
	names := tokenReads(t)
	for _, name := range names {
		if grammar.IsRef(name) {
			if _, _, err := grammar.RefShape(name); err != nil {
				return nil, fmt.Errorf("token {%s}: %w", t.Body, err)
			}
			continue // resolveRefs checks its target once the tree is assembled
		}
		missing, err := checkArm(name, fields, !operands && len(names) == 1)
		if err == nil {
			continue
		}
		err = fmt.Errorf("token {%s}: %w", t.Body, err)
		if !missing {
			return nil, err
		}
		unbound = append(unbound, unboundRead{head: splitArm(name, nil).head, body: t.Body, err: err})
	}
	return unbound, nil
}

// checkBind proves a {x as name} token is spelled once, binds a reference or a path into a
// field, and names a valid name that is no option.
func checkBind(t grammar.Token, fields map[string]node) error {
	ref, name := t.BoundRef, t.Bound
	if trimmed := strings.TrimSpace(ref) + grammar.AsWord + strings.TrimSpace(name); trimmed != t.Body {
		return fmt.Errorf("token {%s}: write {%s}", t.Body, trimmed)
	}
	if err := checkBound(ref, fields); err != nil {
		return fmt.Errorf("token {%s}: %w", t.Body, err)
	}
	if name == "" {
		return fmt.Errorf("token {%s}: a binding names nothing; write {%s as n}", t.Body, ref)
	}
	if isOption(name) {
		return fmt.Errorf("token {%s}: %q is an option and can never be a name; rename it", t.Body, name)
	}
	if err := grammar.CheckName(name); err != nil {
		return fmt.Errorf("token {%s}: name %w", t.Body, err)
	}
	return nil
}

// checkBound proves what a binding names is a reference or a path into a field of the binding
// template.
func checkBound(ref string, fields map[string]node) error {
	if grammar.IsRef(ref) {
		_, _, err := grammar.RefShape(ref)
		return err
	}
	missing, err := checkArm(ref, fields, false)
	if missing {
		return fmt.Errorf("%w; a binding names a field, or a reference, which starts with /, . or ..", err)
	}
	return err
}

// checkArm validates one sibling name or path against a template's fields.
// wholeToken says the name is the token's entire body, so {/name} would render
// the same value and can be offered as the reference spelling. missing says no
// field holds the head, which a name may yet answer.
func checkArm(name string, fields map[string]node, wholeToken bool) (missing bool, err error) {
	a := splitArm(name, nil)
	if err := checkSegments(a); err != nil {
		return false, err
	}
	field, ok := fields[a.head]
	if !ok {
		if a.head == "" {
			return false, fmt.Errorf("a name is never empty, so this token can name no field")
		}
		if isOption(a.head) {
			return false, fmt.Errorf("%q is an option and can never be a field", a.head)
		}
		if len(fields) == 0 {
			hint := ""
			if wholeToken && hintableRef(name) {
				hint = fmt.Sprintf(" — write {/%s} to reference the data", name)
			}
			return true, fmt.Errorf("no field %q; a token names a sibling field, and this template has none%s", a.head, hint)
		}
		return true, fmt.Errorf("no field %q", a.head)
	}
	if err := checkPathResolves(field, a.tail, a.head); err != nil {
		return false, fmt.Errorf("field %q: %w", a.head, err)
	}
	return false, nil
}

// hintableRef reports whether {/name} is a reference the grammar accepts, so the
// hint never names a spelling that fails too.
func hintableRef(name string) bool {
	return grammar.CheckPathNames(name) == nil
}

// tokenReads lists the names t reads: a read's arms, or the operands a call's builtin reads.
func tokenReads(t grammar.Token) []string {
	if t.Kind == grammar.BuiltinCall {
		return builtinfunc.Operands(t.Fn, t.Args)
	}
	return t.Arms
}

// checkNoRepeatedArm rejects {a|a|b}: an alternation draws its arms evenly, so a
// repeated one is a second spelling of weight. The error names the spelling that
// does skew a draw.
func checkNoRepeatedArm(body string, names []string) error {
	if len(names) < 2 {
		return nil
	}
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if seen[name] {
			return fmt.Errorf("token {%s}: arm %q is repeated; an alternation picks its arms evenly, so skew the odds with a choice's weights instead", body, name)
		}
		seen[name] = true
	}
	return nil
}

// arm is one alternative of a {a|b} token or one operand, split into the head,
// whose node `template.startOf` finds, and the tail of a dotted path into it.
type arm struct {
	spelling    string // as written, for messages
	head        string
	writtenHead string // head as written, sigil included
	tail        []string
	levels      []pickKey // the key of each level the steps pass, from the head they start at to the leaf
	steps       []pathStep
	leaves      []node // every node the path may land on, one per variant it passes
	kind        armKind
	named       *nameBinding // namedRead: the binding of the name it reads through
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
		b, linked := refs[name]
		if !linked || len(b.tail) == 0 {
			head := name
			if linked {
				head = b.head
			}
			return arm{spelling: name, head: head, levels: []pickKey{pickKey(head)}}
		}
		sigil, rest, _ := grammar.RefShape(name) // resolveRefs proved it, and took b.tail as a suffix of its segments
		written, _ := grammar.SplitPath(rest)
		return pathArm(name, b.head, sigil+grammar.JoinSegments(written[:len(written)-len(b.tail)]), b.tail)
	}
	segs, err := grammar.SplitPath(name)
	if err != nil || len(segs) == 1 {
		return arm{spelling: name, head: name, levels: []pickKey{pickKey(name)}}
	}
	return pathArm(name, segs[0], segs[0], segs[1:])
}

func pathArm(name, head, writtenHead string, segs []string) arm {
	return arm{spelling: name, head: head, writtenHead: writtenHead, tail: segs, levels: levelKeys(append([]string{head}, segs...), 1)}
}

// levelKeys is the key of each prefix of path holding at least from segments, shortest first.
func levelKeys(path []string, from int) []pickKey {
	levels := make([]pickKey, len(path)-from+1)
	for i := range levels {
		levels[i] = pickKey(grammar.JoinSegments(path[:from+i]))
	}
	return levels
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
	arms []arm // grammar.NameRead: the '|' alternatives, split into head and tail once
	call builtinfunc.Call
	// operands are the fields the builtin reads, in the order its operands func
	// fixed; expand reads them before the call. nil for a builtin that reads none.
	operands []arm
}

// formatOps is a compiled format: its ops, and the size of its literal text.
type formatOps struct {
	ops  []op
	grow int
}

// compileFormat compiles t's parsed format against its resolved references and its names'
// targets; a {ref as n} token compiles to no op. checkTokens proved every token valid.
func compileFormat(t *template, targets map[*nameBinding]nameTarget) (formatOps, error) {
	var c formatOps
	for _, tok := range t.tokens {
		o := op{Token: tok}
		var err error
		switch tok.Kind {
		case grammar.LiteralRun:
			c.grow += len(tok.Lit)
		case grammar.BuiltinCall:
			o.call = builtinfunc.Prep(tok.Fn, tok.Args)
			o.operands, err = t.compileArms(builtinfunc.Operands(tok.Fn, tok.Args), targets)
		case grammar.NameRead:
			o.arms, err = t.compileArms(tok.Arms, targets)
		case grammar.NameBind:
			continue
		}
		if err != nil {
			return c, fmt.Errorf("token {%s}: %w", tok.Body, err)
		}
		c.ops = append(c.ops, o)
	}
	return c, nil
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
			panic(invariant.Broken("{%s} reads a head nothing bound", a.spelling))
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
	if err := checkPathResolves(target.start, full, a.head); err != nil {
		return a, err
	}
	w := compilePath(target.start, full)
	a.kind, a.named, a.steps, a.leaves = namedRead, b, w.steps, w.leaves
	a.levels = levelKeys(full, 0)
	return a, nil
}
