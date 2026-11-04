package builtinfunc

import (
	"fmt"
	"slices"
	"testing"
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
		tails := bgmod11Tails(prefix)
		if len(tails) != 9 && len(tails) != 10 {
			t.Fatalf("bgmod11Tails(%s) = %v, want 9 or 10 tails: each free digit but the one whose check digit would be 10", prefix, tails)
		}
		var want []string
		for d := range 10 {
			for c := range 10 {
				if tail := fmt.Sprintf("%d%d", d, c); bankgirotMod11Valid(prefix + tail) {
					want = append(want, tail)
				}
			}
		}
		if !slices.Equal(tails, want) {
			t.Fatalf("bgmod11Tails(%s) = %v, want %v", prefix, tails, want)
		}
	}
}
