package builtinfunc

import (
	"fmt"
	"testing"
)

// cprTail completes the first seven digits, a date and a seventh digit, so the ten digits,
// weighted 4 3 2 7 6 5 4 3 2 1, add up to a multiple of 11. The sum depends on the seven only
// through its residue modulo 11, so a prefix for each of the 11 residues, with every parity,
// eighth digit and start, covers every prefix. cprTail panics where it finds no completion.
func TestCprTailCompletesAPrefixOfEveryResidue(t *testing.T) {
	weights := []int{4, 3, 2, 7, 6, 5, 4, 3, 2, 1}
	seen := map[int]bool{}
	for n := 0; len(seen) < 11; n++ {
		if n == 10_000_000 {
			t.Fatalf("seven digits reach %d of the 11 residues", len(seen))
		}
		prefix := fmt.Sprintf("%07d", n)
		sum := 0
		for i := range 7 {
			sum += weights[i] * int(prefix[i]-'0')
		}
		if seen[sum%11] {
			continue
		}
		seen[sum%11] = true
		for last := byte('0'); last <= '9'; last++ {
			for eighth := range 10 {
				for start := range 10 {
					tail := cprTail(prefix, last, eighth, start)
					total := sum
					for i, w := range weights[7:] {
						total += w * int(tail[i]-'0')
					}
					if len(tail) != 3 || tail[0] != byte('0'+eighth) || (tail[2]-'0')%2 != (last-'0')%2 || total%11 != 0 {
						t.Fatalf("cprTail(%s, %c, %d, %d) = %q, want the eighth digit, a ninth, and a check digit of the parity of %c that makes the sum a multiple of 11", prefix, last, eighth, start, tail, last)
					}
				}
			}
		}
	}
}

func TestCprSeventhDigitsAreTheTablesOfCprKontoret(t *testing.T) {
	for year, want := range map[int]string{
		1857: "", 1858: "5678", 1899: "5678", 1900: "0123", 1936: "0123", 1937: "012349", 1999: "012349",
		2000: "456789", 2036: "456789", 2037: "5678", 2057: "5678", 2058: "",
	} {
		if got := cprSeventhDigits(year); got != want {
			t.Errorf("cprSeventhDigits(%d) = %q, want %q", year, got, want)
		}
	}
}
