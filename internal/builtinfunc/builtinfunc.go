// Package builtinfunc holds the {name(args)} functions a format calls: their checks, the
// draws they make, and the numbers their text is proven to read as.
package builtinfunc

import (
	"fmt"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/avoid"
	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/invariant"
	"github.com/larvit/fejkdata/internal/proven"
)

// Call is a builtin prepared for one call site: its args already parsed. It reads the
// output emitted so far in the current expansion (a derivation's payload) and the
// values of the operands it named, which the caller read for it.
type Call func(s *drawstate.State, emitted string, operands []string) string

// builtin is a format-string function invoked as {name(args)}. Its Call stays pure over its
// inputs, so seeded output is reproducible. arity is the exact arg count, or -1 for variadic,
// where checkArgs does all the validation.
type builtin struct {
	arity int
	calls []string
	// prep turns validated args into the closure a render calls, at compile time. It is nil
	// for adddays: AddDays builds that call once the load knows the date's layout.
	prep func(args []string, p parsed) Call
	// checkArgs refuses a bad arg and returns the numbers it read, the one parse of them: prep,
	// proveNumber and avoids take those and never parse a number themselves.
	checkArgs func(args []string) (parsed, error)
	// operands names the fields the call reads, which a render reads for it; nil
	// for a builtin that reads none.
	operands func(args []string) []string
	// proveNumber bounds the number a call's text reads as, token its body, and says which
	// datatypes that text is not; set it where every render reads as a finite number, which
	// makes the call a calc operand, and leave it nil otherwise.
	proveNumber func(token string, prints datatype.DataType, p parsed) proven.Facts
	// prints is the datatype a call's text is, handed to proveNumber: datatype.String where it
	// reads as a number no column should type, as digits' leading zeros; unset where
	// proveNumber is nil.
	prints datatype.DataType
	// avoids is the runs of every value a sample prints, and what it prints where no value passes
	// its where clause; nil for a builtin a where clause can't filter.
	avoids func(p parsed) (avoid.Runs, string)
}

// parsed is the integers and the floats checkArgs read from a call's args, each in the order
// the call writes them.
type parsed struct {
	ints   []int
	floats []float64
}

// plain makes a checkArgs for a builtin whose args hold no number: it checks and returns none.
func plain(check func(args []string) error) func([]string) (parsed, error) {
	return func(a []string) (parsed, error) { return parsed{}, check(a) }
}

// CalcName is the builtin whose operands internal/engine/fence checks and proves.
const CalcName = "calc"

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
	_, err := b.parse(args)
	return err
}

// parse runs a builtin's check over args. A builtin with no check takes args of any kind.
func (b builtin) parse(args []string) (parsed, error) {
	if b.checkArgs == nil {
		return parsed{}, nil
	}
	return b.checkArgs(args)
}

// accepted parses args that Check accepted. A failure means they skipped Check.
func accepted(name string, args []string) parsed {
	p, err := builtins[name].parse(args)
	if err != nil {
		panic(invariant.Broken("%s(%q) skipped Check: %v", name, args, err))
	}
	return p
}

// Prep parses a call's args, which Check accepted, into the Call a render makes, or nil for adddays.
func Prep(name string, args []string) Call {
	if b := builtins[name]; b.prep != nil {
		return b.prep(args, accepted(name, args))
	}
	return nil
}

// Operands lists the fields a call reads as operands, empty for a builtin that reads none
// or is unknown.
func Operands(name string, args []string) []string {
	b, known := builtins[name]
	if !known || b.operands == nil {
		return nil
	}
	return b.operands(args)
}

// Calls lists every way to call a builtin, sorted.
func Calls() []string {
	var calls []string
	for _, b := range builtins {
		calls = append(calls, b.calls...)
	}
	slices.Sort(calls)
	return calls
}

// IsTransform reports a builtin that rewrites the text of its one operand.
func IsTransform(name string) bool {
	_, isTransform := transforms[name]
	return isTransform
}

// ProveNumber proves a call whose every render reads as a finite number, token its body;
// false for any other call, calc included: proven.OfCalc proves a calc from its operands,
// which only the caller holds.
func ProveNumber(name, token string, args []string) (proven.Facts, bool) {
	b := builtins[name]
	if b.proveNumber == nil {
		return proven.Facts{}, false
	}
	return b.proveNumber(token, b.prints, accepted(name, args)), true
}

// Runs is every value name(args) can print; ok is false for a builtin a where clause can't filter.
func Runs(name string, args []string) (rs avoid.Runs, ok bool) {
	b := builtins[name]
	if b.avoids == nil {
		return nil, false
	}
	rs, _ = b.avoids(accepted(name, args))
	return rs, true
}

// Default is what name(args) prints where no value passes its where clause, and the clause names no
// fallback.
func Default(name string, args []string) string {
	_, d := builtins[name].avoids(accepted(name, args))
	return d
}

// FilteredCalls lists the calls a where clause can filter, as an error names them.
func FilteredCalls() string {
	var calls []string
	for name, b := range builtins {
		if b.avoids != nil {
			calls = append(calls, "{"+name+"()}")
		}
	}
	slices.Sort(calls)
	return strings.Join(calls[:len(calls)-1], ", ") + " or " + calls[len(calls)-1]
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
	return strings.Join(calls, ", ") + " or {" + CalcName + "()}"
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
