package fejkdata

import "fmt"

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

func ibanArg(_ map[string]node, a []string) error {
	if _, ok := ibanLen[a[0]]; !ok {
		return fmt.Errorf("iban(%q): unsupported country code", a[0])
	}
	return nil
}

// iban generates a structurally valid IBAN for cc: a numeric BBAN of the right
// length, then mod-97 check digits. Real bank/branch structure isn't modelled —
// the result passes length and checksum validation, which is what fake data needs.
func iban(r rng, cc string) string {
	bban := make([]byte, ibanLen[cc]-4)
	for i := range bban {
		bban[i] = byte('0' + r.IntN(10))
	}
	rem := 0
	feed := func(d int) { rem = (rem*10 + d) % 97 }
	for _, c := range bban {
		feed(int(c - '0'))
	}
	for i := 0; i < len(cc); i++ { // letters A-Z -> 10..35, fed as two digits
		v := int(cc[i]-'A') + 10
		feed(v / 10)
		feed(v % 10)
	}
	feed(0)
	feed(0)
	return fmt.Sprintf("%s%02d%s", cc, 98-rem, bban)
}
