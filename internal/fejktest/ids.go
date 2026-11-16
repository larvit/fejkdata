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

func elevenDigits(s string) ([11]int, bool) {
	var d [11]int
	if len(s) != 11 {
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

func ValidFnr(s string) bool {
	d, ok := elevenDigits(s)
	if !ok {
		return false
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

// ValidSteuerID is the check on a Steueridentifikationsnummer in the BZSt's "Erzeugung von
// Identifikationsnummern" and ELSTER's "Prüfung der Steuer- und Steueridentifikationsnummer": eleven
// digits, of the first ten exactly one digit twice or three times and never three in a row, and an
// eleventh that is the check digit of ISO 7064 MOD 11,10 over the first ten. It accepts a leading
// zero, which ELSTER allows for a test number.
func ValidSteuerID(s string) bool {
	if len(s) != 11 {
		return false
	}
	var count [10]int
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
		if i < 10 {
			count[s[i]-'0']++
		}
	}
	repeated := 0
	for _, n := range count {
		switch {
		case n > 3:
			return false
		case n > 1:
			repeated++
		}
	}
	for i := 2; i < 10; i++ {
		if s[i] == s[i-1] && s[i] == s[i-2] {
			return false
		}
	}
	product := 10
	for i := range 10 {
		sum := (int(s[i]-'0') + product) % 10
		if sum == 0 {
			sum = 10
		}
		product = 2 * sum % 11
	}
	check := (11 - product) % 10
	return repeated == 1 && int(s[10]-'0') == check
}
