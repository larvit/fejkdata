package builtinfunc

import (
	"fmt"
	"time"

	"github.com/larvit/fejkdata/internal/drawstate"
)

// personasKodsWeights are the weights python-stdnum's lv.pvn gives the first ten digits of a
// personas kods: the check digit is 1 plus each digit times its weight, added up, modulo 11, with 10
// written as 0.
var personasKodsWeights = [10]int{10, 5, 8, 4, 2, 1, 6, 3, 7, 9}

// personasKodsUndatedFrom is the first birth date whose code holds no date: from 1 July 2017 the
// register gives each new person a code that starts with personasKodsUndated.
var personasKodsUndatedFrom = time.Date(2017, 7, 1, 0, 0, 0, 0, time.UTC)

const personasKodsUndated = "32"

func personasKodsPrep([]string, parsed) Call {
	return func(s *drawstate.State, _ string, operands []string) string {
		date := operands[0]
		d, err := time.Parse("20060102", date)
		if err != nil || d.Year() < 1800 || d.Year() > 2099 {
			return ""
		}
		digits := personasKodsDigits(s, d)
		return digits + string(personasKodsCheck(digits))
	}
}

// personasKodsDigits are the first ten digits of the code of someone born on d: the date as day,
// month and year of two digits, a digit for the century, 0 for the 1800s up to 2 for the 2000s, and
// a serial number of three digits; or, from 1 July 2017, personasKodsUndated and eight more digits.
func personasKodsDigits(s *drawstate.State, d time.Time) string {
	if !d.Before(personasKodsUndatedFrom) {
		return personasKodsUndated + fmt.Sprintf("%08d", s.IntN(100000000))
	}
	return d.Format("020106") + fmt.Sprintf("%d%03d", (d.Year()-1800)/100, s.IntN(1000))
}

// personasKodsCheck is the check digit of the first ten digits of a code.
func personasKodsCheck(digits string) byte {
	sum := 1
	for i, w := range personasKodsWeights {
		sum += w * int(digits[i]-'0')
	}
	return byte('0' + sum%11%10)
}
