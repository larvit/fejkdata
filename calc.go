package fejkdata

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// checkCalcFields runs checkOperands over each calc operand naming a field; an operand naming no
// field is a name, which checkCalcNames checks once names resolve.
func checkCalcFields(args []string, fields map[string]tree.Node) error {
	return checkOperands(args[0], builtinfunc.ParsedCalc(args[0]), func(name string) []tree.Node {
		if n, ok := fields[name]; ok {
			return []tree.Node{n}
		}
		return nil
	})
}

// checkCalcNames holds each calc of t reading a name to the checks checkCalcFields makes of a field.
func checkCalcNames(label string, t *tree.Template) error {
	for _, o := range t.Compiled().Ops() {
		tok := o.Token()
		if tok.Fn != builtinfunc.CalcName || !slices.ContainsFunc(o.Operands(), func(a tree.Arm) bool { return a.Kind() == tree.NamedRead }) {
			continue
		}
		if i := slices.IndexFunc(o.Operands(), tree.Arm.ReadsJoined); i >= 0 {
			return fmt.Errorf("%s: token {%s}: operand %s joins several values as text, never a number", label, tok.Body, o.Operands()[i].Spelling())
		}
		if err := checkOperands(tok.Args[0], builtinfunc.ParsedCalc(tok.Args[0]), operandNodes(o)); err != nil {
			return fmt.Errorf("%s: token {%s}: %w", label, tok.Body, err)
		}
	}
	return nil
}

// operandNodes lists every node each operand of o may render, once its template and names resolve.
func operandNodes(o tree.Op) func(name string) []tree.Node {
	return func(name string) []tree.Node {
		a := o.Operands()[slices.IndexFunc(o.Operands(), func(a tree.Arm) bool { return a.Spelling() == name })]
		if len(a.Leaves()) == 0 {
			panic(invariant.Broken("calc operand %q was read before it was compiled", name))
		}
		return a.Leaves()
	}
}

// checkOperands refuses an operand that is never a number and a division by a constant zero.
// operand lists every node an operand may render, nil while that is unknown.
func checkOperands(text string, c grammar.Calc, operand func(name string) []tree.Node) error {
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
func constantZeroDivisor(n grammar.CalcNode, operand func(string) []tree.Node) (string, bool) {
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
func constantValue(n grammar.CalcNode, operand func(string) []tree.Node) (float64, bool) {
	switch n := n.(type) {
	case grammar.CalcNum:
		return float64(n), true
	case grammar.CalcVar:
		nodes := operand(n.Name)
		if len(nodes) != 1 {
			return 0, false
		}
		t, ok := nodes[0].(*tree.Template)
		if !ok || t.Repeat() != 1 {
			return 0, false
		}
		lit, fixed := t.FixedText()
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

// neverNumeric reports a node no render of which is a number: a null, a repeat of 0, fixed
// text that does not parse, or a choice of only such items. text is one such render.
func neverNumeric(n tree.Node) (text string, never bool) {
	switch n := n.(type) {
	case *tree.NullItem:
		return "", true
	case *tree.Template:
		lit, fixed := n.FixedText()
		if n.Repeat() == 0 {
			return "", true
		}
		if !fixed || n.Repeat() != 1 {
			return "", false
		}
		if _, err := strconv.ParseFloat(strings.TrimSpace(lit), 64); err != nil {
			return lit, true
		}
	case *tree.Choice:
		return allNeverNumeric(n.Items())
	}
	return "", false
}

// allNeverNumeric reports whether no render of any of nodes is a number; false for no nodes.
func allNeverNumeric(nodes []tree.Node) (text string, never bool) {
	for _, n := range nodes {
		t, nodeNever := neverNumeric(n)
		if !nodeNever {
			return "", false
		}
		text = t
	}
	return text, len(nodes) > 0
}
