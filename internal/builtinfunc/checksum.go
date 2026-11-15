package builtinfunc

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/larvit/fejkdata/internal/drawstate"
)

func luhnArg(a []string) error {
	if len(a) > 1 {
		return fmt.Errorf("luhn takes at most one count, got %d", len(a))
	}
	if len(a) == 0 {
		return nil
	}
	if err := countArg(a); err != nil {
		return fmt.Errorf("luhn(n): %w", err)
	}
	if atoi(a[0]) == 0 {
		return errors.New("luhn(n): count 0 reads no digit; write luhn() to read them all")
	}
	return nil
}

func luhnPrep(a []string) Call {
	n := -1
	if len(a) == 1 {
		n = atoi(a[0])
	}
	return func(_ *drawstate.State, emitted string, _ []string) string {
		return string(rune('0' + luhnCheck(lastDigits(emitted, n))))
	}
}

// lastDigits cuts s to start at its n-th digit from the end; a negative n, or one past the digits
// s holds, leaves it whole.
func lastDigits(s string, n int) string {
	for i := len(s) - 1; i >= 0 && n > 0; i-- {
		if s[i] >= '0' && s[i] <= '9' {
			if n--; n == 0 {
				return s[i:]
			}
		}
	}
	return s
}

// luhnCheck returns the Luhn check digit (0-9) over the digits of s; non-digit
// runes are skipped. Doubling runs from the rightmost digit, so the result is
// correct whatever the payload length.
func luhnCheck(s string) int {
	sum, double := 0, true
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if c < '0' || c > '9' {
			continue
		}
		d := int(c - '0')
		if double {
			if d *= 2; d > 9 {
				d -= 9
			}
		}
		double = !double
		sum += d
	}
	return (10 - sum%10) % 10
}

// mod11Check returns the weighted mod-11 check character over the digits of s
// (weights 2..7 cycling from the right). A would-be value of 10 emits 'X', as in
// ISBN-10 / ISO 7064; non-digits are skipped.
func mod11Check(s string) string {
	sum, w := 0, 2
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if c < '0' || c > '9' {
			continue
		}
		sum += int(c-'0') * w
		if w++; w > 7 {
			w = 2
		}
	}
	if chk := (11 - sum%11) % 11; chk != 10 {
		return string(rune('0' + chk))
	}
	return "X"
}

// eanCheck returns the EAN-13 / UPC-A / ISBN-13 / GTIN check digit over the
// digits of s: weights 3 and 1 alternating from the rightmost digit, mod 10.
func eanCheck(s string) string {
	sum, w := 0, 3
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if c < '0' || c > '9' {
			continue
		}
		sum += int(c-'0') * w
		w = 4 - w // 3 <-> 1
	}
	return string(rune('0' + (10-sum%10)%10))
}

// ibanLen maps a supported country code to the full IBAN length. The check digits
// sit between the country code and the BBAN, so — unlike luhn/ean — iban can't be
// a left-to-right derivation; it generates the whole value instead.
var ibanLen = map[string]int{"BE": 16, "DE": 22, "DK": 18, "ES": 24, "FI": 18, "NO": 15, "SE": 24}

// swedishIBANIDs are the IBAN IDs in Bankinfrastruktur i Sverige's clearing number table.
var swedishIBANIDs = []string{
	"120", "211", "212", "213", "230", "300", "500", "600", "800", "902", "904", "907", "910", "915",
	"917", "918", "919", "923", "925", "927", "928", "930", "939", "940", "946", "947", "950", "955",
	"957", "958", "959", "960", "963", "964", "965", "966", "967", "968", "970", "971", "975", "978",
}

func ibanArg(a []string) error {
	if _, ok := ibanLen[a[0]]; !ok {
		return fmt.Errorf("iban(%q): unsupported country code", a[0])
	}
	return nil
}

func randomDigits(r rng, n int) string {
	d := make([]byte, n)
	for i := range d {
		d[i] = byte('0' + r.IntN(10))
	}
	return string(d)
}

// spanishControl returns the Banco de España control digit of ten digits.
func spanishControl(ten string) byte {
	sum := 0
	for i, w := range [...]int{1, 2, 4, 8, 5, 10, 9, 7, 3, 6} {
		sum += int(ten[i]-'0') * w
	}
	switch dc := 11 - sum%11; dc {
	case 10:
		return '1'
	case 11:
		return '0'
	default:
		return byte('0' + dc)
	}
}

// bban draws cc's account number: BE, ES, FI and NO with their account check digits, SE with a
// listed IBAN ID, and any other country as random digits.
func bban(r rng, cc string) string {
	if draw, ok := bbanDraws[cc]; ok {
		return draw(r)
	}
	return randomDigits(r, ibanLen[cc]-4)
}

var bbanDraws = map[string]func(rng) string{
	"BE": belgianBBAN,
	"ES": spanishBBAN,
	"FI": finnishBBAN,
	"NO": norwegianBBAN,
	"SE": func(r rng) string { return swedishIBANIDs[r.IntN(len(swedishIBANIDs))] + randomDigits(r, 17) },
}

func belgianBBAN(r rng) string {
	n := randomDigits(r, 10)
	v, _ := strconv.Atoi(n)
	if v %= 97; v == 0 {
		v = 97
	}
	return fmt.Sprintf("%s%02d", n, v)
}

func spanishBBAN(r rng) string {
	bank, account := randomDigits(r, 8), randomDigits(r, 10)
	return bank + string([]byte{spanishControl("00" + bank), spanishControl(account)}) + account
}

func finnishBBAN(r rng) string {
	n := randomDigits(r, 13)
	return n + string(rune('0'+luhnCheck(n)))
}

func norwegianBBAN(r rng) string {
	for {
		n := randomDigits(r, 10)
		if c := mod11Check(n); c != "X" {
			return n + c
		}
	}
}

func iban(r rng, cc string) string {
	account := bban(r, cc)
	rem := 0
	feed := func(d int) { rem = (rem*10 + d) % 97 }
	for _, c := range []byte(account) {
		feed(int(c - '0'))
	}
	for i := 0; i < len(cc); i++ { // letters A-Z -> 10..35, fed as two digits
		v := int(cc[i]-'A') + 10
		feed(v / 10)
		feed(v % 10)
	}
	feed(0)
	feed(0)
	return fmt.Sprintf("%s%02d%s", cc, 98-rem, account)
}
