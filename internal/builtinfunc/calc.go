package builtinfunc

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
	"github.com/larvit/fejkdata/internal/proven"
)

// evalCalc evaluates an expression over the operand values a render already read, so it
// draws nothing.
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
		return Arith(n.Operator, evalCalc(n.L, operands), evalCalc(n.R, operands))
	}
	panic(invariant.Broken("calc node %#v has no evaluation", n))
}

// Arith applies a calc operator.
func Arith(operator byte, l, r float64) float64 {
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

// checkCalc validates a calc token's own args: a parseable expression and an optional
// non-negative integer dp. What its operands hold is the caller's to check.
func checkCalc(args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("calc takes an expression and an optional decimals count, got %d args", len(args))
	}
	if _, err := grammar.ParseCalc(args[0]); err != nil {
		return fmt.Errorf("calc(%q): %w", args[0], err)
	}
	if len(args) == 2 {
		dp, err := intArg(args[1])
		if err != nil {
			return fmt.Errorf("calc decimals %w", err)
		}
		if dp < 0 || dp > maxDecimals {
			return fmt.Errorf("calc decimals %d must be in 0..%d", dp, maxDecimals)
		}
	}
	return nil
}

// calcPrep parses the expression and decimals once, at compile time. checkCalc proved
// both args valid, so no step here can fail.
func calcPrep(args []string) Call {
	expr := ParsedCalc(args[0]).Expr
	dp := CalcDecimals(args)
	return func(_ *drawstate.State, _ string, operands []string) string {
		return formatFloat(evalCalc(expr, operands), dp)
	}
}

// ParsedCalc parses an expression Check accepted.
func ParsedCalc(expr string) grammar.Calc {
	c, err := grammar.ParseCalc(expr)
	if err != nil {
		panic(invariant.Broken("calc(%q) passed its check unparsed: %v", expr, err))
	}
	return c
}

// CalcDecimals is the decimals count of a calc Check accepted, or proven.ShortestDecimals
// where it names none.
func CalcDecimals(args []string) int {
	if len(args) == 2 {
		return atoi(args[1])
	}
	return proven.ShortestDecimals
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
