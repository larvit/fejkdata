package builtinfunc

import (
	"strconv"
	"strings"
	"testing"
)

// validFnr applies the weights Skatteetaten publishes: k1 over the nine digits with 3, 7, 6, 1,
// 8, 9, 4, 5, 2, k2 over those and k1 with 5, 4, 3, 2, 7, 6, 5, 4, 3, 2; each is 11 less the
// sum modulo 11, a sum of 11 giving 0, and a 10 discards the number.
func validFnr(s string) bool {
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

func TestValidFnrAcceptsAPublishedTestNumber(t *testing.T) {
	if !validFnr("10108000398") {
		t.Error("10108000398 is a published test number, want it valid")
	}
	for _, bad := range []string{"10108000399", "10108000388", "1010800039", "10108000398x"} {
		if validFnr(bad) {
			t.Errorf("%s is no valid fødselsnummer", bad)
		}
	}
}

// TestFnrCompletesEverySevenDigitPrefix proves fnr never meets a prefix it cannot complete:
// every seven digits, with either parity of the last individ digit, take two individ digits
// and two check digits that make a valid number.
func TestFnrCompletesEverySevenDigitPrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("walks ten million prefixes")
	}
	for _, last := range []string{"02468", "13579"} {
		for n := 0; n < 10_000_000; n++ {
			prefix := strconv.Itoa(n)
			prefix = strings.Repeat("0", 7-len(prefix)) + prefix
			tail := fnrTail(prefix, last, n)
			if full := prefix + tail; len(tail) != 4 || !validFnr(full) || !strings.Contains(last, tail[1:2]) {
				t.Fatalf("fnrTail(%s, %s) = %q, want two individ digits, the second from %s, and valid check digits", prefix, last, tail, last)
			}
		}
	}
}

func TestFnrTailStartsWhereTheDrawSays(t *testing.T) {
	seen := map[string]bool{}
	for start := 0; start < 50; start++ {
		seen[fnrTail("0181804", "13579", start)] = true
	}
	if len(seen) < 25 {
		t.Errorf("50 starts gave %d tails, want at least 25", len(seen))
	}
}

func TestFnrReadsTheLastSevenDigitsEmitted(t *testing.T) {
	if got, want := fnrTail("x-0181804", "13579", 7), fnrTail("0181804", "13579", 7); got != want {
		t.Errorf("fnrTail with text before the digits = %q, want %q", got, want)
	}
}

func TestFnrTakesAnyDigitWhenTheOperandHoldsNone(t *testing.T) {
	tail := fnrTail("0181804", "", 3)
	if !validFnr("0181804" + tail) {
		t.Errorf("fnrTail with no digits to choose from = %q, want a valid completion", tail)
	}
}
