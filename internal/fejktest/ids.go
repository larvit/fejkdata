package fejktest

import "strings"

// CprDigits reads a CPR number, written with or without its hyphen, as ten digits.
func CprDigits(s string) (d [10]int, ok bool) {
	if len(s) == 11 && s[6] == '-' {
		s = s[:6] + s[7:]
	}
	if len(s) != 10 {
		return d, false
	}
	for i := range d {
		if s[i] < '0' || s[i] > '9' {
			return d, false
		}
		d[i] = int(s[i] - '0')
	}
	return d, true
}

// ValidCpr is CPR-kontoret's check in "Personnummeret i CPR-systemet", appendix "Beregning af
// kontrolciffer": the ten digits, weighted 4 3 2 7 6 5 4 3 2 1, add up to a multiple of 11.
func ValidCpr(s string) bool {
	d, ok := CprDigits(s)
	if !ok {
		return false
	}
	sum := 0
	for i, w := range []int{4, 3, 2, 7, 6, 5, 4, 3, 2, 1} {
		sum += w * d[i]
	}
	return sum%11 == 0
}

// CprYear is the year that the table "Personnummerets opbygning", in the same document, gives for
// year digits yy and seventh digit seventh.
func CprYear(yy, seventh int) int {
	switch {
	case seventh <= 3:
		return 1900 + yy
	case seventh == 4 || seventh == 9:
		if yy <= 36 {
			return 2000 + yy
		}
		return 1900 + yy
	default:
		if yy <= 57 {
			return 2000 + yy
		}
		return 1800 + yy
	}
}

func ValidFnr(s string) bool {
	if len(s) != 11 {
		return false
	}
	var d [11]int
	for i := range d {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
		d[i] = int(s[i] - '0')
	}
	check := func(weights []int, want int) bool {
		sum := 0
		for i, w := range weights {
			sum += w * d[i]
		}
		k := 11 - sum%11
		return k != 10 && k%11 == want
	}
	return check([]int{3, 7, 6, 1, 8, 9, 4, 5, 2}, d[9]) && check([]int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}, d[10])
}

func DigitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// LuhnValid verifies a full number (payload + trailing check digit). It doubles
// from the second-from-right, independent of the builtin's own Luhn code.
func LuhnValid(s string) bool {
	sum, double := 0, false
	for i := len(s) - 1; i >= 0; i-- {
		n := int(s[i] - '0')
		if double {
			if n *= 2; n > 9 {
				n -= 9
			}
		}
		double = !double
		sum += n
	}
	return sum%10 == 0
}

// DVVControlTable maps the remainder of a hetu's nine digits divided by 31 to its control
// character, per DVV's "Personal identity code" page: ddmmyy, a century sign, an individual number
// and that character.
const DVVControlTable = "0123456789ABCDEFHJKLMNPRSTUVWXY"
