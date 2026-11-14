package builtinfunc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
)

// hetuControlCharacters is DVV's table of control characters, read by the remainder of the nine
// digits of the date and the individual number divided by 31.
const hetuControlCharacters = "0123456789ABCDEFHJKLMNPRSTUVWXY"

// hetuRange is a run of individual numbers: its lowest even number and how many it holds of one
// parity.
type hetuRange struct{ first, count int }

// DVV issues the individual numbers 002 to 899, and 900 to 999 to no one.
var (
	hetuUnissued = hetuRange{first: 900, count: 50}
	hetuIssued   = hetuRange{first: 2, count: 449}
)

func hetuArgs(a []string) error {
	if len(a) != 1 && len(a) != 2 {
		return fmt.Errorf("hetu takes the digit of the sex's parity and an optional 'issued', got %d args", len(a))
	}
	if _, _, isCall := grammar.FuncCall(a[0]); isCall {
		return errors.New("hetu takes the field or path that holds a digit of the sex's parity")
	}
	if len(a) == 1 || a[1] == "'issued'" {
		return nil
	}
	if !quotedLayout(a[1]) {
		return fmt.Errorf("hetu's range %s is not quoted; write '%s'", a[1], strings.Trim(a[1], `'"`))
	}
	return fmt.Errorf("hetu's range %s is not 'issued', the one range it takes beside its default of 900 to 999", a[1])
}

func hetuPrep(a []string) Call {
	individuals := hetuUnissued
	if len(a) == 2 {
		individuals = hetuIssued
	}
	return func(s *drawstate.State, emitted string, operands []string) string {
		if len(operands[0]) != 1 || operands[0][0] < '0' || operands[0][0] > '9' {
			return ""
		}
		date, yy := lastSixDigits(emitted)
		sign := hetuSign(s.Today().Year(), yy)
		if sign == 0 {
			return ""
		}
		return hetuTail(date, sign, operands[0][0], individuals, s.IntN(individuals.count))
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

// hetuTail completes a henkilötunnus after its date: the sign, the pick-th individual number of
// individuals whose parity is that of the digit parity, odd for a man and even for a woman, and
// the control character.
func hetuTail(date string, sign, parity byte, individuals hetuRange, pick int) string {
	individual := individuals.first + 2*pick
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
