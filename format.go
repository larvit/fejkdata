package fejkdata

import (
	"fmt"
	"strings"
)

// ftoken is one unit of a scanned format string: a literal rune or the body of a
// {…} token.
type ftoken struct {
	kind byte // 'l' literal rune, 't' token body
	char rune
	body string
}

// eachToken scans a format string once and calls fn for each unit, the single
// source of truth for how braces are read: "{{" and "}}" are literal braces, a "{"
// opens a token that must reach its "}", and a lone "}" is an error. A table-shaped
// scanner, one case per rune kind, kept whole on purpose.
func eachToken(format string, fn func(ftoken) error) error {
	rs := []rune(format)
	for i := 0; i < len(rs); i++ {
		var t ftoken
		switch c := rs[i]; c {
		case '{':
			if i+1 < len(rs) && rs[i+1] == '{' {
				t.kind, t.char = 'l', '{'
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
			t.kind, t.body = 't', string(rs[i+1:end])
			i = end
		case '}':
			if i+1 < len(rs) && rs[i+1] == '}' {
				t.kind, t.char = 'l', '}'
				i++
				break
			}
			return fmt.Errorf("lone '}' in %q; a literal brace is written }}", format)
		default:
			t.kind, t.char = 'l', c
		}
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}

// formatToken is one parsed unit of a format: a literal run, a field alternation or a
// builtin call.
type formatToken struct {
	kind  byte     // 'l' literal run, 'f' field alternation, 'b' builtin call
	lit   string   // kind 'l'
	body  string   // the braces' content, as written
	fn    string   // kind 'b'
	args  []string // kind 'b'
	names []string // kind 'f': the '|' arms; kind 'b': the operands its builtin reads
}

// parseFormat is the one reading of a format's tokens: a '(' outside a selector makes
// a token a call.
func parseFormat(format string) ([]formatToken, error) {
	var toks []formatToken
	var lit strings.Builder
	flush := func() {
		if lit.Len() > 0 {
			toks = append(toks, formatToken{kind: 'l', lit: lit.String()})
			lit.Reset()
		}
	}
	err := eachToken(format, func(t ftoken) error {
		if t.kind == 'l' {
			lit.WriteRune(t.char)
			return nil
		}
		flush()
		if indexOutside(t.body, '(') < 0 {
			toks = append(toks, formatToken{kind: 'f', body: t.body, names: splitOutside(t.body, '|')})
			return nil
		}
		name, args, ok := funcCall(t.body)
		if !ok {
			return fmt.Errorf("malformed function token {%s}", t.body)
		}
		toks = append(toks, formatToken{kind: 'b', body: t.body, fn: name, args: args, names: builtinOperands(name, args)})
		return nil
	})
	if err != nil {
		return nil, err
	}
	flush()
	return toks, nil
}

// builtin is a format-string function invoked as {name(args)}. It receives the
// session (its rng, and the {seq()} counters), the output emitted so far in the
// current expansion (for derivations such as a checksum over preceding digits), and
// the values of the operands it named (calc and the transforms name them). All must
// stay pure over (rng, emitted, args) so seeded output is reproducible; seq advances
// per-session counter state, which is itself deterministic. arity is the exact arg
// count, or -1 for variadic (then checkArgs does all the validation).
type builtin struct {
	arity int
	// prep parses validated args once, at compile time, into the closure expand calls.
	prep      func(args []string) callFn
	checkArgs func(fields map[string]node, args []string) error
	// operands names the fields the call reads, which expand renders for it; nil
	// for a builtin that reads none.
	operands func(args []string) []string
	// proveNumber proves what a call prints, token its body: the bounds of its number and the
	// datatype of its text, prints or one inside it; nil for a builtin whose text is no number.
	proveNumber func(token string, prints DataType, args []string) proven
	prints      DataType
}

// funcCall splits a "{token}" body shaped name(args) into its parts; ok is false
// for a plain field or alternation body. A '(' without a trailing ')' yields
// ok=false; checkFunc reports it as malformed at compile time.
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

func parseChecked(format string, fields map[string]node) ([]formatToken, error) {
	toks, err := parseFormat(format)
	if err != nil {
		return nil, err
	}
	return toks, checkTokens(toks, fields)
}

// checkTokens proves every parsed token names an existing field or a known function,
// so a typo'd or dangling reference is a New-time error.
func checkTokens(toks []formatToken, fields map[string]node) error {
	for _, t := range toks {
		if err := checkToken(t, fields); err != nil {
			return err
		}
	}
	return nil
}

func checkToken(t formatToken, fields map[string]node) error {
	switch t.kind {
	case 'b':
		return checkFunc(t, fields)
	case 'f':
		for _, name := range t.names {
			if isRef(name) {
				if _, _, err := refShape(name); err != nil {
					return fmt.Errorf("token {%s}: %w", t.body, err)
				}
				continue // its target is checked at New (see linkRefs)
			}
			if err := checkArm(name, fields, len(t.names) == 1); err != nil {
				return fmt.Errorf("token {%s}: %w", t.body, err)
			}
		}
		// Last, so an arm broken on its own terms is reported as that: a repeat is
		// the consequence of such a mistake, not the mistake itself.
		return checkNoRepeatedArm(t.body, t.names)
	}
	return nil
}

// checkArm validates one sibling name or path against a template's fields.
// wholeToken says the name is the token's entire body, so {/name} would render
// the same value and can be offered as the reference spelling.
func checkArm(name string, fields map[string]node, wholeToken bool) error {
	a := splitArm(name, nil)
	if err := checkSegments(a); err != nil {
		return err
	}
	head, ok := fields[a.key]
	if !ok {
		if a.key == "" {
			return fmt.Errorf("a name is never empty, so this token can name no field")
		}
		if isOption(a.key) {
			return fmt.Errorf("%q is an option and can never be a field", a.key)
		}
		if len(fields) == 0 {
			hint := ""
			if wholeToken && hintableRef(name) {
				hint = fmt.Sprintf(" — write {/%s} to reference the data", name)
			}
			return fmt.Errorf("no field %q; a token names a sibling field, and this template has none%s", a.key, hint)
		}
		return fmt.Errorf("no field %q", a.key)
	}
	if err := checkPath(head, a.tail, a.key); err != nil {
		return fmt.Errorf("field %q: %w", a.key, err)
	}
	return nil
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

// arm is one alternative of a {a|b} token or one operand, split into the key
// `template.head` resolves and the tail of a dotted path into it. A non-empty tail
// is what makes the arm a bound draw: its head is drawn once per expansion (see
// compileOps).
type arm struct {
	spelling string // as written, for messages
	key      string
	tail     []string
	levels   []string // the key at each level the tail passes through, the head's first
	path     string   // key and tail, the one path every way of writing this read shares
}

// splitArm splits one name into key and tail. refs maps a reference to what
// linkRefs bound it to; before linking, a reference is whole.
func splitArm(name string, refs map[string]refBinding) arm {
	if isRef(name) {
		b, bound := refs[name]
		if !bound || len(b.tail) == 0 {
			key := name
			if bound {
				key = b.key
			}
			return arm{spelling: name, key: key, path: key}
		}
		return pathArm(name, b.key, b.tail)
	}
	segs, err := splitPath(name)
	if err != nil || len(segs) == 1 {
		return arm{spelling: name, key: name, path: name}
	}
	return pathArm(name, segs[0], segs[1:])
}

func pathArm(name, key string, segs []string) arm {
	levels := []string{key}
	for i := 0; i < len(segs)-1; i++ {
		levels = append(levels, key+"."+strings.Join(segs[:i+1], "."))
	}
	return arm{spelling: name, key: key, tail: segs, levels: levels, path: key + "." + strings.Join(segs, ".")}
}

// checkSegments rejects an unfinished path: "{a.}" and "{a..b}" each have a
// segment naming nothing. A field really named "" would otherwise make them
// resolve, so a typo would read as a path that worked.
func checkSegments(a arm) error {
	if len(a.tail) == 0 {
		return nil
	}
	if a.key == "" {
		return fmt.Errorf("path has an empty segment")
	}
	for _, seg := range a.tail {
		if seg == "" {
			return fmt.Errorf("path has an empty segment")
		}
	}
	return nil
}

// callFn is a builtin bound to one call site: its args already parsed. It reads the
// output emitted so far in the current expansion (a derivation's payload) and the
// values of the operands it named, which expand read for it.
type callFn func(s *session, emitted string, operands []string) string

// op is one compiled unit of a format string: a literal run, a field alternation,
// or a builtin already bound to its args. compile builds these so render never
// re-scans the format.
type op struct {
	formatToken
	arms []arm // kind 'f': the '|' alternatives, split into key and path once
	call callFn
	// operands are the fields the builtin reads, in the order its operands func
	// fixed; expand reads them before the call. nil for a builtin that reads none.
	operands []arm
}

// formatOps is a compiled format: its ops, the size of its literal text (to size
// the render buffer), and the names drawn once per expansion. bound maps each level
// a path reads into to the first such path; held is every such level plus the
// fields an operand reads; holder maps each held name to the first reader holding
// it, for error messages. The maps are nil when the format holds nothing, so data
// that holds nothing carries no render-time cost.
type formatOps struct {
	ops       []op
	grow      int
	bound     map[string]string
	held      map[string]bool
	holder    map[string]string
	heldLocal bool // some held name is kept by the expansion itself rather than the render's hold
}

func (c *formatOps) holdName(a arm, label string) {
	if c.held == nil {
		c.held = map[string]bool{}
		c.holder = map[string]string{}
	}
	c.held[a.key] = true
	if !isRef(a.key) || len(a.tail) == 0 {
		c.heldLocal = true
	}
	if _, named := c.holder[a.key]; !named {
		c.holder[a.key] = label
	}
	if len(a.tail) > 0 {
		if c.bound == nil {
			c.bound = map[string]string{}
		}
		if _, named := c.bound[a.key]; !named {
			c.bound[a.key] = a.spelling
		}
	}
}

func (c *formatOps) function(tok formatToken, refs map[string]refBinding) {
	var operands []arm
	for _, operand := range tok.names {
		a := splitArm(operand, refs)
		c.holdName(a, fmt.Sprintf("%s operand %q", tok.fn, operand))
		operands = append(operands, a)
	}
	c.ops = append(c.ops, op{formatToken: tok, call: builtins[tok.fn].prep(tok.args), operands: operands})
}

func (c *formatOps) field(tok formatToken, refs map[string]refBinding) {
	arms := make([]arm, len(tok.names))
	for i, name := range tok.names {
		arms[i] = splitArm(name, refs)
		if len(arms[i].tail) > 0 {
			c.holdName(arms[i], "token {"+arms[i].spelling+"}")
		}
	}
	c.ops = append(c.ops, op{formatToken: tok, arms: arms})
}

// compileOps compiles a parsed format. Call checkTokens first: it is what proves
// every token valid.
func compileOps(toks []formatToken, refs map[string]refBinding) formatOps {
	var c formatOps
	for _, tok := range toks {
		switch tok.kind {
		case 'l':
			c.grow += len(tok.lit)
			c.ops = append(c.ops, op{formatToken: tok})
		case 'b':
			c.function(tok, refs)
		case 'f':
			c.field(tok, refs)
		}
	}
	return c
}
