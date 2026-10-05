package fejkdata

import (
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/proven"
)

// builtin is a format-string function invoked as {name(args)}. It receives the
// draw state, the output emitted so far in the current expansion (for derivations
// such as a checksum over preceding digits), and the values of the operands it named
// (calc and the transforms name them). All must stay pure over (draw state, emitted, operands) so
// seeded output is reproducible. arity is the exact arg count, or -1 for variadic (then
// checkArgs does all the validation).
type builtin struct {
	arity int
	// prep parses validated args once, at compile time, into the closure expand calls.
	prep      func(args []string) callFn
	checkArgs func(fields map[string]node, args []string) error
	// operands names the fields the call reads, which expand renders for it; nil
	// for a builtin that reads none.
	operands func(args []string) []string
	// noRefOperands says no operand can be a reference, so a name read once there is not refused.
	noRefOperands bool
	// proveNumber bounds the number a call's text reads as, token its body, and says which
	// datatypes that text is not; set it where every render reads as a finite number, which
	// makes the call a calc operand, and leave it nil otherwise.
	proveNumber func(token string, prints DataType, args []string) proven.Value
	// prints is the datatype a call's text is, handed to proveNumber: DataTypeString where it
	// reads as a number no column should type, as digits' leading zeros; unset where
	// proveNumber is nil.
	prints DataType
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// checkFunc validates a call at compile time: naming a known builtin, with the arg
// count that builtin takes and args its check accepts. fields is passed through for
// the builtins (calc, the transforms) that validate against them.
func checkFunc(tok grammar.Token, fields map[string]node) error {
	body, name, args := tok.Body, tok.Fn, tok.Args
	b, known := builtins[name]
	if !known {
		return fmt.Errorf("token {%s}: unknown function %q", body, name)
	}
	if b.arity >= 0 && len(args) != b.arity {
		return fmt.Errorf("token {%s}: %s takes %d argument%s, got %d", body, name, b.arity, plural(b.arity), len(args))
	}
	if b.checkArgs != nil {
		if err := b.checkArgs(fields, args); err != nil {
			return fmt.Errorf("token {%s}: %w", body, err)
		}
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
// template can answer; err is the refusal where none does. tail is the path read into the name,
// "" for the name itself, operand marks a builtin's read of it, noRef marks such a read where no
// reference can stand, and body is the token's.
type unboundRead struct {
	head, tail, body string
	operand, noRef   bool
	err              error
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
			continue // its target is checked at New (see linkRefs)
		}
		missing, err := checkArm(name, fields, !operands && len(names) == 1)
		if err == nil {
			continue
		}
		err = fmt.Errorf("token {%s}: %w", t.Body, err)
		if !missing {
			return nil, err
		}
		a := splitArm(name, nil)
		unbound = append(unbound, unboundRead{head: a.head, tail: grammar.JoinSegments(a.tail), body: t.Body, operand: operands, noRef: operands && builtins[t.Fn].noRefOperands, err: err})
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

// checkBound proves what a binding picks is a reference or a path into a field of the binding
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
		return builtinOperands(t.Fn, t.Args)
	}
	return t.Arms
}

// builtinOperands lists the fields a call reads as operands, empty for a builtin that
// reads none or is unknown.
func builtinOperands(name string, args []string) []string {
	b, known := builtins[name]
	if !known || b.operands == nil {
		return nil
	}
	return b.operands(args)
}

// checkNoRepeatedArm rejects {a|a|b}: an alternation picks its arms evenly, so a
// repeated one is a second spelling of weight. The error names the spelling that
// does skew a pick.
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

// arm is one alternative of a {a|b} token or one operand, split into the head
// `template.head` resolves and the tail of a dotted path into it.
type arm struct {
	spelling    string // as written, for messages
	head        string
	writtenHead string // head as written, sigil included
	tail        []string
	levels      []string // the path at each level the tail passes through, the head first
	path        string   // head and tail, the one path every way of writing this read shares
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
// linkRefs resolved it to; before linking, a reference is whole.
func splitArm(name string, refs map[string]refBinding) arm {
	if grammar.IsRef(name) {
		b, linked := refs[name]
		if !linked || len(b.tail) == 0 {
			head := name
			if linked {
				head = b.head
			}
			return arm{spelling: name, head: head, path: head}
		}
		sigil, rest, _ := grammar.RefShape(name) // resolveLink proved it, and took b.tail as a suffix of its segments
		written, _ := grammar.SplitPath(rest)
		return pathArm(name, b.head, sigil+grammar.JoinSegments(written[:len(written)-len(b.tail)]), b.tail)
	}
	segs, err := grammar.SplitPath(name)
	if err != nil || len(segs) == 1 {
		return arm{spelling: name, head: name, path: name}
	}
	return pathArm(name, segs[0], segs[0], segs[1:])
}

func pathArm(name, head, writtenHead string, segs []string) arm {
	levels := []string{head}
	for i := 0; i < len(segs)-1; i++ {
		levels = append(levels, head+"."+strings.Join(segs[:i+1], "."))
	}
	return arm{spelling: name, head: head, writtenHead: writtenHead, tail: segs, levels: levels, path: head + "." + strings.Join(segs, ".")}
}

func checkSegments(a arm) error {
	if len(a.tail) == 0 {
		return nil
	}
	return grammar.CheckSegments(append([]string{a.head}, a.tail...))
}

// callFn is a builtin prepared for one call site: its args already parsed. It reads the
// output emitted so far in the current expansion (a derivation's payload) and the
// values of the operands it named, which expand read for it.
type callFn func(s *drawstate.State, emitted string, operands []string) string

// op is one compiled unit of a format string: a literal run, a name read,
// or a builtin already prepared with its args.
type op struct {
	grammar.Token
	arms []arm // grammar.NameRead: the '|' alternatives, split into head and tail once
	call callFn
	// operands are the fields the builtin reads, in the order its operands func
	// fixed; expand reads them before the call. nil for a builtin that reads none.
	operands []arm
}

// formatOps is a compiled format: its ops, and the size of its literal text.
type formatOps struct {
	ops  []op
	grow int
}

func (c *formatOps) function(tok grammar.Token, refs map[string]refBinding, isName func(string) bool) {
	var operands []arm
	for _, operand := range builtinOperands(tok.Fn, tok.Args) {
		a := splitArm(operand, refs)
		if isName(a.head) {
			a.kind = namedRead
		}
		operands = append(operands, a)
	}
	c.ops = append(c.ops, op{Token: tok, call: builtins[tok.Fn].prep(tok.Args), operands: operands})
}

func (c *formatOps) field(tok grammar.Token, refs map[string]refBinding, isName func(string) bool) {
	arms := make([]arm, len(tok.Arms))
	for i, name := range tok.Arms {
		arms[i] = splitArm(name, refs)
		if isName(arms[i].head) {
			arms[i].kind = namedRead
		}
	}
	c.ops = append(c.ops, op{Token: tok, arms: arms})
}

// compileFormat compiles a parsed format; a {ref as n} token compiles to no op. Call
// checkTokens first: it is what proves every token valid.
func compileFormat(toks []grammar.Token, refs map[string]refBinding, isName func(head string) bool) formatOps {
	var c formatOps
	for _, tok := range toks {
		switch tok.Kind {
		case grammar.LiteralRun:
			c.grow += len(tok.Lit)
			c.ops = append(c.ops, op{Token: tok})
		case grammar.BuiltinCall:
			c.function(tok, refs, isName)
		case grammar.NameRead:
			c.field(tok, refs, isName)
		}
	}
	return c
}
