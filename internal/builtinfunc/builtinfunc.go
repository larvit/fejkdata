// Package builtinfunc holds the {name(args)} functions a format calls: their checks, the
// draws they make, and the numbers their text is proven to read as.
package builtinfunc

import (
	"fmt"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/proven"
)

// Call is a builtin prepared for one call site: its args already parsed. It reads the
// output emitted so far in the current expansion (a derivation's payload) and the
// values of the operands it named, which the caller read for it.
type Call func(s *drawstate.State, emitted string, operands []string) string

// builtin is a format-string function invoked as {name(args)}. It receives the
// draw state, the output emitted so far in the current expansion (for derivations
// such as a checksum over preceding digits), and the values of the operands it named
// (calc and the transforms name them). All must stay pure over (draw state, emitted, operands) so
// seeded output is reproducible. arity is the exact arg count, or -1 for variadic (then
// checkArgs does all the validation).
type builtin struct {
	arity int
	// prep parses validated args once, at compile time, into the closure a render calls.
	prep      func(args []string) Call
	checkArgs func(args []string) error
	// operands names the fields the call reads, which a render reads for it; nil
	// for a builtin that reads none.
	operands func(args []string) []string
	// noRefOperands says no operand can be a reference.
	noRefOperands bool
	// proveNumber bounds the number a call's text reads as, token its body, and says which
	// datatypes that text is not; set it where every render reads as a finite number, which
	// makes the call a calc operand, and leave it nil otherwise.
	proveNumber func(token string, prints datatype.DataType, args []string) proven.Value
	// prints is the datatype a call's text is, handed to proveNumber: datatype.String where it
	// reads as a number no column should type, as digits' leading zeros; unset where
	// proveNumber is nil.
	prints datatype.DataType
}

// Calc is the name of the builtin whose operands the caller checks and proves, as nodes only
// it holds.
const Calc = "calc"

// rng is the randomness a builtin sample draws from, which *drawstate.State satisfies.
type rng interface {
	IntN(n int) int
	Float64() float64
}

// Check validates a call at compile time: naming a known builtin, with the arg count that
// builtin takes and args its check accepts.
func Check(name string, args []string) error {
	b, known := builtins[name]
	if !known {
		return fmt.Errorf("unknown function %q", name)
	}
	if b.arity >= 0 && len(args) != b.arity {
		return fmt.Errorf("%s takes %d argument%s, got %d", name, b.arity, plural(b.arity), len(args))
	}
	if b.checkArgs != nil {
		return b.checkArgs(args)
	}
	return nil
}

// Prep parses a call's args, which Check accepted, into the Call a render makes.
func Prep(name string, args []string) Call { return builtins[name].prep(args) }

// Operands lists the fields a call reads as operands, empty for a builtin that reads none
// or is unknown.
func Operands(name string, args []string) []string {
	b, known := builtins[name]
	if !known || b.operands == nil {
		return nil
	}
	return b.operands(args)
}

// NoRefOperands reports a builtin no operand of which can be a reference.
func NoRefOperands(name string) bool { return builtins[name].noRefOperands }

// IsTransform reports a builtin that rewrites the text of its one operand.
func IsTransform(name string) bool {
	_, isTransform := transforms[name]
	return isTransform
}

// ProveNumber proves a call whose every render reads as a finite number, token its body;
// false for any other call. A calc's proof reads its operands, which the caller proves.
func ProveNumber(name, token string, args []string) (proven.Value, bool) {
	b := builtins[name]
	if b.proveNumber == nil {
		return proven.Value{}, false
	}
	return b.proveNumber(token, b.prints, args), true
}

// TypedCalls lists the calls whose text a typed column may hold, as an error names them.
func TypedCalls() string { return numberCalls(false) }

// OperandCalls lists the calls whose text a calc reads as a number, as an error names them.
func OperandCalls() string { return numberCalls(true) }

func numberCalls(text bool) string {
	var calls []string
	for name, b := range builtins {
		if b.proveNumber != nil && (text || b.prints != datatype.String) {
			calls = append(calls, "{"+name+"()}")
		}
	}
	slices.Sort(calls)
	return strings.Join(calls, ", ") + " or {calc()}"
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
