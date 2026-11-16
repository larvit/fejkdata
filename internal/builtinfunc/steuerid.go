package builtinfunc

// steuerID draws a Steueridentifikationsnummer of ELSTER's test series: the first of its ten digits is
// 0, which the BZSt doesn't give any issued number. Among the ten, one digit appears twice or three times,
// never three in a row, and the eleventh digit is the check digit of ISO 7064 MOD 11,10.
func steuerID(r rng) string {
	repeats := 2 + r.IntN(2)
	for {
		digits, ok := steuerDigits(r, repeats)
		if ok {
			return string(digits[:]) + string(rune('0'+steuerCheck(digits)))
		}
	}
}

// steuerDigits draws the first ten digits, a zero first and one digit repeats times. It reports false
// when the draw holds no zero or puts three equal digits in a row.
func steuerDigits(r rng, repeats int) (digits [10]byte, ok bool) {
	repeated := byte(r.IntN(10))
	others := make([]byte, 0, 9)
	for d := range byte(10) {
		if d != repeated {
			others = append(others, d)
		}
	}
	shuffle(r, others)
	multiset := make([]byte, 0, 10)
	for range repeats {
		multiset = append(multiset, repeated)
	}
	multiset = append(multiset, others[:10-repeats]...)
	zero := indexOfDigit(multiset, 0)
	if zero < 0 {
		return digits, false
	}
	multiset[0], multiset[zero] = multiset[zero], multiset[0]
	shuffle(r, multiset[1:])
	for i, d := range multiset {
		digits[i] = '0' + d
	}
	return digits, !threeInARow(digits)
}

func shuffle(r rng, s []byte) {
	for i := len(s) - 1; i > 0; i-- {
		j := r.IntN(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}

func indexOfDigit(s []byte, d byte) int {
	for i, x := range s {
		if x == d {
			return i
		}
	}
	return -1
}

func threeInARow(digits [10]byte) bool {
	for i := 2; i < len(digits); i++ {
		if digits[i] == digits[i-1] && digits[i] == digits[i-2] {
			return true
		}
	}
	return false
}

// steuerCheck is ELSTER's check digit, ISO 7064 MOD 11,10.
func steuerCheck(digits [10]byte) int {
	product := 10
	for _, d := range digits {
		sum := (int(d-'0') + product) % 10
		if sum == 0 {
			sum = 10
		}
		product = 2 * sum % 11
	}
	return (11 - product) % 10
}
