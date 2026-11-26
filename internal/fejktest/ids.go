package fejktest

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

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

// ValidIsin is the check on an ISIN in ISO 6166: two capital letters, nine digits or capital
// letters and a check digit. It writes each letter as its value, `A` 10 to `Z` 35, and then
// requires the whole string of digits to pass the Luhn check.
func ValidIsin(s string) bool {
	if !isinShape.MatchString(s) {
		return false
	}
	var digits strings.Builder
	for i := 0; i < len(s); i++ {
		digits.WriteString(strconv.Itoa(charValue(s[i])))
	}
	return LuhnValid(digits.String())
}

var isinShape = regexp.MustCompile(`^[A-Z]{2}[0-9A-Z]{9}[0-9]$`)

// ValidCusip is the check on a CUSIP in ANSI X9.6: eight digits, capital letters, `*`, `@` or
// `#`, and a check digit. Each of the eight has a value, `0` to `9`, `A` 10 to `Z` 35, `*` 36, `@` 37
// and `#` 38, and every second one, from the second, is doubled. The check digit brings the sum
// of the digits of those values up to a multiple of 10.
func ValidCusip(s string) bool {
	if !cusipShape.MatchString(s) {
		return false
	}
	sum := 0
	for i := 0; i < 8; i++ {
		v := charValue(s[i])
		if i%2 == 1 {
			v *= 2
		}
		sum += v/10 + v%10
	}
	return int(s[8]-'0') == (10-sum%10)%10
}

var cusipShape = regexp.MustCompile(`^[0-9A-Z*@#]{8}[0-9]$`)

// charValue is the value of a character the two shape expressions admit.
func charValue(c byte) int {
	if c >= '0' && c <= '9' {
		return int(c - '0')
	}
	if c >= 'A' && c <= 'Z' {
		return int(c-'A') + 10
	}
	return 36 + strings.IndexByte("*@#", c)
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
	d, ok := steuerDigits(s)
	return ok && oneDigitTwiceOrThrice(d) && !threeEqualInARow(d) && d[10] == steuerCheckDigit(d)
}

func steuerDigits(s string) (d [11]int, ok bool) {
	if len(s) != len(d) {
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

func oneDigitTwiceOrThrice(d [11]int) bool {
	var count [10]int
	for _, x := range d[:10] {
		count[x]++
	}
	repeated := 0
	for _, n := range count {
		if n > 3 {
			return false
		}
		if n > 1 {
			repeated++
		}
	}
	return repeated == 1
}

func threeEqualInARow(d [11]int) bool {
	for i := 2; i < 10; i++ {
		if d[i] == d[i-1] && d[i] == d[i-2] {
			return true
		}
	}
	return false
}

func steuerCheckDigit(d [11]int) int {
	product := 10
	for _, x := range d[:10] {
		sum := (x + product) % 10
		if sum == 0 {
			sum = 10
		}
		product = 2 * sum % 11
	}
	return (11 - product) % 10
}

// ValidNir is INSEE's check of a NIR: thirteen digits, where Corsica's department may be 2A or
// 2B, then a key that is 97 less the remainder of those digits divided by 97, with 2A and 2B
// counted as 19 and 18.
func ValidNir(s string) bool {
	if len(s) != 15 {
		return false
	}
	number := strings.NewReplacer("2A", "19", "2B", "18").Replace(s[5:7])
	number = s[:5] + number + s[7:13]
	if DigitsOnly(number) != number || DigitsOnly(s[13:]) != s[13:] {
		return false
	}
	n, _ := strconv.Atoi(number)
	key, _ := strconv.Atoi(s[13:])
	return key == 97-n%97
}

// ValidIsikukood is the check of an Estonian personal code, with python-stdnum's ee.ik check digit:
// eleven digits, a first digit 1 to 6, a date that exists in the century the
// first digit names, and a check digit that is the weighted sum of the first ten digits, weights 1
// to 9 and 1, modulo 11; where that is 10, the sum weighted 3 to 9, 1, 2 and 3, modulo 11, and
// where that is 10 too, 0.
func ValidIsikukood(s string) bool {
	if len(s) != 11 || DigitsOnly(s) != s || s[0] == '0' || s[0] > '6' {
		return false
	}
	year := 1800 + 100*(int(s[0]-'1')/2) + atoi(s[1:3])
	month, day := atoi(s[3:5]), atoi(s[5:7])
	if d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC); d.Month() != time.Month(month) || d.Day() != day {
		return false
	}
	return int(s[10]-'0') == isikukoodCheckDigit(s)
}

// ValidAsmensKodas is the check of a Lithuanian personal code, which python-stdnum's lt.asmens takes
// from ee.ik: the structure and the check digit of an Estonian one. It refuses a code whose date
// doesn't exist, as one with a month or day of 00 does, and one whose first digit is 9. The
// Lithuanian Wikipedia describes both forms.
func ValidAsmensKodas(s string) bool { return ValidIsikukood(s) }

func isikukoodCheckDigit(s string) int {
	sum := func(weights string) int {
		n := 0
		for i := range 10 {
			n += int(weights[i]-'0') * int(s[i]-'0')
		}
		return n % 11
	}
	if check := sum("1234567891"); check != 10 {
		return check
	}
	return sum("3456789123") % 10
}

// ValidPersonasKods is the check of a Latvian personal code, with the weights of the State Revenue
// Service's formula as Wikipedia gives them, which yield python-stdnum's digit: eleven digits, where either the first two are 32, a
// code with no date, or the first six are a date that exists in the century the seventh digit
// names, 0 for the 1800s, 1 for the 1900s and 2 for the 2000s; and a check digit that is 1101
// less the sum of the first ten digits weighted 1, 6, 3, 7, 9, 10, 5, 8, 4 and 2, modulo 11, and
// modulo 10 where that is 10.
func ValidPersonasKods(s string) bool {
	if len(s) != 11 || DigitsOnly(s) != s || s[:2] != "32" && !personasKodsDateExists(s) {
		return false
	}
	sum := 0
	for i, w := range [10]int{1, 6, 3, 7, 9, 10, 5, 8, 4, 2} {
		sum += w * int(s[i]-'0')
	}
	return int(s[10]-'0') == (1101-sum)%11%10
}

// personasKodsDateExists is whether the first seven digits of a code are a date that exists, in the
// century the seventh names.
func personasKodsDateExists(s string) bool {
	year := 1800 + 100*int(s[6]-'0') + atoi(s[4:6])
	month, day := atoi(s[2:4]), atoi(s[:2])
	d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return s[6] <= '2' && d.Month() == time.Month(month) && d.Day() == day
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
