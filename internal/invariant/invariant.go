// Package invariant spells what every package panics with when an invariant breaks.
package invariant

import "fmt"

// Broken is the panic value for a broken invariant, one phrase every break greps to.
func Broken(format string, a ...any) string {
	return "fejkdata: internal error: " + fmt.Sprintf(format, a...)
}
