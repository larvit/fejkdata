package fejkdata

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

// calcNode is a parsed expression node. It evaluates over the operand values expand
// read, so the evaluator touches neither the rng nor the node tree.
type calcNode interface {
	eval(operands []string) float64
}

type calcNum float64 // a number literal
type calcVar string  // an operand, a sibling field or a name, before indexVars places it
type calcIdx int     // an operand, by its position in the values expand read
type calcNeg struct{ x calcNode }
type calcBin struct { // a + - * / b
	operator byte
	l, r     calcNode
}

func (n calcNum) eval([]string) float64 { return float64(n) }

func (n calcIdx) eval(operands []string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(operands[n]), 64)
	if err != nil {
		return math.NaN() // a non-numeric operand stays visible, never an error
	}
	return v
}

func (n calcVar) eval([]string) float64 {
	panic(internalError("calc operand %q was never placed", string(n)))
}

func (n calcNeg) eval(operands []string) float64 { return -n.x.eval(operands) }

func (n calcBin) eval(operands []string) float64 {
	l, r := n.l.eval(operands), n.r.eval(operands)
	switch n.operator {
	case '+':
		return l + r
	case '-':
		return l - r
	case '*':
		return l * r
	default: // '/'
		return l / r
	}
}

// checkCalc validates a calc token at compile time: a parseable expression, operands
// that can be numbers, and an optional non-negative integer dp. An operand no field
// holds is left for a name to answer, and checkCalcNames checks it once names link.
func checkCalc(fields map[string]node, args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("calc takes an expression and an optional decimals count, got %d args", len(args))
	}
	expr, err := parseCalc(args[0])
	if err != nil {
		return fmt.Errorf("calc(%q): %w", args[0], err)
	}
	if err := checkOperands(args[0], expr, func(name string) []node {
		if n, ok := fields[name]; ok {
			return []node{n}
		}
		return nil
	}); err != nil {
		return err
	}
	if len(args) == 2 {
		dp, err := plainInt(args[1])
		if err != nil {
			return fmt.Errorf("calc decimals %w", err)
		}
		if dp < 0 || dp > maxDecimals {
			return fmt.Errorf("calc decimals %d must be in 0..%d", dp, maxDecimals)
		}
	}
	return nil
}

// checkCalcNames holds each calc of t reading a name to the checks checkCalc makes of a field.
func checkCalcNames(path string, t *template) error {
	for _, o := range t.compiled.ops {
		if o.fn != "calc" || !slices.ContainsFunc(o.operands, func(a arm) bool { return a.kind == namedRead }) {
			continue
		}
		if err := checkOperands(o.args[0], parsedCalc(o.args[0]), operandNodes(o)); err != nil {
			return fmt.Errorf("%s: token {%s}: %w", t.site.label(path), o.body, err)
		}
	}
	return nil
}

// operandNodes lists every node each operand of o may render, once linkNames compiled the names.
func operandNodes(o op) func(name string) []node {
	return func(name string) []node {
		a := o.operands[slices.IndexFunc(o.operands, func(a arm) bool { return a.head == name })]
		if len(a.leaves) == 0 {
			panic(internalError("calc operand %q was read before it was compiled", name))
		}
		return a.leaves
	}
}

// checkOperands refuses an operand that is never a number and a division by a constant zero.
// operand lists every node an operand may render, nil while that is unknown.
func checkOperands(text string, expr calcNode, operand func(name string) []node) error {
	for _, name := range calcVars(expr) {
		if rendered, never := noneNumeric(operand(name)); never {
			return fmt.Errorf("calc(%q): operand %q is never a number: it renders %q", text, name, rendered)
		}
	}
	if divisor, zero := constantZeroDivisor(expr, operand); zero {
		return fmt.Errorf("calc(%q) divides by %s, which is always zero", text, divisor)
	}
	return nil
}

// constantZeroDivisor finds a division whose right side is a constant zero: number
// literals and fixed operands folded, anything that varies left unknown.
func constantZeroDivisor(n calcNode, operand func(string) []node) (string, bool) {
	switch n := n.(type) {
	case calcNeg:
		return constantZeroDivisor(n.x, operand)
	case calcBin:
		if n.operator == '/' {
			if v, known := constantValue(n.r, operand); known && v == 0 {
				return calcText(n.r), true
			}
		}
		if d, zero := constantZeroDivisor(n.l, operand); zero {
			return d, true
		}
		return constantZeroDivisor(n.r, operand)
	}
	return "", false
}

// constantValue evaluates an expression whose every operand is fixed.
func constantValue(n calcNode, operand func(string) []node) (float64, bool) {
	switch n := n.(type) {
	case calcNum:
		return float64(n), true
	case calcVar:
		nodes := operand(string(n))
		if len(nodes) != 1 {
			return 0, false
		}
		t, ok := nodes[0].(*template)
		if !ok || t.repeat > 1 {
			return 0, false
		}
		lit, fixed := t.fixedText()
		if !fixed {
			return 0, false
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(lit), 64)
		return v, err == nil
	case calcNeg:
		v, ok := constantValue(n.x, operand)
		return -v, ok
	case calcBin:
		l, lok := constantValue(n.l, operand)
		r, rok := constantValue(n.r, operand)
		if !lok || !rok {
			return 0, false
		}
		return calcBin{n.operator, calcNum(l), calcNum(r)}.eval(nil), true
	}
	return 0, false
}

// calcText spells an expression node the way an author would read it.
func calcText(n calcNode) string {
	switch n := n.(type) {
	case calcNum:
		return strconv.FormatFloat(float64(n), 'f', -1, 64)
	case calcVar:
		return string(n)
	case calcNeg:
		return "-" + calcText(n.x)
	case calcBin:
		return "(" + calcText(n.l) + " " + string(n.operator) + " " + calcText(n.r) + ")"
	}
	return "?"
}

// neverNumeric reports a node no render of which is a number: a null, fixed text that
// does not parse, or a choice of only such items. text is one such render.
func neverNumeric(n node) (text string, never bool) {
	switch n := n.(type) {
	case *nullItem:
		return "", true
	case *template:
		lit, fixed := n.fixedText()
		if !fixed || n.repeat > 1 {
			return "", false
		}
		if _, err := strconv.ParseFloat(strings.TrimSpace(lit), 64); err != nil {
			return lit, true
		}
	case *choice:
		return noneNumeric(n.items)
	}
	return "", false
}

// noneNumeric reports nodes no render of which is a number, and none where there are none.
func noneNumeric(nodes []node) (text string, never bool) {
	for _, n := range nodes {
		t, nodeNever := neverNumeric(n)
		if !nodeNever {
			return "", false
		}
		text = t
	}
	return text, len(nodes) > 0
}

// calcPrep parses the expression and decimals once, at compile time, and places each
// operand name at the position expand will read it into. checkCalc proved both args
// valid, so no step here can fail.
func calcPrep(args []string) callFn {
	expr := parsedCalc(args[0])
	at := make(map[string]int)
	for i, name := range calcVars(expr) {
		at[name] = i
	}
	placed := indexVars(expr, at)
	dp := calcDecimals(args)
	return func(_ *generatorState, _ string, operands []string) string {
		return formatFloat(placed.eval(operands), dp)
	}
}

// parsedCalc parses an expression checkCalc accepted. A nil AST would be a nil dereference
// per render, with no message.
func parsedCalc(expr string) calcNode {
	n, err := parseCalc(expr)
	if err != nil {
		panic(internalError("calc(%q) passed its check unparsed: %v", expr, err))
	}
	return n
}

// calcDecimals is a calc's decimals count, or shortestDecimals where it names none.
func calcDecimals(args []string) int {
	if len(args) == 2 {
		return atoi(args[1])
	}
	return shortestDecimals
}

// indexVars replaces each operand name with its position in the values expand reads.
// Both sides take that order from calcVars, so they cannot drift.
func indexVars(n calcNode, at map[string]int) calcNode {
	switch n := n.(type) {
	case calcVar:
		i, placed := at[string(n)]
		if !placed { // calcVars named every operand, so a miss means the two disagree
			panic(internalError("calc operand %q is not among the names read for it", string(n)))
		}
		return calcIdx(i)
	case calcNeg:
		return calcNeg{indexVars(n.x, at)}
	case calcBin:
		return calcBin{n.operator, indexVars(n.l, at), indexVars(n.r, at)}
	}
	return n
}

// calcOperands lists the operands a calc's args read, fields or names. checkCalc reports
// an expression that does not parse, so one that does not simply names nothing.
func calcOperands(args []string) []string {
	if len(args) == 0 {
		return nil
	}
	expr, err := parseCalc(args[0])
	if err != nil {
		return nil
	}
	return calcVars(expr)
}

// calcVars lists the distinct operands an expression reads, in the order it first
// names each. That order is the contract between expand, which reads the operands
// into a slice, and indexVars, which places each name at its position in it.
func calcVars(n calcNode) []string {
	var out []string
	seen := map[string]bool{}
	var walk func(calcNode)
	walk = func(n calcNode) {
		switch n := n.(type) {
		case calcVar:
			if name := string(n); !seen[name] {
				seen[name] = true
				out = append(out, name)
			}
		case calcNeg:
			walk(n.x)
		case calcBin:
			walk(n.l)
			walk(n.r)
		}
	}
	walk(n)
	return out
}

// calcParser is a recursive-descent parser over the expression runes, threading
// expr -> term -> factor for the standard * / before + - precedence.
type calcParser struct {
	rs  []rune
	pos int
}

// parseCalc parses a whole expression, requiring it to consume all input.
func parseCalc(expr string) (calcNode, error) {
	p := &calcParser{rs: []rune(expr)}
	if p.space(); p.pos >= len(p.rs) {
		return nil, fmt.Errorf("empty expression")
	}
	n, err := p.expr()
	if err != nil {
		return nil, err
	}
	if p.space(); p.pos != len(p.rs) {
		return nil, fmt.Errorf("unexpected %q", string(p.rs[p.pos:]))
	}
	return n, nil
}

func (p *calcParser) space() {
	for p.pos < len(p.rs) && unicode.IsSpace(p.rs[p.pos]) {
		p.pos++
	}
}

func (p *calcParser) expr() (calcNode, error) { return p.binary(p.term, '+', '-') }
func (p *calcParser) term() (calcNode, error) { return p.binary(p.factor, '*', '/') }

// binary parses a left-associative run of next() operands joined by the given
// operators, the one shape expr and term share.
func (p *calcParser) binary(next func() (calcNode, error), ops ...byte) (calcNode, error) {
	n, err := next()
	if err != nil {
		return nil, err
	}
	for {
		p.space()
		if p.pos >= len(p.rs) || !contains(ops, byte(p.rs[p.pos])) {
			return n, nil
		}
		op := byte(p.rs[p.pos])
		p.pos++
		r, err := next()
		if err != nil {
			return nil, err
		}
		n = calcBin{op, n, r}
	}
}

func (p *calcParser) factor() (calcNode, error) {
	p.space()
	if p.pos >= len(p.rs) {
		return nil, fmt.Errorf("unexpected end of expression")
	}
	switch c := p.rs[p.pos]; {
	case c == '-':
		p.pos++
		x, err := p.factor()
		if err != nil {
			return nil, err
		}
		return calcNeg{x}, nil
	case c == '(':
		p.pos++
		n, err := p.expr()
		if err != nil {
			return nil, err
		}
		if p.space(); p.pos >= len(p.rs) || p.rs[p.pos] != ')' {
			return nil, fmt.Errorf("missing ')'")
		}
		p.pos++
		return n, nil
	case c == '.' || c >= '0' && c <= '9':
		return p.number()
	case c == '_' || unicode.IsLetter(c):
		return p.ident()
	default:
		return nil, fmt.Errorf("unexpected %q", string(c))
	}
}

func (p *calcParser) number() (calcNode, error) {
	start, dot := p.pos, false
	for p.pos < len(p.rs) {
		if c := p.rs[p.pos]; c >= '0' && c <= '9' {
			p.pos++
		} else if c == '.' && !dot {
			dot, p.pos = true, p.pos+1
		} else {
			break
		}
	}
	v, err := strconv.ParseFloat(string(p.rs[start:p.pos]), 64)
	if err != nil {
		return nil, fmt.Errorf("bad number %q", string(p.rs[start:p.pos]))
	}
	return calcNum(v), nil
}

// ident reads a field or name: a letter or '_', then letters, digits or '_'. A '-'
// is always the minus operator, so a hyphenated one can't be an operand.
func (p *calcParser) ident() (calcNode, error) {
	start := p.pos
	for p.pos < len(p.rs) {
		if c := p.rs[p.pos]; c == '_' || unicode.IsLetter(c) || unicode.IsDigit(c) {
			p.pos++
		} else {
			break
		}
	}
	return calcVar(string(p.rs[start:p.pos])), nil
}

func contains(bs []byte, b byte) bool {
	for _, x := range bs {
		if x == b {
			return true
		}
	}
	return false
}
