package fejkdata

import (
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/grammar"
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
	case grammar.PathRead:
		return checkReads(t, fields, false)
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
	if err := grammar.CheckIdentifier(name); err != nil {
		return fmt.Errorf("token {%s}: name %w", t.Body, err)
	}
	return nil
}

// checkBound proves what a binding binds is a reference or a path into a field of its binder.
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
	if err := provePath(field, a.tail, a.head); err != nil {
		return false, fmt.Errorf("field %q: %w", a.head, err)
	}
	return false, nil
}

// hintableRef reports whether {/name} is a reference the grammar accepts, so the
// hint never names a spelling that fails too.
func hintableRef(name string) bool {
	return grammar.CheckPathIdentifiers(name) == nil
}

// tokenReads lists the names t reads: a read's arms, or the operands a call's builtin reads.
func tokenReads(t grammar.Token) []string {
	if t.Kind == grammar.BuiltinCall {
		return builtinfunc.Operands(t.Fn, t.Args)
	}
	return t.Arms
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
		case grammar.PathRead:
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
