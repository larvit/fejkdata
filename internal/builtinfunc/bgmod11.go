package builtinfunc

import (
	"fmt"
	"strconv"

	"github.com/larvit/fejkdata/internal/drawstate"
)

func bgmod11Arg(a []string) error {
	if digitsOf(a[0]) != a[0] {
		return fmt.Errorf("bgmod11 takes the digits that come before the emitted ones, as in {bgmod11(0)}, got %q", a[0])
	}
	return nil
}

func bgmod11Prep(a []string) Call {
	weighted := a[0]
	return func(s *drawstate.State, emitted string, _ []string) string {
		tails := bgmod11Tails(weighted + digitsOf(emitted))
		return tails[s.IntN(len(tails))]
	}
}

// bgmod11Tails returns every pair of digits that completes weighted to a number Bankgirot's 11-modul
// accepts: a free digit and the check digit, weighted 2 and 1, since weights run 1, 2, 3 and up to
// 10 from the right and then begin at 1 again. A free digit whose check digit would be 10 has none,
// so there are nine or ten tails.
func bgmod11Tails(weighted string) []string {
	rest := 0
	for i := range len(weighted) {
		rest += int(weighted[i]-'0') * ((len(weighted)-1-i+2)%10 + 1)
	}
	var tails []string
	for free := range 10 {
		if check := (11 - (rest+2*free)%11) % 11; check != 10 {
			tails = append(tails, strconv.Itoa(free)+strconv.Itoa(check))
		}
	}
	return tails
}
