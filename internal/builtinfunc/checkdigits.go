package builtinfunc

import (
	"fmt"
	"slices"
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
)

// alnumValue is the ISIN value of c: 0 to 9 for a digit, 10 to 35 for a capital letter.
func alnumValue(c byte) (int, bool) {
	if c >= '0' && c <= '9' {
		return int(c - '0'), true
	}
	if c >= 'A' && c <= 'Z' {
		return int(c-'A') + 10, true
	}
	return 0, false
}

// lastValues returns, left to right, the values of the last n characters of s that value accepts,
// so that a literal before the payload, such as `ISIN `, stays out of the check.
func lastValues(s string, n int, value func(byte) (int, bool)) []int {
	var rev []int
	for i := len(s) - 1; i >= 0 && len(rev) < n; i-- {
		if v, ok := value(s[i]); ok {
			rev = append(rev, v)
		}
	}
	slices.Reverse(rev)
	return rev
}

// isinCheck is the Luhn digit over the last 11 digits and capital letters of s, each letter
// written as its two-digit value.
func isinCheck(s string) string {
	var digits strings.Builder
	for _, v := range lastValues(s, 11, alnumValue) {
		fmt.Fprint(&digits, v)
	}
	return string(rune('0' + luhnCheck(digits.String())))
}

// cusipValue is c's value in a CUSIP: its ISIN value, or 36, 37 and 38 for `*`, `@` and `#`.
func cusipValue(c byte) (int, bool) {
	if v, ok := alnumValue(c); ok {
		return v, true
	}
	if i := strings.IndexByte("*@#", c); i >= 0 {
		return 36 + i, true
	}
	return 0, false
}

// cusipCheck doubles the last of the final 8 values of s and every second one before it, then
// sums the digits of each, as ANSI X9.6 asks.
func cusipCheck(s string) string {
	values := lastValues(s, 8, cusipValue)
	sum, double := 0, true
	for i := len(values) - 1; i >= 0; i-- {
		v := values[i]
		if double {
			v *= 2
		}
		double = !double
		sum += v/10 + v%10
	}
	return string(rune('0' + (10-sum%10)%10))
}

// abaCheck weights the last 8 digits of s 3, 7, 1, 3, 7, 1, 3, 7. It counts from the right, so
// fewer than 8 digits read as if zeros came first.
func abaCheck(s string) string {
	digits := lastValues(s, 8, func(c byte) (int, bool) { return int(c - '0'), c >= '0' && c <= '9' })
	sum := 0
	for i, d := range digits {
		sum += d * [...]int{7, 3, 1}[(len(digits)-1-i)%3]
	}
	return string(rune('0' + (10-sum%10)%10))
}

const (
	vinYears = "ABCDEFGHJKLMNPRSTVWXY123456789"
	// vinLetters and vinLetterValues are parallel: each letter's value in the check sum.
	vinLetters      = "ABCDEFGHJKLMNPRSTUVWXYZ"
	vinChars        = decimalDigits + vinLetters
	vinLetterValues = "12345678123457923456789"
)

var vinWeights = [17]int{8, 7, 6, 5, 4, 3, 2, 10, 0, 9, 8, 7, 6, 5, 4, 3, 2}

func vinArg(a []string) error {
	if len(a[0]) != 3 || strings.Trim(a[0], vinChars) != "" {
		return fmt.Errorf("vin(%q): a WMI is three digits or capital letters other than I, O and Q", a[0])
	}
	return nil
}

func vinCheck(vin string) byte {
	sum := 0
	for i, w := range vinWeights {
		sum += vinValue(vin[i]) * w
	}
	if sum %= 11; sum == 10 {
		return 'X'
	}
	return byte('0' + sum)
}

func vinValue(c byte) int {
	if c <= '9' {
		return int(c - '0')
	}
	return int(vinLetterValues[strings.IndexByte(vinLetters, c)] - '0')
}

// vin draws a VIN of the given WMI: five characters of vehicle description, a model year code,
// a plant and six digits of serial. The check character replaces the `0` at position 9, and
// is the one that makes the sum of all 17 positions, weighted, divisible as the VIN standard asks.
func vin(r rng, wmi string) string {
	b := []byte(wmi + randChars(r, 5, vinChars) + "0" + randChars(r, 1, vinYears) +
		randChars(r, 1, vinChars) + randChars(r, 6, decimalDigits))
	b[8] = vinCheck(string(b))
	return string(b)
}

func vinPrep(a []string, _ parsed) Call {
	wmi := a[0]
	return func(s *drawstate.State, _ string, _ []string) string { return vin(s, wmi) }
}
