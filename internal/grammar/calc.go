package grammar

import (
	"fmt"
	"slices"
	"strconv"
	"unicode"
)

// CalcNode is a parsed calc expression node.
type CalcNode interface{ calcNode() }

type CalcNum float64 // a number literal

// CalcVar is an operand, a sibling field or a name. At is its position among the
// distinct operands in the order the expression first names each, which CalcVars lists.
type CalcVar struct {
	Name string
	At   int
}

type CalcNeg struct{ X CalcNode }

type CalcBin struct { // L + - * / R
	Operator byte
	L, R     CalcNode
}

func (CalcNum) calcNode() {}
func (CalcVar) calcNode() {}
func (CalcNeg) calcNode() {}
func (CalcBin) calcNode() {}

// CalcVars lists the distinct operands an expression reads, each at its At.
func CalcVars(n CalcNode) []string {
	var out []string
	var walk func(CalcNode)
	walk = func(n CalcNode) {
		switch n := n.(type) {
		case CalcVar:
			if n.At == len(out) {
				out = append(out, n.Name)
			}
		case CalcNeg:
			walk(n.X)
		case CalcBin:
			walk(n.L)
			walk(n.R)
		}
	}
	walk(n)
	return out
}

// CalcText spells an expression node the way an author would read it.
func CalcText(n CalcNode) string {
	switch n := n.(type) {
	case CalcNum:
		return strconv.FormatFloat(float64(n), 'f', -1, 64)
	case CalcVar:
		return n.Name
	case CalcNeg:
		return "-" + CalcText(n.X)
	case CalcBin:
		return "(" + CalcText(n.L) + " " + string(n.Operator) + " " + CalcText(n.R) + ")"
	}
	return "?"
}

// calcParser is a recursive-descent parser over the expression runes, threading
// expr -> term -> factor for the standard * / before + - precedence.
type calcParser struct {
	rs   []rune
	pos  int
	vars map[string]int
}

// ParseCalc parses a whole expression, requiring it to consume all input.
func ParseCalc(expr string) (CalcNode, error) {
	p := &calcParser{rs: []rune(expr), vars: map[string]int{}}
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

func (p *calcParser) expr() (CalcNode, error) { return p.binary(p.term, '+', '-') }
func (p *calcParser) term() (CalcNode, error) { return p.binary(p.factor, '*', '/') }

// binary parses a left-associative run of next() operands joined by the given
// operators, the one shape expr and term share.
func (p *calcParser) binary(next func() (CalcNode, error), ops ...byte) (CalcNode, error) {
	n, err := next()
	if err != nil {
		return nil, err
	}
	for {
		p.space()
		if p.pos >= len(p.rs) || !slices.Contains(ops, byte(p.rs[p.pos])) {
			return n, nil
		}
		op := byte(p.rs[p.pos])
		p.pos++
		r, err := next()
		if err != nil {
			return nil, err
		}
		n = CalcBin{op, n, r}
	}
}

func (p *calcParser) factor() (CalcNode, error) {
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
		return CalcNeg{x}, nil
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

func (p *calcParser) number() (CalcNode, error) {
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
	return CalcNum(v), nil
}

// ident reads a field or name: a letter or '_', then letters, digits or '_'. A '-'
// is always the minus operator, so a hyphenated one can't be an operand.
func (p *calcParser) ident() (CalcNode, error) {
	start := p.pos
	for p.pos < len(p.rs) {
		if c := p.rs[p.pos]; c == '_' || unicode.IsLetter(c) || unicode.IsDigit(c) {
			p.pos++
		} else {
			break
		}
	}
	name := string(p.rs[start:p.pos])
	at, seen := p.vars[name]
	if !seen {
		at = len(p.vars)
		p.vars[name] = at
	}
	return CalcVar{name, at}, nil
}
