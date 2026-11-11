package load

import (
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

// checkFunc validates a call at compile time: a known builtin, its args, and for a where clause a
// builtin it can filter.
func checkFunc(tok grammar.Token) error {
	err := builtinfunc.Check(tok.Fn, tok.Args)
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

func parseChecked(format string, fields map[string]tree.Node) ([]grammar.Token, []tree.UnboundRead, error) {
	toks, err := grammar.ParseFormat(format)
	if err != nil {
		return nil, nil, err
	}
	unbound, err := checkTokens(toks, fields)
	return toks, unbound, err
}

// checkTokens proves every token's grammar, functions and field paths, and returns the reads
// whose head no field holds, for bindNames to look up among the names.
func checkTokens(toks []grammar.Token, fields map[string]tree.Node) ([]tree.UnboundRead, error) {
	var unbound []tree.UnboundRead
	for _, t := range toks {
		u, err := checkToken(t, fields)
		if err != nil {
			return nil, err
		}
		unbound = append(unbound, u...)
	}
	return unbound, nil
}

func checkToken(t grammar.Token, fields map[string]tree.Node) ([]tree.UnboundRead, error) {
	switch t.Kind {
	case grammar.BuiltinCall:
		if err := checkFunc(t); err != nil {
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

// checkReads proves that each spelling t reads (its alternatives, its builtin's operands and its clause's) is a reference or a path into a field. It returns the reads whose head no field holds.
func checkReads(t grammar.Token, fields map[string]tree.Node, operands bool) ([]tree.UnboundRead, error) {
	own, clause := tokenReads(t)
	wholeToken := !operands && len(own)+len(clause) == 1
	var unbound []tree.UnboundRead
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
func checkRead(t grammar.Token, spelling string, fields map[string]tree.Node, wholeToken bool, unbound func() error) ([]tree.UnboundRead, error) {
	if grammar.IsRef(spelling) {
		if _, _, err := grammar.RefShape(spelling); err != nil {
			return nil, fmt.Errorf("token {%s}: %w", t.Body, err)
		}
		return nil, nil // resolveRefs checks its target once the tree is assembled
	}
	missing, err := checkFieldRead(spelling, fields, wholeToken)
	switch {
	case err == nil:
		return nil, nil
	case !missing:
		return nil, fmt.Errorf("token {%s}: %w", t.Body, err)
	case unbound != nil:
		err = unbound()
	}
	return []tree.UnboundRead{tree.NewUnboundRead(splitRead(spelling, nil).head, t.Body, fmt.Errorf("token {%s}: %w", t.Body, err))}, nil
}

// checkBind proves a {x as name} token is spelled once, binds a reference or a path into a
// field, and names a valid name that is no option.
func checkBind(t grammar.Token, fields map[string]tree.Node) error {
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
func checkBound(ref string, fields map[string]tree.Node) error {
	if grammar.IsRef(ref) {
		_, _, err := grammar.RefShape(ref)
		return err
	}
	missing, err := checkFieldRead(ref, fields, false)
	if missing {
		return fmt.Errorf("%w; a binding names a field, or a reference, which starts with /, . or ..", err)
	}
	return err
}

// checkFieldRead validates one read of a sibling field against a template's fields.
// wholeToken says the spelling is the token's entire body, so {/name} would render
// the same value and can be offered as the reference spelling. missing says no
// field holds the head, which a name may yet answer.
func checkFieldRead(spelling string, fields map[string]tree.Node, wholeToken bool) (missing bool, err error) {
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

// tokenReads lists the spellings t reads: own, a token's alternatives or the operands a call's builtin reads,
// and clause, the spellings its where clause reads.
func tokenReads(t grammar.Token) (own []string, clause []clauseSpelling) {
	own = t.Arms
	if t.Kind == grammar.BuiltinCall {
		own = builtinfunc.Operands(t.Fn, t.Args)
	}
	return own, clauseSpellings(t.Where)
}

// compileFormat compiles t's parsed format against its resolved references and its names'
// targets; a {ref as n} token compiles to no op, and a bind with a where or distinct clause to an
// op that expand skips, holding the reads of its where clause. It leaves each op's call and draw to
// newOp. checkTokens proved every token valid.
func compileFormat(t *tree.Template, l *tree.LoadFinds) (tree.FormatOps, error) {
	var ops []tree.Op
	grow := 0
	for _, tok := range t.Tokens() {
		var alternatives, operands, clause []tree.Read
		var err error
		switch tok.Kind {
		case grammar.LiteralRun:
			grow += len(tok.Lit)
		case grammar.BuiltinCall:
			operands, err = compileReads(t, builtinfunc.Operands(tok.Fn, tok.Args), l)
			if err == nil && tok.Where != nil {
				clause, err = compileClauseNames(t, tok.Where, l)
			}
		case grammar.PathRead:
			alternatives, err = compileReads(t, tok.Arms, l)
			if err == nil && tok.Where != nil {
				clause, err = compileClauseNames(t, tok.Where, l)
			}
		case grammar.NameBind:
			if tok.Where == nil && tok.Distinct == nil {
				continue
			}
			clause, err = compileClauseNames(t, tok.Where, l)
		}
		if err != nil {
			return tree.FormatOps{}, fmt.Errorf("token {%s}: %w", tok.Body, err)
		}
		ops = append(ops, tree.NewOp(tok, alternatives, operands, clause))
	}
	return tree.NewFormatOps(ops, grow), nil
}
