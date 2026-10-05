package fejkdata

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
)

// evalCalc evaluates an expression over the operand values expand already read, so it
// draws nothing and reads no template node.
func evalCalc(n grammar.CalcNode, operands []string) float64 {
	switch n := n.(type) {
	case grammar.CalcNum:
		return float64(n)
	case grammar.CalcVar:
		v, err := strconv.ParseFloat(strings.TrimSpace(operands[n.At]), 64)
		if err != nil {
			return math.NaN() // a non-numeric operand stays visible, never an error
		}
		return v
	case grammar.CalcNeg:
		return -evalCalc(n.X, operands)
	case grammar.CalcBin:
		return arith(n.Operator, evalCalc(n.L, operands), evalCalc(n.R, operands))
	}
	panic(internalError("calc node %#v has no evaluation", n))
}

func arith(operator byte, l, r float64) float64 {
	switch operator {
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
	c, err := grammar.ParseCalc(args[0])
	if err != nil {
		return fmt.Errorf("calc(%q): %w", args[0], err)
	}
	if err := checkOperands(args[0], c, func(name string) []node {
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
		if o.Fn != "calc" || !slices.ContainsFunc(o.operands, func(a arm) bool { return a.kind == namedRead }) {
			continue
		}
		if err := checkOperands(o.Args[0], parsedCalc(o.Args[0]), operandNodes(o)); err != nil {
			return fmt.Errorf("%s: token {%s}: %w", t.site.label(path), o.Body, err)
		}
	}
	return nil
}

// operandNodes lists every node each operand of o may render, once its template and names link.
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
func checkOperands(text string, c grammar.Calc, operand func(name string) []node) error {
	for _, name := range c.Operands {
		if rendered, never := allNeverNumeric(operand(name)); never {
			return fmt.Errorf("calc(%q): operand %q is never a number: it renders %q", text, name, rendered)
		}
	}
	if divisor, zero := constantZeroDivisor(c.Expr, operand); zero {
		return fmt.Errorf("calc(%q) divides by %s, which is always zero", text, divisor)
	}
	return nil
}

// constantZeroDivisor finds a division whose right side is a constant zero: number
// literals and fixed operands folded, anything that varies left unknown.
func constantZeroDivisor(n grammar.CalcNode, operand func(string) []node) (string, bool) {
	switch n := n.(type) {
	case grammar.CalcNeg:
		return constantZeroDivisor(n.X, operand)
	case grammar.CalcBin:
		if n.Operator == '/' {
			if v, known := constantValue(n.R, operand); known && v == 0 {
				return grammar.CalcText(n.R), true
			}
		}
		if d, zero := constantZeroDivisor(n.L, operand); zero {
			return d, true
		}
		return constantZeroDivisor(n.R, operand)
	}
	return "", false
}

// constantValue evaluates an expression whose every operand is fixed.
func constantValue(n grammar.CalcNode, operand func(string) []node) (float64, bool) {
	switch n := n.(type) {
	case grammar.CalcNum:
		return float64(n), true
	case grammar.CalcVar:
		nodes := operand(n.Name)
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
	case grammar.CalcNeg:
		v, ok := constantValue(n.X, operand)
		return -v, ok
	case grammar.CalcBin:
		l, lok := constantValue(n.L, operand)
		r, rok := constantValue(n.R, operand)
		if !lok || !rok {
			return 0, false
		}
		return arith(n.Operator, l, r), true
	}
	return 0, false
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
		return allNeverNumeric(n.items)
	}
	return "", false
}

// allNeverNumeric reports whether no render of any of nodes is a number; false for no nodes.
func allNeverNumeric(nodes []node) (text string, never bool) {
	for _, n := range nodes {
		t, nodeNever := neverNumeric(n)
		if !nodeNever {
			return "", false
		}
		text = t
	}
	return text, len(nodes) > 0
}

// calcPrep parses the expression and decimals once, at compile time. checkCalc proved
// both args valid, so no step here can fail.
func calcPrep(args []string) callFn {
	expr := parsedCalc(args[0]).Expr
	dp := calcDecimals(args)
	return func(_ *drawstate.State, _ string, operands []string) string {
		return formatFloat(evalCalc(expr, operands), dp)
	}
}

// parsedCalc parses an expression checkCalc accepted.
func parsedCalc(expr string) grammar.Calc {
	c, err := grammar.ParseCalc(expr)
	if err != nil {
		panic(internalError("calc(%q) passed its check unparsed: %v", expr, err))
	}
	return c
}

// calcDecimals is a calc's decimals count, or shortestDecimals where it names none.
func calcDecimals(args []string) int {
	if len(args) == 2 {
		return atoi(args[1])
	}
	return shortestDecimals
}

// calcOperands lists the operands a calc's args read, fields or names. checkCalc reports
// an expression that does not parse, so one that does not simply names nothing.
func calcOperands(args []string) []string {
	if len(args) == 0 {
		return nil
	}
	c, err := grammar.ParseCalc(args[0])
	if err != nil {
		return nil
	}
	return c.Operands
}
