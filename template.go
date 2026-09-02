package fejkdata

import (
	"fmt"
	"strings"
)

// ftoken is one unit of a scanned format string: a literal rune or the body of a
// {…} token.
type ftoken struct {
	kind byte // 'l' literal rune, 'b' brace body
	r    rune
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
				t.kind, t.r = 'l', '{'
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
			t.kind, t.body = 'b', string(rs[i+1:end])
			i = end
		case '}':
			if i+1 < len(rs) && rs[i+1] == '}' {
				t.kind, t.r = 'l', '}'
				i++
				break
			}
			return fmt.Errorf("lone '}' in %q; a literal brace is written }}", format)
		default:
			t.kind, t.r = 'l', c
		}
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}

// builtin is a format-string function invoked as {name(args)}. It receives the
// session (its rng, and the {seq()} counters), the output emitted so far in the
// current expansion (for derivations such as a checksum over preceding digits), and
// the values of the operands it named (only calc names any). All must stay pure
// over (rng, emitted, args) so seeded output is reproducible; seq advances
// per-session counter state, which is itself deterministic. arity is the exact arg
// count, or -1 for variadic (then check does all the validation).
type builtin struct {
	arity int
	// prep parses validated args once, at compile time, into the closure expand calls.
	prep  func(args []string) callFn
	check func(fields map[string]node, args []string) error
	// operands names the fields the call reads, which expand renders for it; nil
	// for a builtin that reads none.
	operands func(args []string) []string
}

// funcCall splits a "{token}" body shaped name(args) into its parts; ok is false
// for a plain field or alternation body. A '(' without a trailing ')' yields
// ok=false; checkFunc reports it as malformed at compile time.
func funcCall(body string) (name string, args []string, ok bool) {
	lp := strings.IndexByte(body, '(')
	if lp < 0 || !strings.HasSuffix(body, ")") {
		return "", nil, false
	}
	return body[:lp], splitArgs(body[lp+1 : len(body)-1]), true
}

// splitArgs parses a function arg list: comma-separated, trimmed; empty -> none.
func splitArgs(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	args := strings.Split(s, ",")
	for i := range args {
		args[i] = strings.TrimSpace(args[i])
	}
	return args
}

// checkFunc validates a function token at compile time: well-formed, naming a
// known builtin, with the arg count that builtin takes and args its check accepts.
// fields is passed through for the one builtin (calc) that validates against them.
func checkFunc(body string, fields map[string]node) error {
	name, args, ok := funcCall(body)
	if !ok {
		return fmt.Errorf("malformed function token {%s}", body)
	}
	b, known := builtins[name]
	if !known {
		return fmt.Errorf("token {%s}: unknown function %q", body, name)
	}
	if b.arity >= 0 && len(args) != b.arity {
		return fmt.Errorf("token {%s}: %s takes %d args, got %d", body, name, b.arity, len(args))
	}
	if b.check != nil {
		if err := b.check(fields, args); err != nil {
			return fmt.Errorf("token {%s}: %w", body, err)
		}
	}
	return nil
}

// checkTokens validates a format string the way expand scans it, so every
// "{token}" is balanced and names an existing field (or a known function). This
// makes a typo'd or dangling reference a New-time error, never a random
// render-time one.
func checkTokens(format string, fields map[string]node) error {
	return eachToken(format, func(t ftoken) error {
		if t.kind != 'b' {
			return nil
		}
		if strings.IndexByte(t.body, '(') >= 0 { // a function token, not a field
			return checkFunc(t.body, fields)
		}
		names := strings.Split(t.body, "|")
		for _, name := range names {
			if isRef(name) {
				if _, _, err := refShape(name); err != nil {
					return fmt.Errorf("token {%s}: %w", t.body, err)
				}
				continue // its target is checked at New (see linkRefs)
			}
			if err := checkArm(name, fields); err != nil {
				return fmt.Errorf("token {%s}: %w", t.body, err)
			}
		}
		// Last, so an arm broken on its own terms is reported as that: a repeat is
		// the consequence of such a mistake, not the mistake itself.
		return checkNoRepeatedArm(t.body, names)
	})
}

// checkArm validates one sibling name or path against a template's fields.
func checkArm(name string, fields map[string]node) error {
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
		return fmt.Errorf("no field %q", a.key)
	}
	if err := checkPath(head, a.tail, a.key); err != nil {
		return fmt.Errorf("field %q: %w", a.key, err)
	}
	return nil
}

// tokenOperands lists the fields one {token} body reads as operands, empty for a
// field token or a builtin that reads none.
func tokenOperands(body string) []string {
	name, args, ok := funcCall(body)
	if !ok {
		return nil
	}
	b, known := builtins[name]
	if !known || b.operands == nil {
		return nil
	}
	return b.operands(args)
}

// operandTokens lists every operand the builtins in a format read.
func operandTokens(format string) []string {
	var names []string
	_ = eachToken(format, func(t ftoken) error {
		if t.kind == 'b' {
			names = append(names, tokenOperands(t.body)...)
		}
		return nil
	})
	return names
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

// fieldTokens returns the field and reference names a format renders via {name}
// or {a|..b} tokens (function tokens, which carry no field edges, are excluded).
// These are exactly the child nodes expand recurses into, through readField.
func fieldTokens(format string) []string {
	var names []string
	_ = eachToken(format, func(t ftoken) error {
		if t.kind == 'b' && strings.IndexByte(t.body, '(') < 0 {
			names = append(names, strings.Split(t.body, "|")...)
		}
		return nil
	})
	return names
}

// checkSegments rejects an unfinished path: "{a.}", "{.b}" and "{a..b}" each have
// a segment naming nothing. A field really named "" would otherwise make them
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
	kind byte   // 'l' literal run, 'f' field alternation, 'b' builtin
	lit  string // kind 'l'
	arms []arm  // kind 'f': the '|' alternatives, split into key and path once
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
	ops    []op
	grow   int
	bound  map[string]string
	held   map[string]bool
	holder map[string]string
}

func (c *formatOps) hold(a arm, label string) {
	if c.held == nil {
		c.held = map[string]bool{}
		c.holder = map[string]string{}
	}
	c.held[a.key] = true
	if _, named := c.holder[a.key]; !named {
		c.holder[a.key] = label
	}
	if len(a.tail) > 0 {
		if c.bound == nil {
			c.bound = map[string]string{}
		}
		if _, named := c.bound[a.key]; !named {
			c.bound[a.key] = a.name
		}
	}
}

func (c *formatOps) function(body string, refs map[string]refBinding) {
	name, args, _ := funcCall(body)
	var operands []arm
	for _, operand := range tokenOperands(body) {
		a := splitArm(operand, refs)
		c.hold(a, fmt.Sprintf("%s operand %q", name, operand))
		operands = append(operands, a)
	}
	c.ops = append(c.ops, op{kind: 'b', call: builtins[name].prep(args), operands: operands})
}

func (c *formatOps) field(body string, refs map[string]refBinding) {
	arms := splitArms(body, refs)
	for _, a := range arms {
		if len(a.tail) > 0 {
			c.hold(a, "token {"+a.name+"}")
		}
	}
	c.ops = append(c.ops, op{kind: 'f', arms: arms})
}

// compileOps compiles a format string. Call checkTokens first: it is what proves
// the scan and every token are valid.
func compileOps(format string, refs map[string]refBinding) formatOps {
	var c formatOps
	var lit strings.Builder
	flush := func() {
		if lit.Len() > 0 {
			c.grow += lit.Len()
			c.ops = append(c.ops, op{kind: 'l', lit: lit.String()})
			lit.Reset()
		}
	}
	_ = eachToken(format, func(t ftoken) error {
		switch t.kind {
		case 'l':
			lit.WriteRune(t.r)
		case 'b':
			flush()
			if _, _, isFunc := funcCall(t.body); isFunc {
				c.function(t.body, refs)
			} else {
				c.field(t.body, refs)
			}
		}
		return nil
	})
	flush()
	return c
}
