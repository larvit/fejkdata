package fejkdata

import (
	"fmt"
	"strings"
)

// scanUnit is one unit of a scanned format string: a literal rune or the body of a
// {…} token.
type scanUnit struct {
	isToken bool // a {…} body, else a literal rune
	char    rune
	body    string
}

// eachScanUnit scans a format string once and calls fn for each unit, the single
// source of truth for how braces are read: "{{" and "}}" are literal braces, a "{"
// opens a token that must reach its "}", and a lone "}" is an error.
func eachScanUnit(format string, fn func(scanUnit) error) error {
	rs := []rune(format)
	for i := 0; i < len(rs); i++ {
		var u scanUnit
		switch c := rs[i]; c {
		case '{':
			if i+1 < len(rs) && rs[i+1] == '{' {
				u.char = '{'
				i++
				break
			}
			end := i + 1
			for end < len(rs) && rs[end] != '}' {
				if rs[end] == '{' {
					return fmt.Errorf("'{' inside a token in %q; a literal brace is written {{", format)
				}
				end++
			}
			if end >= len(rs) {
				return fmt.Errorf("unterminated '{' in %q", format)
			}
			u.isToken, u.body = true, string(rs[i+1:end])
			i = end
		case '}':
			if i+1 < len(rs) && rs[i+1] == '}' {
				u.char = '}'
				i++
				break
			}
			return fmt.Errorf("lone '}' in %q; a literal brace is written }}", format)
		default:
			u.char = c
		}
		if err := fn(u); err != nil {
			return err
		}
	}
	return nil
}

// formatToken is one parsed unit of a format.
type formatToken struct {
	kind  tokenKind
	lit   string   // literalRun
	body  string   // the braces' content, as written
	fn    string   // builtinCall
	args  []string // builtinCall
	names []string // nameRead: the '|' arms; builtinCall: the operands its builtin reads
	// nameBind: the reference it binds, and the name.
	boundRef, bound string
}

type tokenKind uint8

const (
	builtinCall tokenKind = iota + 1
	nameRead
	literalRun
	nameBind
)

const asWord = " as "

// parseFormat is the one reading of a format's tokens: a '(' outside a selector makes
// a token a call.
func parseFormat(format string) ([]formatToken, error) {
	var toks []formatToken
	var lit strings.Builder
	flush := func() {
		if lit.Len() > 0 {
			toks = append(toks, formatToken{kind: literalRun, lit: lit.String()})
			lit.Reset()
		}
	}
	err := eachScanUnit(format, func(u scanUnit) error {
		if !u.isToken {
			lit.WriteRune(u.char)
			return nil
		}
		flush()
		if ref, name, binds := cutOutside(u.body, asWord); binds && indexOutside(u.body, '(') < 0 {
			toks = append(toks, formatToken{kind: nameBind, body: u.body, boundRef: ref, bound: name})
			return nil
		}
		if indexOutside(u.body, '(') < 0 {
			toks = append(toks, formatToken{kind: nameRead, body: u.body, names: splitOutside(u.body, '|')})
			return nil
		}
		name, args, ok := funcCall(u.body)
		if !ok {
			return fmt.Errorf("malformed function token {%s}", u.body)
		}
		toks = append(toks, formatToken{kind: builtinCall, body: u.body, fn: name, args: args, names: builtinOperands(name, args)})
		return nil
	})
	if err != nil {
		return nil, err
	}
	flush()
	return toks, nil
}

// builtin is a format-string function invoked as {name(args)}. It receives the
// generatorState, the output emitted so far in the current expansion (for derivations
// such as a checksum over preceding digits), and the values of the operands it named
// (calc and the transforms name them). All must stay pure over (rng, emitted, args) so
// seeded output is reproducible; seq advances per-generator counter state, which is
// itself deterministic. arity is the exact arg count, or -1 for variadic (then
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
	proveNumber func(token string, prints DataType, args []string) proven
	// prints is the datatype a call's text is, handed to proveNumber: DataTypeString where it
	// reads as a number no column should type, as digits' leading zeros; unset where
	// proveNumber is nil.
	prints DataType
}

// funcCall splits a "{token}" body shaped name(args) into its parts; ok is false
// for a name-read body. A '(' without a trailing ')' yields ok=false.
func funcCall(body string) (name string, args []string, ok bool) {
	lp := indexOutside(body, '(')
	if lp < 0 || !strings.HasSuffix(body, ")") {
		return "", nil, false
	}
	return body[:lp], splitArgs(body[lp+1 : len(body)-1]), true
}

// splitArgs parses a function arg list: comma-separated outside a selector or a
// quoted layout, trimmed; empty -> none.
func splitArgs(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var args []string
	depth, quoted, start := 0, false, 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '\'' && depth == 0:
			quoted = !quoted
		case quoted:
		case c == '[':
			depth++
		case c == ']' && depth > 0:
			depth--
		case c == ',' && depth == 0:
			args, start = append(args, strings.TrimSpace(s[start:i])), i+1
		}
	}
	return append(args, strings.TrimSpace(s[start:]))
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
func checkFunc(tok formatToken, fields map[string]node) error {
	body, name, args := tok.body, tok.fn, tok.args
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

func parseChecked(format string, fields map[string]node) ([]formatToken, []unboundRead, error) {
	toks, err := parseFormat(format)
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
func checkTokens(toks []formatToken, fields map[string]node) ([]unboundRead, error) {
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

func checkToken(t formatToken, fields map[string]node) ([]unboundRead, error) {
	switch t.kind {
	case builtinCall:
		if err := checkFunc(t, fields); err != nil {
			return nil, err
		}
		return checkReads(t, fields, true)
	case nameBind:
		return nil, checkBind(t)
	case nameRead:
		unbound, err := checkReads(t, fields, false)
		if err != nil {
			return nil, err
		}
		// Last, so an arm broken on its own terms is reported as that: a repeat is
		// the consequence of such a mistake, not the mistake itself.
		return unbound, checkNoRepeatedArm(t.body, t.names)
	}
	return nil, nil
}

// checkReads proves each name t reads, an arm or an operand, is a reference or a path into a
// field, returning those whose head no field holds.
func checkReads(t formatToken, fields map[string]node, operands bool) ([]unboundRead, error) {
	var unbound []unboundRead
	for _, name := range t.names {
		if isRef(name) {
			if _, _, err := refShape(name); err != nil {
				return nil, fmt.Errorf("token {%s}: %w", t.body, err)
			}
			continue // its target is checked at New (see linkRefs)
		}
		missing, err := checkArm(name, fields, !operands && len(t.names) == 1)
		if err == nil {
			continue
		}
		err = fmt.Errorf("token {%s}: %w", t.body, err)
		if !missing {
			return nil, err
		}
		a := splitArm(name, nil)
		unbound = append(unbound, unboundRead{head: a.head, tail: joinSegments(a.tail), body: t.body, operand: operands, noRef: operands && builtins[t.fn].noRefOperands, err: err})
	}
	return unbound, nil
}

// checkBind proves a {ref as name} token is spelled once, binds a reference, and names a valid
// name that is no option.
func checkBind(t formatToken) error {
	ref, name := t.boundRef, t.bound
	if trimmed := strings.TrimSpace(ref) + asWord + strings.TrimSpace(name); trimmed != t.body {
		return fmt.Errorf("token {%s}: write {%s}", t.body, trimmed)
	}
	if !isRef(ref) {
		return fmt.Errorf("token {%s}: %q is no reference; a name binds a pick of a reference, which starts with /, . or ..", t.body, ref)
	}
	if _, _, err := refShape(ref); err != nil {
		return fmt.Errorf("token {%s}: %w", t.body, err)
	}
	if name == "" {
		return fmt.Errorf("token {%s}: a binding names nothing; write {%s as n}", t.body, ref)
	}
	if isOption(name) {
		return fmt.Errorf("token {%s}: %q is an option and can never be a name; rename it", t.body, name)
	}
	if err := checkName(name); err != nil {
		return fmt.Errorf("token {%s}: name %w", t.body, err)
	}
	return nil
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
	return checkPathNames(name) == nil
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
// `template.head` resolves and the tail of a dotted path into it. A non-empty tail
// makes the arm a held draw: a reference path's for the render, any other for the
// expansion.
type arm struct {
	spelling    string // as written, for messages
	head        string
	writtenHead string // head as written, sigil included
	tail        []string
	levels      []string // the path at each level the tail passes through, the head first
	path        string   // head and tail, the one path every way of writing this read shares
	steps       []pathStep
	leaves      []node // every node the path may land on, one per variant it passes
	cover       node   // what holding the path pins: the first choice it passes, else its leaf
	kind        armKind
	named       *nameBinding // namedRead: the binding of the name it reads through
}

// armKind is how expand reads an arm, fixed at compile.
type armKind uint8

const (
	freshRead   armKind = iota // drawn afresh at every read
	heldRead                   // drawn once per expansion, kept in its hold
	refPathRead                // drawn once per draw group, kept in its memo
	namedRead                  // read through a name, kept in its pick
)

func (a arm) isRefPath() bool {
	return isRef(a.head) && len(a.tail) > 0
}

// splitArm splits one name into head and tail. refs maps a reference to what
// linkRefs resolved it to; before linking, a reference is whole.
func splitArm(name string, refs map[string]refBinding) arm {
	if isRef(name) {
		b, linked := refs[name]
		if !linked || len(b.tail) == 0 {
			head := name
			if linked {
				head = b.head
			}
			return arm{spelling: name, head: head, path: head}
		}
		sigil, rest, _ := refShape(name) // resolveLink proved it, and took b.tail as a suffix of its segments
		written, _ := splitPath(rest)
		return pathArm(name, b.head, sigil+joinSegments(written[:len(written)-len(b.tail)]), b.tail)
	}
	segs, err := splitPath(name)
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

// checkSegments rejects an unfinished path: "{a.}" and "{a...b}" each have a
// segment naming nothing. A field really named "" would otherwise make them
// resolve, so a typo would read as a path that worked.
func checkSegments(a arm) error {
	if len(a.tail) == 0 {
		return nil
	}
	if a.head == "" {
		return fmt.Errorf("path has an empty segment")
	}
	for _, seg := range a.tail {
		if seg == "" {
			return fmt.Errorf("path has an empty segment")
		}
	}
	return nil
}

// callFn is a builtin prepared for one call site: its args already parsed. It reads the
// output emitted so far in the current expansion (a derivation's payload) and the
// values of the operands it named, which expand read for it.
type callFn func(s *generatorState, emitted string, operands []string) string

// op is one compiled unit of a format string: a literal run, a name read,
// or a builtin already prepared with its args.
type op struct {
	formatToken
	arms []arm // nameRead: the '|' alternatives, split into head and tail once
	call callFn
	// operands are the fields the builtin reads, in the order its operands func
	// fixed; expand reads them before the call. nil for a builtin that reads none.
	operands []arm
}

// formatOps is a compiled format: its ops, the size of its literal text, and the
// names its expansion holds: every head a path that is not a reference starts from,
// plus the fields an operand reads.
type formatOps struct {
	ops  []op
	grow int
	held map[string]firstReach
}

// firstReach is how a format first reaches a held name: the reader holding it, for
// error messages, and the first path starting from it, "" where none does.
type firstReach struct {
	holder string
	path   string
}

func (c *formatOps) holdName(a arm, label string) {
	if a.isRefPath() {
		return
	}
	if c.held == nil {
		c.held = map[string]firstReach{}
	}
	h, seen := c.held[a.head]
	if !seen {
		h.holder = label
	}
	if len(a.tail) > 0 && h.path == "" {
		h.path = a.spelling
	}
	c.held[a.head] = h
}

func (c *formatOps) function(tok formatToken, refs map[string]refBinding, isName func(string) bool) {
	var operands []arm
	for _, operand := range tok.names {
		a := splitArm(operand, refs)
		if isName(a.head) {
			a.kind = namedRead
			operands = append(operands, a)
			continue
		}
		c.holdName(a, fmt.Sprintf("%s operand %q", tok.fn, operand))
		operands = append(operands, a)
	}
	c.ops = append(c.ops, op{formatToken: tok, call: builtins[tok.fn].prep(tok.args), operands: operands})
}

func (c *formatOps) field(tok formatToken, refs map[string]refBinding, isName func(string) bool) {
	arms := make([]arm, len(tok.names))
	for i, name := range tok.names {
		arms[i] = splitArm(name, refs)
		if isName(arms[i].head) {
			arms[i].kind = namedRead
			continue
		}
		if len(arms[i].tail) > 0 {
			c.holdName(arms[i], "token {"+arms[i].spelling+"}")
		}
	}
	c.ops = append(c.ops, op{formatToken: tok, arms: arms})
}

// compileFormat compiles a parsed format; a {ref as n} token compiles to no op. Call
// checkTokens first: it is what proves every token valid.
func compileFormat(toks []formatToken, refs map[string]refBinding, isName func(head string) bool) formatOps {
	var c formatOps
	for _, tok := range toks {
		switch tok.kind {
		case literalRun:
			c.grow += len(tok.lit)
			c.ops = append(c.ops, op{formatToken: tok})
		case builtinCall:
			c.function(tok, refs, isName)
		case nameRead:
			c.field(tok, refs, isName)
		}
	}
	return c
}
