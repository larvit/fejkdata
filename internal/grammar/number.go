package grammar

import "strings"

// Underflows reports a number written as text that reads as f = 0 though a digit before its
// exponent is not 0, as 1e-400 does: no float64 holds the number written.
func Underflows(text string, f float64) bool {
	mantissa, _, _ := strings.Cut(strings.ToLower(text), "e")
	return f == 0 && strings.ContainsAny(mantissa, "123456789")
}
