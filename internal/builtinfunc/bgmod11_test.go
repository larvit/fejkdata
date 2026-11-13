package builtinfunc

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/larvit/fejkdata/internal/drawstate"
)

// bankgirotMod11Valid is Bankgirot's check of a number under the 11-modul: weights 1, 2, 3, up to
// 10 from the right, then 1 again, and a sum that 11 divides.
func bankgirotMod11Valid(digits string) bool {
	sum := 0
	for i := range digits {
		sum += int(digits[i]-'0') * ((len(digits)-1-i)%10 + 1)
	}
	return sum%11 == 0
}

// Bankgirot's example of a correct 13-digit number under the 11-modul is 1912763608957.
func TestBankgirotMod11ValidAcceptsBankgirotsExample(t *testing.T) {
	if !bankgirotMod11Valid("1912763608957") {
		t.Fatal("1912763608957 is Bankgirot's example of a correct number")
	}
	if bankgirotMod11Valid("1912763608958") {
		t.Fatal("1912763608958 changes the check digit of Bankgirot's example")
	}
}

func TestBgmod11TailsCompleteBankgirotsExample(t *testing.T) {
	tails := bgmod11Tails("19127636089")
	if !slices.Contains(tails, "57") {
		t.Errorf("bgmod11Tails(19127636089) = %v, want 57 among them, as in Bankgirot's example", tails)
	}
}

// The tails of a prefix depend on it only through its weighted sum modulo 11, and the first 1000
// numbers reach every residue.
func TestBgmod11TailsAreEveryValidPair(t *testing.T) {
	for n := range 1000 {
		prefix := fmt.Sprintf("%03d", n)
		t.Run(prefix, func(t *testing.T) {
			tails := bgmod11Tails(prefix)
			if len(tails) != 9 && len(tails) != 10 {
				t.Fatalf("bgmod11Tails(%s) = %v, want 9 or 10 tails: each free digit but the one whose check digit would be 10", prefix, tails)
			}
			if want := pairsBankgirotMod11Accepts(prefix); !slices.Equal(tails, want) {
				t.Fatalf("bgmod11Tails(%s) = %v, want %v", prefix, tails, want)
			}
		})
	}
}

func pairsBankgirotMod11Accepts(prefix string) []string {
	var pairs []string
	for d := range 10 {
		for c := range 10 {
			if tail := fmt.Sprintf("%d%d", d, c); bankgirotMod11Valid(prefix + tail) {
				pairs = append(pairs, tail)
			}
		}
	}
	return pairs
}

func TestBgmod11RefusesWhatIsNotDigits(t *testing.T) {
	for _, arg := range []string{"1-2", "a", "1 2", "-1"} {
		t.Run(arg, func(t *testing.T) {
			if err := Check("bgmod11", []string{arg}); err == nil {
				t.Errorf("Check(bgmod11, %q) = nil, want an argument of digits only", arg)
			}
		})
	}
}

func TestBgmod11AcceptsDigitsWithALeadingZero(t *testing.T) {
	if err := Check("bgmod11", []string{"0100"}); err != nil {
		t.Errorf("Check(bgmod11, 0100) = %v, want digits with a leading zero accepted", err)
	}
}

// The free digit is the rng's pick among the tails, and the digits emitted before the call count
// by their digits alone.
func TestBgmod11DrawsEveryTailAndSkipsWhatIsNotADigit(t *testing.T) {
	seen := bgmod11TailsDrawnAfter5And1234(t, 400)
	if want := len(bgmod11Tails("51234")); len(seen) != want {
		t.Errorf("%d tails drawn in 400, want each of the %d valid ones", len(seen), want)
	}
}

func bgmod11TailsDrawnAfter5And1234(t *testing.T, draws int) map[string]bool {
	t.Helper()
	call := Prep("bgmod11", []string{"5"})
	s := drawstate.New(1, time.Time{})
	seen := map[string]bool{}
	for range draws {
		tail := call(s, "1 2-3 4", nil)
		if !bankgirotMod11Valid("51234" + tail) {
			t.Fatalf("5 and 1 2-3 4 completed by %q, which Bankgirot's 11-modul refuses", tail)
		}
		seen[tail] = true
	}
	return seen
}
