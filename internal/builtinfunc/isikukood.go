package builtinfunc

import (
	"fmt"
	"time"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
)

// isikukoodWeights are EVS 585's two series of weights for the first ten digits of an Estonian
// personal code: its check digit is the sum with the first, modulo 11, or where that is 10, the
// sum with the second, modulo 11, or where that is 10 too, 0.
var isikukoodWeights = [2][10]int{{1, 2, 3, 4, 5, 6, 7, 8, 9, 1}, {3, 4, 5, 6, 7, 8, 9, 1, 2, 3}}

func isikukoodArgs(a []string) error {
	if len(a) != 2 {
		return fmt.Errorf("isikukood takes the field or path that holds the sex's digit, 1 for a man and 2 for a woman, and the name of a date as YYYYMMDD, got %d args", len(a))
	}
	for _, arg := range a {
		if _, _, isCall := grammar.FuncCall(arg); isCall {
			return fmt.Errorf("isikukood takes a field, path or name, not the call %s", arg)
		}
	}
	return nil
}

func isikukoodPrep([]string, parsed) Call {
	return func(s *drawstate.State, _ string, operands []string) string {
		sex, date := operands[0], operands[1]
		if sex != "1" && sex != "2" {
			return ""
		}
		first, ok := isikukoodFirstDigit(date, sex[0])
		if !ok {
			return ""
		}
		digits := fmt.Sprintf("%c%s%03d", first, date[2:], 1+s.IntN(999))
		return digits + string(isikukoodCheck(digits))
	}
}

// isikukoodFirstDigit is the first digit of the code of someone born on date, written YYYYMMDD, that
// holds the sex's digit, 1 for a man and 2 for a woman: it names the century as well, 1 and 2 for
// the 1800s up to 5 and 6 for the 2000s. It is not ok for a date outside those centuries or that
// does not exist.
func isikukoodFirstDigit(date string, sex byte) (first byte, ok bool) {
	year, ok := isikukoodYear(date)
	if !ok {
		return 0, false
	}
	return byte(int(sex) + 2*((year-1800)/100)), true
}

// isikukoodYear is the year of a date written YYYYMMDD, ok where it is a day of 1800 to 2099.
func isikukoodYear(date string) (year int, ok bool) {
	var month, day int
	if len(date) != 8 || digitsOf(date) != date {
		return 0, false
	}
	if _, err := fmt.Sscanf(date, "%4d%2d%2d", &year, &month, &day); err != nil || year < 1800 || year > 2099 {
		return 0, false
	}
	d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return year, d.Month() == time.Month(month) && d.Day() == day
}

// isikukoodCheck is the check digit of the first ten digits of a code.
func isikukoodCheck(digits string) byte {
	check := 0
	for _, weights := range isikukoodWeights {
		check = 0
		for i, w := range weights {
			check += w * int(digits[i]-'0')
		}
		if check %= 11; check != 10 {
			break
		}
	}
	return byte('0' + check%10)
}
