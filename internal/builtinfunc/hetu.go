package builtinfunc

import (
	"errors"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
)

// hetuControlCharacters is DVV's table of control characters, read by the remainder of the nine
// digits of the date and the individual number divided by 31.
const hetuControlCharacters = "0123456789ABCDEFHJKLMNPRSTUVWXY"

// hetuIndividuals is how many individual numbers of one parity are in 900 to 999, which DVV issues
// to no one: it issues 002 to 899.
const hetuIndividuals = 50

func hetuArg(a []string) error {
	if _, _, isCall := grammar.FuncCall(a[0]); isCall {
		return errors.New("hetu takes the field or path that holds a digit of the sex's parity")
	}
	return nil
}

func hetuPrep([]string) Call {
	return func(s *drawstate.State, emitted string, operands []string) string {
		if len(operands[0]) != 1 || operands[0][0] < '0' || operands[0][0] > '9' {
			return ""
		}
		date, yy := lastSixDigits(emitted)
		sign := hetuSign(s.Today().Year(), yy)
		if sign == 0 {
			return ""
		}
		return hetuTail(date, sign, operands[0][0], s.IntN(hetuIndividuals))
	}
}

// hetuSign is the century sign DVV gives the latest year not after thisYear that bears the year
// digits yy: + for the 1800s, - for the 1900s and A for the 2000s. It is 0 for any other century.
func hetuSign(thisYear, yy int) byte {
	switch year := latestYear(thisYear, yy); {
	case year >= 1800 && year <= 1899:
		return '+'
	case year >= 1900 && year <= 1999:
		return '-'
	case year >= 2000 && year <= 2099:
		return 'A'
	}
	return 0
}

// hetuTail completes a henkilötunnus after its date: the sign, the pick-th individual number from
// 900 whose parity is that of the digit parity, odd for a man and even for a woman, and the
// control character.
func hetuTail(date string, sign, parity byte, pick int) string {
	individual := 900 + 2*pick
	if (parity-'0')%2 == 1 {
		individual++
	}
	digits := date + string([]byte{
		byte('0' + individual/100), byte('0' + individual/10%10), byte('0' + individual%10),
	})
	return string(sign) + digits[6:] + string(hetuCheck(digits))
}

// hetuCheck is the control character of the nine digits of a date and an individual number.
func hetuCheck(digits string) byte {
	n := 0
	for i := range digits {
		n = n*10 + int(digits[i]-'0')
	}
	return hetuControlCharacters[n%31]
}
