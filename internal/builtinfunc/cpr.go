package builtinfunc

import (
	"errors"
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// cprWeights are the weights CPR-kontoret defines for the ten digits, which must add up to a
// multiple of 11.
var cprWeights = [10]int{4, 3, 2, 7, 6, 5, 4, 3, 2, 1}

func cprArg(a []string) error {
	if _, _, isCall := grammar.FuncCall(a[0]); isCall {
		return errors.New("cpr takes the field or path that holds a digit of the sex's parity")
	}
	return nil
}

func cprPrep([]string) Call {
	return func(s *drawstate.State, emitted string, operands []string) string {
		if len(operands[0]) != 1 || operands[0][0] < '0' || operands[0][0] > '9' {
			return ""
		}
		date := digitsOf(emitted)
		date = strings.Repeat("0", max(0, 6-len(date))) + date[max(0, len(date)-6):]
		yy := int(date[4]-'0')*10 + int(date[5]-'0')
		thisYear := s.Today().Year()
		sevenths := cprSeventhDigits(thisYear - ((thisYear-yy)%100+100)%100)
		if sevenths == "" {
			return ""
		}
		seventh := sevenths[s.IntN(len(sevenths))]
		return string(seventh) + cprTail(date+string(seventh), operands[0][0], s.IntN(10), s.IntN(10))
	}
}

// cprSeventhDigits are the seventh digits CPR-kontoret's table of "Personnummerets opbygning"
// gives a year, which together with the year digits name the century.
func cprSeventhDigits(year int) string {
	switch {
	case year >= 1858 && year <= 1899, year >= 2037 && year <= 2057:
		return "5678"
	case year >= 1900 && year <= 1936:
		return "0123"
	case year >= 1937 && year <= 1999:
		return "012349"
	case year >= 2000 && year <= 2036:
		return "456789"
	}
	return ""
}

// cprTail completes a CPR number after its first seven digits, a date and the seventh digit: the
// given second digit, a third that it tries from start on, and a check digit that has the parity
// of last. A third digit whose check digit would be 10 or of the other parity is no number.
func cprTail(prefix string, last byte, second, start int) string {
	for i := range 10 {
		third := (start + i) % 10
		digits := prefix + string(rune('0'+second)) + string(rune('0'+third))
		sum := 0
		for j, d := range digits {
			sum += cprWeights[j] * int(d-'0')
		}
		check := (11 - sum%11) % 11
		if check != 10 && check%2 == int(last-'0')%2 {
			return string(rune('0'+second)) + string(rune('0'+third)) + string(rune('0'+check))
		}
	}
	// Unreachable: the ten thirds give ten different check digits of the eleven residues, and
	// each parity holds five of the digits 0 to 9, so at most one of them is lost.
	panic(invariant.Broken("cpr found no valid completion of %q with parity %c", prefix, last))
}
