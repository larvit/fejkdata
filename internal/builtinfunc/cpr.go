package builtinfunc

import (
	"errors"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// cprWeights are CPR-kontoret's weights for the ten digits; a number's weighted sum is a multiple of 11.
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
		date, yy := lastSixDigits(emitted)
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

// cprTail completes a CPR number after its first seven digits: the given eighth digit, a ninth
// that it tries from start on, and a check digit that has the parity of last. A ninth digit whose
// check digit would be 10 or of the other parity is no number.
func cprTail(prefix string, last byte, eighth, start int) string {
	for i := range 10 {
		ninth := (start + i) % 10
		digits := prefix + string(rune('0'+eighth)) + string(rune('0'+ninth))
		sum := 0
		for j, d := range digits {
			sum += cprWeights[j] * int(d-'0')
		}
		check := (11 - sum%11) % 11
		if check != 10 && check%2 == int(last-'0')%2 {
			return string(rune('0'+eighth)) + string(rune('0'+ninth)) + string(rune('0'+check))
		}
	}
	// Unreachable: the ten ninths give ten different residues of the eleven, so at most one of
	// the five digits of the wanted parity is missing as a check digit.
	panic(invariant.Broken("cpr found no valid completion of %q with parity %c", prefix, last))
}
