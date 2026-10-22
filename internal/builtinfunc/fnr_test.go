package builtinfunc

import (
	"fmt"
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

// TestFnrCompletesEveryPrefixForEveryDigit proves fnr never meets a prefix it cannot complete:
// the completions depend on the seven digits only through their weighted sums for k1 and k2
// modulo 11, so a prefix for each of the 121 pairs of residues, with every third individ digit
// and every start, covers every prefix and every operand.
func TestFnrCompletesEveryPrefixForEveryDigit(t *testing.T) {
	residue := func(prefix string, weights []int) int {
		sum := 0
		for i, w := range weights {
			sum += w * int(prefix[i]-'0')
		}
		return sum % 11
	}
	seen := map[[2]int]bool{}
	for n := 0; len(seen) < 121; n++ {
		if n == 10_000_000 {
			t.Fatalf("seven digits reach %d of the 121 pairs of residues", len(seen))
		}
		prefix := fmt.Sprintf("%07d", n)
		r := [2]int{residue(prefix, []int{3, 7, 6, 1, 8, 9, 4}), residue(prefix, []int{5, 4, 3, 2, 7, 6, 5})}
		if seen[r] {
			continue
		}
		seen[r] = true
		for last := byte('0'); last <= '9'; last++ {
			for start := range 10 {
				tail := fnrTail(prefix, last, start)
				if len(tail) != 4 || tail[1] != last || !validFnr(prefix+tail) {
					t.Fatalf("fnrTail(%s, %c, %d) = %q, want an individ digit, %c, and valid check digits", prefix, last, start, tail, last)
				}
			}
		}
	}
}

func TestFnrTailStartsWhereTheDrawSays(t *testing.T) {
	seen := map[string]bool{}
	for _, last := range []byte("13579") {
		for start := range 10 {
			seen[fnrTail("0181804", last, start)] = true
		}
	}
	if len(seen) < 40 {
		t.Errorf("50 draws gave %d tails, want at least 40, since at most two second individ digits fail for each third", len(seen))
	}
}

func TestFnrReadsTheLastSevenDigitsEmitted(t *testing.T) {
	if got, want := fnrTail("x-0181804", '7', 3), fnrTail("0181804", '7', 3); got != want {
		t.Errorf("fnrTail with text before the digits = %q, want %q", got, want)
	}
}
