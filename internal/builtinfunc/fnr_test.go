package builtinfunc

import (
	"fmt"
	"testing"
)

// The checks of a completion depend on the seven digits only through the weighted sums for k1
// and k2 modulo 11, so a prefix for each of the 121 pairs of residues, with every third
// individ digit and every start, covers every prefix. fnrTail panics where it finds no
// completion, and the root tests hold the oracle that the completions it returns are valid.
func TestFnrCompletesAPrefixOfEveryPairOfResidues(t *testing.T) {
	for pair, prefix := range fnrPrefixOfEachPairOfResidues(t) {
		t.Run(fmt.Sprintf("residues %d and %d", pair[0], pair[1]), func(t *testing.T) {
			fnrTailsComplete(t, prefix)
		})
	}
}

func fnrPrefixResidue(prefix string, weights []int) int {
	sum := 0
	for i, w := range weights[:7] {
		sum += w * int(prefix[i]-'0')
	}
	return sum % 11
}

func fnrPrefixOfEachPairOfResidues(t *testing.T) map[[2]int]string {
	t.Helper()
	prefixes := map[[2]int]string{}
	for n := 0; len(prefixes) < 121; n++ {
		if n == 10_000_000 {
			t.Fatalf("seven digits reach %d of the 121 pairs of residues", len(prefixes))
		}
		prefix := fmt.Sprintf("%07d", n)
		pair := [2]int{fnrPrefixResidue(prefix, fnrWeights[0]), fnrPrefixResidue(prefix, fnrWeights[1])}
		if _, seen := prefixes[pair]; !seen {
			prefixes[pair] = prefix
		}
	}
	return prefixes
}

func fnrTailsComplete(t *testing.T, prefix string) {
	t.Helper()
	for last := byte('0'); last <= '9'; last++ {
		for start := range 10 {
			if tail := fnrTail(prefix, last, start); len(tail) != 4 || tail[1] != last {
				t.Fatalf("fnrTail(%s, %c, %d) = %q, want a second individ digit, %c and two check digits", prefix, last, start, tail, last)
			}
		}
	}
}
