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
		c := checkRead(t, spelling, fields, wholeToken)
		if c.refused != nil {
			return nil, c.refused
		}
		if c.noField != nil {
			unbound = append(unbound, unboundRead(t, spelling, c.noField))
		}
	}
	inClause, err := checkClauseReads(t, clause, fields, wholeToken)
	return append(unbound, inClause...), err
}

// checkClauseReads is checkReads for the spellings t's where clause reads, whose refusal names
// the clause's word.
func checkClauseReads(t grammar.Token, clause []clauseSpelling, fields map[string]tree.Node, wholeToken bool) ([]tree.UnboundRead, error) {
	var unbound []tree.UnboundRead
	for _, c := range clause {
		found := checkRead(t, c.spelling, fields, wholeToken)
		if found.refused != nil {
			return nil, found.refused
		}
		if found.noField != nil {
			unbound = append(unbound, unboundRead(t, c.spelling, c.unbound()))
		}
	}
	return unbound, nil
}

// readCheck is what checking one read against a template's fields found. The zero value is a
// read proved; at most one of its fields is set.
type readCheck struct {
	// refused is the error that refuses the read.
	refused error
	// noField says no field holds the read's head, and is the refusal that stands unless a name
	// bound around the template answers the read.
	noField error
}

// checkRead proves spelling, one read of t, is a reference or a path into a field.
func checkRead(t grammar.Token, spelling string, fields map[string]tree.Node, wholeToken bool) readCheck {
	if grammar.IsRef(spelling) {
		if _, _, err := grammar.RefShape(spelling); err != nil {
			return readCheck{refused: fmt.Errorf("token {%s}: %w", t.Body, err)}
		}
		return readCheck{} // resolveRefs checks its target once the tree is assembled
	}
	c := checkFieldRead(spelling, fields, wholeToken)
	if c.refused != nil {
		c.refused = fmt.Errorf("token {%s}: %w", t.Body, c.refused)
	}
	return c
}

// unboundRead is spelling, a read of t whose head no field holds, with the refusal that stands
// where no name answers it either.
func unboundRead(t grammar.Token, spelling string, refusal error) tree.UnboundRead {
	return tree.NewUnboundRead(splitRead(spelling, nil).head, t.Body, fmt.Errorf("token {%s}: %w", t.Body, refusal))
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
	c := checkFieldRead(ref, fields, false)
	if c.noField != nil {
		return fmt.Errorf("%w; a binding names a field, or a reference, which starts with /, . or ..", c.noField)
	}
	return c.refused
}

// checkFieldRead validates one read of a sibling field against a template's fields.
// wholeToken says the spelling is the token's entire body, so {/name} would render
// the same value and can be offered as the reference spelling.
func checkFieldRead(spelling string, fields map[string]tree.Node, wholeToken bool) readCheck {
	a := splitRead(spelling, nil)
	if err := checkSegments(a); err != nil {
		return readCheck{refused: err}
	}
	field, ok := fields[a.head]
	if !ok {
		return noFieldNamed(a.head, spelling, len(fields) == 0, wholeToken)
	}
	if err := provePath(field, a.tail, a.head, 0); err != nil {
		return readCheck{refused: fmt.Errorf("field %q: %w", a.head, err)}
	}
	return readCheck{}
}

func noFieldNamed(head, spelling string, noFields, wholeToken bool) readCheck {
	if head == "" {
		return readCheck{refused: fmt.Errorf("a name is never empty, so this token can name no field")}
	}
	if isOption(head) {
		return readCheck{refused: fmt.Errorf("%q is an option and can never be a field", head)}
	}
	if !noFields {
		return readCheck{noField: fmt.Errorf("no field %q", head)}
	}
	hint := ""
	if wholeToken && grammar.CheckPathIdentifiers(spelling) == nil {
		hint = fmt.Sprintf(" — write {/%s} to reference the data", spelling)
	}
	return readCheck{noField: fmt.Errorf("no field %q; a token names a sibling field, and this template has none%s", head, hint)}
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
		if tok.Kind == grammar.LiteralRun {
			grow += len(tok.Lit)
		}
		if tok.Kind == grammar.NameBind && tok.Where == nil && tok.Distinct == nil {
			continue
		}
		op, err := compileOp(t, tok, l)
		if err != nil {
			return tree.FormatOps{}, fmt.Errorf("token {%s}: %w", tok.Body, err)
		}
		ops = append(ops, op)
	}
	return tree.NewFormatOps(ops, grow), nil
}

// compileOp compiles the reads of tok's op: a call's operands, a path read's alternatives, and the
// names its where clause reads.
func compileOp(t *tree.Template, tok grammar.Token, l *tree.LoadFinds) (tree.Op, error) {
	var alternatives, operands, clause []tree.Read
	var err error
	switch tok.Kind {
	case grammar.BuiltinCall:
		operands, err = compileReads(t, builtinfunc.Operands(tok.Fn, tok.Args), l)
	case grammar.PathRead:
		alternatives, err = compileReads(t, tok.Arms, l)
	}
	if err == nil && tok.Where != nil {
		clause, err = compileClauseNames(t, tok.Where, l)
	}
	return tree.NewOp(tok, alternatives, operands, clause), err
}
