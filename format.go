package fejkdata

import (
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/grammar"
)

// checkFunc validates a call at compile time: a known builtin, its args, for calc the operands held
// in fields, and for a where clause a builtin it can filter.
func checkFunc(tok grammar.Token, fields map[string]node) error {
	err := builtinfunc.Check(tok.Fn, tok.Args)
	if err == nil && tok.Fn == builtinfunc.CalcName {
		err = checkCalcFields(tok.Args, fields)
	}
	if err == nil && tok.Where != nil {
		if _, filters := builtinfunc.Runs(tok.Fn, tok.Args); !filters {
			err = fmt.Errorf("a where clause filters only %s", builtinfunc.FilteredCalls())
		}
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
		if err := checkBind(t, fields); err != nil {
			return nil, err
		}
		return checkReads(t, fields, true)
	case grammar.PathRead:
		return checkReads(t, fields, false)
	}
	return nil, nil
}

// checkReads proves that each spelling t reads (its arms, its builtin's operands and its clause's) is a reference or a path into a field. It returns the reads whose head no field holds.
func checkReads(t grammar.Token, fields map[string]node, operands bool) ([]unboundRead, error) {
	own, clause := tokenReads(t)
	wholeToken := !operands && len(own)+len(clause) == 1
	var unbound []unboundRead
	for _, spelling := range own {
		u, err := checkRead(t, spelling, fields, wholeToken, nil)
		if err != nil {
			return nil, err
		}
		unbound = append(unbound, u...)
	}
	for _, c := range clause {
		u, err := checkRead(t, c.spelling, fields, wholeToken, c.unbound)
		if err != nil {
			return nil, err
		}
		unbound = append(unbound, u...)
	}
	return unbound, nil
}

// checkRead proves spelling, one read of t, is a reference or a path into a field. Where no field holds
// its head, it returns the read as unbound; its error is unbound() where the unbound func is not nil.
func checkRead(t grammar.Token, spelling string, fields map[string]node, wholeToken bool, unbound func() error) ([]unboundRead, error) {
	if grammar.IsRef(spelling) {
		if _, _, err := grammar.RefShape(spelling); err != nil {
			return nil, fmt.Errorf("token {%s}: %w", t.Body, err)
		}
		return nil, nil // resolveRefs checks its target once the tree is assembled
	}
	missing, err := checkArm(spelling, fields, wholeToken)
	switch {
	case err == nil:
		return nil, nil
	case !missing:
		return nil, fmt.Errorf("token {%s}: %w", t.Body, err)
	case unbound != nil:
		err = unbound()
	}
	return []unboundRead{{head: splitRead(spelling, nil).head, body: t.Body, err: fmt.Errorf("token {%s}: %w", t.Body, err)}}, nil
}

// checkBind proves a {x as name} token is spelled once, binds a reference or a path into a
// field, and names a valid name that is no option.
func checkBind(t grammar.Token, fields map[string]node) error {
	ref, name := t.BoundRef, t.Bound
	if ref != strings.TrimSpace(ref) || name != strings.TrimSpace(name) {
		between := strings.TrimPrefix(strings.TrimSuffix(t.Body, name), ref)
		return fmt.Errorf("token {%s}: write {%s%s%s}", t.Body, strings.TrimSpace(ref), between, strings.TrimSpace(name))
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

// checkArm validates one read of a sibling field against a template's fields.
// wholeToken says the spelling is the token's entire body, so {/name} would render
// the same value and can be offered as the reference spelling. missing says no
// field holds the head, which a name may yet answer.
func checkArm(spelling string, fields map[string]node, wholeToken bool) (missing bool, err error) {
	a := splitRead(spelling, nil)
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
			if wholeToken && grammar.CheckPathIdentifiers(spelling) == nil {
				hint = fmt.Sprintf(" — write {/%s} to reference the data", spelling)
			}
			return true, fmt.Errorf("no field %q; a token names a sibling field, and this template has none%s", a.head, hint)
		}
		return true, fmt.Errorf("no field %q", a.head)
	}
	if err := provePath(field, a.tail, a.head, 0); err != nil {
		return false, fmt.Errorf("field %q: %w", a.head, err)
	}
	return false, nil
}

// tokenReads lists the spellings t reads: own, a read's arms or the operands a call's builtin reads,
// and clause, the spellings its where clause reads.
func tokenReads(t grammar.Token) (own []string, clause []clauseSpelling) {
	own = t.Arms
	if t.Kind == grammar.BuiltinCall {
		own = builtinfunc.Operands(t.Fn, t.Args)
	}
	return own, clauseSpellings(t.Where)
}

// formatOps is a compiled format: its ops, and the size of its literal text.
type formatOps struct {
	ops  []op
	grow int
}

// compileFormat compiles t's parsed format against its resolved references and its names'
// targets; a {ref as n} token compiles to no op, and a bind with a where or distinct clause to an
// op that expand skips, holding the arms its where clause reads. It leaves each op's call and draw to
// newOp. checkTokens proved every token valid.
func compileFormat(t *template, l *loadFinds) (formatOps, error) {
	var c formatOps
	for _, tok := range t.tokens {
		o := op{Token: tok}
		var err error
		switch tok.Kind {
		case grammar.LiteralRun:
			c.grow += len(tok.Lit)
		case grammar.BuiltinCall:
			o.operands, err = t.compileArms(builtinfunc.Operands(tok.Fn, tok.Args), l)
			if err == nil && tok.Where != nil {
				o.clause, err = t.compileClauseNames(tok.Where, l)
			}
		case grammar.PathRead:
			o.arms, err = t.compileArms(tok.Arms, l)
			if err == nil && tok.Where != nil {
				o.clause, err = t.compileClauseNames(tok.Where, l)
			}
		case grammar.NameBind:
			if tok.Where == nil && tok.Distinct == nil {
				continue
			}
			o.clause, err = t.compileClauseNames(tok.Where, l)
		}
		if err != nil {
			return c, fmt.Errorf("token {%s}: %w", tok.Body, err)
		}
		c.ops = append(c.ops, o)
	}
	return c, nil
}
