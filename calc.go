package fejkdata

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// checkCalcFields runs checkOperands over the calc operands fields hold. An operand no field
// holds is left for a name, which checkCalcNames checks once names link.
func checkCalcFields(args []string, fields map[string]node) error {
	return checkOperands(args[0], builtinfunc.ParsedCalc(args[0]), func(name string) []node {
		if n, ok := fields[name]; ok {
			return []node{n}
		}
		return nil
	})
}

// checkCalcNames holds each calc of t reading a name to the checks checkCalcFields makes of a field.
func checkCalcNames(path string, t *template) error {
	for _, o := range t.compiled.ops {
		if o.Fn != builtinfunc.CalcName || !slices.ContainsFunc(o.operands, func(a arm) bool { return a.kind == namedRead }) {
			continue
		}
		if err := checkOperands(o.Args[0], builtinfunc.ParsedCalc(o.Args[0]), operandNodes(o)); err != nil {
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
			panic(invariant.Broken("calc operand %q was read before it was compiled", name))
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
		return builtinfunc.Arith(n.Operator, l, r), true
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
