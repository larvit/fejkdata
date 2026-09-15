package fejkdata

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// calcNode is a parsed expression node. It evaluates over the operand values expand
// read, so the evaluator touches neither the rng nor the node tree.
type calcNode interface {
	eval(operands []string) float64
}

type calcNum float64 // a number literal
type calcVar string  // a sibling-field name, before indexVars places it
type calcIdx int     // an operand, by its position in the values expand read
type calcNeg struct{ x calcNode }
type calcBin struct { // a + - * / b
	op   byte
	l, r calcNode
}

func (n calcNum) eval([]string) float64 { return float64(n) }

func (n calcIdx) eval(operands []string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(operands[n]), 64)
	if err != nil {
		return math.NaN() // a non-numeric operand stays visible, never an error
	}
	return v
}

// eval on an unplaced name cannot happen: calcPrep runs indexVars over every
// expression it compiles, so only a calcIdx reaches a render. It panics rather
// than returning NaN, so a node kind indexVars forgets is a stack trace and not a
// silently wrong number.
func (n calcVar) eval([]string) float64 {
	panic(fmt.Sprintf("fejkdata: calc operand %q was never placed", string(n)))
}

func (n calcNeg) eval(operands []string) float64 { return -n.x.eval(operands) }

func (n calcBin) eval(operands []string) float64 {
	l, r := n.l.eval(operands), n.r.eval(operands)
	switch n.op {
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

// checkCalc validates a calc token at compile time: a parseable expression whose
// operands all name existing fields, and an optional non-negative integer dp. It
// is a builtin check (fields first), so calc dispatches through the registry like
// every other {name(args)} function.
func checkCalc(fields map[string]node, args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("calc takes an expression and an optional decimals count, got %d args", len(args))
	}
	expr, err := parseCalc(args[0])
	if err != nil {
		return fmt.Errorf("calc(%q): %w", args[0], err)
	}
	for _, name := range calcVars(expr) {
		operand, ok := fields[name]
		if !ok {
			return fmt.Errorf("calc(%q): no field %q", args[0], name)
		}
		if text, never := neverNumeric(operand); never {
			return fmt.Errorf("calc(%q): operand %q is never a number: it renders %q", args[0], name, text)
		}
	}
	if divisor, zero := constantZeroDivisor(expr, fields); zero {
		return fmt.Errorf("calc(%q) divides by %s, which is always zero", args[0], divisor)
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

// constantZeroDivisor finds a division whose right side is a constant zero: number
// literals and fixed operands folded, anything that varies left unknown.
func constantZeroDivisor(n calcNode, fields map[string]node) (string, bool) {
	switch n := n.(type) {
	case calcNeg:
		return constantZeroDivisor(n.x, fields)
	case calcBin:
		if n.op == '/' {
			if v, known := constantValue(n.r, fields); known && v == 0 {
				return calcText(n.r), true
			}
		}
		if d, zero := constantZeroDivisor(n.l, fields); zero {
			return d, true
		}
		return constantZeroDivisor(n.r, fields)
	}
	return "", false
}

// constantValue evaluates an expression whose every operand is fixed.
func constantValue(n calcNode, fields map[string]node) (float64, bool) {
	switch n := n.(type) {
	case calcNum:
		return float64(n), true
	case calcVar:
		t, ok := fields[string(n)].(*template)
		if !ok || !t.fixed || t.repeat > 1 {
			return 0, false
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(t.lit), 64)
		return v, err == nil
	case calcNeg:
		v, ok := constantValue(n.x, fields)
		return -v, ok
	case calcBin:
		l, lok := constantValue(n.l, fields)
		r, rok := constantValue(n.r, fields)
		if !lok || !rok {
			return 0, false
		}
		return calcBin{n.op, calcNum(l), calcNum(r)}.eval(nil), true
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
		return "(" + calcText(n.l) + " " + string(n.op) + " " + calcText(n.r) + ")"
	}
	return "?"
}

// neverNumeric reports a node no render of which is a number: a null, fixed text that
// does not parse, or a choice of only such items. text is one such render.
func neverNumeric(n node) (text string, never bool) {
	switch n := n.(type) {
	case *null:
		return "", true
	case *template:
		if !n.fixed || n.repeat > 1 {
			return "", false
		}
		if _, err := strconv.ParseFloat(strings.TrimSpace(n.lit), 64); err != nil {
			return n.lit, true
		}
	case *choice:
		for _, it := range n.items {
			t, itemNever := neverNumeric(it)
			if !itemNever {
				return "", false
			}
			text = t
		}
		return text, true
	}
	return "", false
}

// calcPrep parses the expression and decimals once, at compile time, and places each
// operand name at the position expand will read it into. checkCalc proved both args
// valid, so no step here can fail; dp -1 prints the minimal form.
func calcPrep(args []string) callFn {
	expr, err := parseCalc(args[0])
	if err != nil { // a nil AST would be a nil dereference per render, with no message
		panic(fmt.Sprintf("fejkdata: calc(%q) reached prep unparsed: %v", args[0], err))
	}
	at := make(map[string]int)
	for i, name := range calcVars(expr) {
		at[name] = i
	}
	placed := indexVars(expr, at)
	dp := calcDecimals(args)
	return func(_ *session, _ string, operands []string) string {
		return strconv.FormatFloat(placed.eval(operands), 'f', dp, 64)
	}
}

// indexVars replaces each operand name with its position in the values expand reads.
// Both sides take that order from calcVars, so they cannot drift.
func indexVars(n calcNode, at map[string]int) calcNode {
	switch n := n.(type) {
	case calcVar:
		i, placed := at[string(n)]
		if !placed { // calcVars named every operand, so a miss means the two disagree
			panic(fmt.Sprintf("fejkdata: calc operand %q is not among the names read for it", string(n)))
		}
		return calcIdx(i)
	case calcNeg:
		return calcNeg{indexVars(n.x, at)}
	case calcBin:
		return calcBin{n.op, indexVars(n.l, at), indexVars(n.r, at)}
	}
	return n
}

// calcOperands lists the sibling-field names a calc's args read. checkCalc reports
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

// calcVars lists the distinct field names an expression reads, in the order it first
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

// factor is a table-shaped scanner, one case per token kind, kept whole on purpose.
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

// ident reads a field name: a letter or '_', then letters, digits or '_'. A '-'
// is always the minus operator, so a hyphenated field name can't be an operand.
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

// calcDecimals is a calc's decimals count, or -1 for the shortest form.
func calcDecimals(args []string) int {
	if len(args) == 2 {
		return atoi(args[1])
	}
	return -1
}

// calcLimit is the largest magnitude a proof accepts as finite, far enough below
// math.MaxFloat64 that rounding in the bounds cannot hide an overflow.
const calcLimit = 1e300

// maxOperandLen is the longest operand text a proof bounds by its length, so that
// bound, 10^maxOperandLen, stays within calcLimit.
const maxOperandLen = 300

// calcBound is what a proof knows of every value a calc can take: it lies in [lo, hi],
// is at least nonZero from zero unless nonZero is 0, and is whole when integral.
type calcBound struct {
	lo, hi, nonZero float64
	integral        bool
}

func magnitude(b calcBound) float64 { return math.Max(math.Abs(b.lo), math.Abs(b.hi)) }

// doubt is why a proof could not show a calc finite, and the render that shows it.
type doubt struct{ render, why string }

type bounded struct {
	b calcBound
	d *doubt
}

// calcProof bounds a typed column's calcs from their operands' renders, to show each
// prints a number rather than NaN or Inf.
type calcProof struct {
	decimal  *textLanguage
	operands map[node]bounded
	lengths  map[node]int
}

func newCalcProof() *calcProof {
	p := &calcProof{operands: map[node]bounded{}, lengths: map[node]int{}}
	p.decimal = newTextLanguage(decimalGrammar, p)
	return p
}

// call bounds one calc token of t.
func (p *calcProof) call(t *template, args []string) (calcBound, *doubt) {
	expr, err := parseCalc(args[0])
	if err != nil {
		panic(fmt.Sprintf("fejkdata: calc(%q) reached a proof unparsed: %v", args[0], err))
	}
	b, d := p.expr(expr, t.fields)
	if d != nil {
		return b, &doubt{d.render, fmt.Sprintf("{calc(%s)}: %s", strings.Join(args, ", "), d.why)}
	}
	return b, nil
}

func (p *calcProof) expr(n calcNode, fields map[string]node) (calcBound, *doubt) {
	switch n := n.(type) {
	case calcNum:
		v := float64(n)
		return calcBound{v, v, v, v == math.Trunc(v)}, nil
	case calcVar:
		return p.operand(string(n), fields[string(n)])
	case calcNeg:
		b, d := p.expr(n.x, fields)
		return calcBound{-b.hi, -b.lo, b.nonZero, b.integral}, d
	case calcBin:
		l, d := p.expr(n.l, fields)
		if d != nil {
			return l, d
		}
		r, d := p.expr(n.r, fields)
		if d != nil {
			return r, d
		}
		return combine(n, l, r)
	}
	panic(fmt.Sprintf("fejkdata: calc node %T has no bound", n))
}

// combine bounds one operation from the bounds of its sides.
func combine(n calcBin, l, r calcBound) (calcBound, *doubt) {
	b := calcBound{integral: l.integral && r.integral}
	switch n.op {
	case '+':
		b.lo, b.hi = l.lo+r.lo, l.hi+r.hi
	case '-':
		b.lo, b.hi = l.lo-r.hi, l.hi-r.lo
	case '*':
		b.lo = min(l.lo*r.lo, l.lo*r.hi, l.hi*r.lo, l.hi*r.hi)
		b.hi = max(l.lo*r.lo, l.lo*r.hi, l.hi*r.lo, l.hi*r.hi)
		b.nonZero = l.nonZero * r.nonZero
	default:
		if r.nonZero == 0 {
			return b, &doubt{"+Inf", fmt.Sprintf("divides by %s, which can be zero", calcText(n.r))}
		}
		m := magnitude(l) / r.nonZero
		b = calcBound{lo: -m, hi: m, nonZero: l.nonZero / magnitude(r)}
	}
	if b.lo > 0 || b.hi < 0 {
		b.nonZero = math.Max(b.nonZero, math.Min(math.Abs(b.lo), math.Abs(b.hi)))
	}
	if !(magnitude(b) <= calcLimit) {
		return b, &doubt{"+Inf", calcText(n) + " can overflow"}
	}
	return b, nil
}

// operand bounds a calc operand, once per node.
func (p *calcProof) operand(name string, n node) (calcBound, *doubt) {
	if seen, done := p.operands[n]; done {
		return seen.b, seen.d
	}
	b, d := p.measure(name, n)
	p.operands[n] = bounded{b, d}
	return b, d
}

// measure bounds an operand through the calc it renders when that is all it renders,
// and otherwise from its text: a plain decimal of at most maxOperandLen bytes.
func (p *calcProof) measure(name string, n node) (calcBound, *doubt) {
	if t, ok := n.(*template); ok {
		if args, isCalc := soleCalc(t); isCalc {
			b, d := p.call(t, args)
			return rounded(b, calcDecimals(args)), d
		}
	}
	text := p.decimal.node(n, nil)
	if w, escapes := text.escape(decimalAccept); escapes {
		why := fmt.Sprintf("operand %q can render %s, which is not a plain decimal", name, w)
		if w.why != "" {
			why += ": " + w.why
		}
		return calcBound{}, &doubt{"NaN", why}
	}
	size := p.length(n)
	if size > maxOperandLen {
		return calcBound{}, &doubt{"NaN", fmt.Sprintf("operand %q can render more than %d bytes, too many to bound", name, maxOperandLen)}
	}
	ends, m := text.to[1], math.Pow(10, float64(size))
	b := calcBound{hi: m, nonZero: 1 / m, integral: ends&decimalFractional == 0}
	if ends&decimalNegative != 0 {
		b.lo = -m
	}
	if ends&decimalZero != 0 {
		b.nonZero = 0
	}
	return b, nil
}

// soleCalc reports a template that renders one calc and nothing else, with its args.
func soleCalc(t *template) ([]string, bool) {
	if t.repeat != 1 || len(t.ops) != 1 || t.ops[0].kind != 'b' {
		return nil, false
	}
	name, args, _ := funcCall(t.format[1 : len(t.format)-1])
	return args, name == "calc"
}

// rounded is b once printed to dp decimals, which moves a value by up to half a unit.
func rounded(b calcBound, dp int) calcBound {
	if dp < 0 {
		return b
	}
	half := math.Pow(10, -float64(dp)) / 2
	return calcBound{b.lo - half, b.hi + half, math.Max(0, b.nonZero-half), b.integral || dp == 0}
}

// length is the most bytes a render of n can take, anything past maxOperandLen
// reported as maxOperandLen+1.
func (p *calcProof) length(n node) int {
	if size, done := p.lengths[n]; done {
		return size
	}
	size := 0
	switch n := n.(type) {
	case *choice:
		for _, it := range n.items {
			size = max(size, p.length(it))
		}
	case *template:
		size = p.formatLength(n)*n.repeat + len(n.separator)*(n.repeat-1)
	}
	size = min(size, maxOperandLen+1)
	p.lengths[n] = size
	return size
}

func (p *calcProof) formatLength(t *template) int {
	size := 0
	_ = eachToken(t.format, func(tok ftoken) error {
		if tok.kind == 'l' {
			size += utf8.RuneLen(tok.r)
		} else {
			size += p.tokenLength(t, tok.body)
		}
		size = min(size, maxOperandLen+1)
		return nil
	})
	return size
}

// tokenLength is the most bytes one token can print. A transform never lengthens a
// render that reads as a decimal: it maps each non-ASCII rune, two bytes or more, to at
// most two ASCII letters.
func (p *calcProof) tokenLength(t *template, body string) int {
	name, args, isFunc := funcCall(body)
	var arms []arm
	switch _, isTransform := transforms[name]; {
	case !isFunc:
		arms = splitArms(body, t.refs)
	case isTransform:
		leaf, _, _ := unwrapTransform(args[0])
		arms = []arm{splitArm(leaf, t.refs)}
	case name == "calc":
		b, d := p.call(t, args)
		if d != nil {
			return len(d.render)
		}
		return shapeLength(printedFloat(b.lo, b.hi, calcDecimals(args), b.integral))
	default:
		return shapeLength(builtins[name].emits(args))
	}
	size := 0
	for _, a := range arms {
		for _, leaf := range pathLeaves(t.fields[a.key], a.tail) {
			size = max(size, p.length(leaf))
		}
	}
	return size
}

// shapeLength is the most bytes a shape can emit, anything past maxOperandLen reported
// as maxOperandLen+1.
func shapeLength(s textShape) int {
	longest := 0
	for _, alt := range s {
		size := 0
		for _, run := range alt {
			if run.max < 0 {
				return maxOperandLen + 1
			}
			size += run.max
		}
		longest = max(longest, size)
	}
	return min(longest, maxOperandLen+1)
}
