package builtinfunc

import (
	"errors"
	"strings"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/grammar"
	"github.com/larvit/fejkdata/internal/invariant"
)

// fnrWeights are the weights Skatteetaten defines for the check digit k1 over the nine digits
// before it, and for k2 over those and k1.
var fnrWeights = [2][]int{{3, 7, 6, 1, 8, 9, 4, 5, 2}, {5, 4, 3, 2, 7, 6, 5, 4, 3, 2}}

// fnrArg refuses what is not a field or a path: the operand's text, read at render, holds the
// digits the last individ digit may take.
func fnrArg(a []string) error {
	if _, _, isCall := grammar.FuncCall(a[0]); isCall || strings.TrimSpace(a[0]) == "" {
		return errors.New("fnr takes the field or path that holds the digits the last individ digit may be")
	}
	return nil
}

func fnrPrep([]string) Call {
	return func(s *drawstate.State, emitted string, operands []string) string {
		return fnrTail(emitted, operands[0], s.IntN(10*len(fnrLast(operands[0]))))
	}
}

// fnrTail completes a fødselsnummer after the last seven digits of emitted, its date and the
// first individ digit: a second individ digit, a third from last, and the check digits. It tries
// each pair from the one at start on, since a pair whose check digit would be 10 is no number.
func fnrTail(emitted, last string, start int) string {
	last = fnrLast(last)
	n := 10 * len(last)
	prefix := digitsOf(emitted)
	prefix = strings.Repeat("0", max(0, 7-len(prefix))) + prefix[max(0, len(prefix)-7):]
	for i := range n {
		pair := (start + i) % n
		digits := prefix + string(rune('0'+pair/len(last))) + string(last[pair%len(last)])
		k1, ok1 := fnrCheck(digits, fnrWeights[0])
		k2, ok2 := fnrCheck(digits+string(rune('0'+k1)), fnrWeights[1])
		if ok1 && ok2 {
			return digits[7:] + string(rune('0'+k1)) + string(rune('0'+k2))
		}
	}
	panic(invariant.Broken("fnr found no valid completion of %q with %q", prefix, last))
}

// fnrLast is the digits s holds, or every digit where it holds none.
func fnrLast(s string) string {
	if last := digitsOf(s); last != "" {
		return last
	}
	return "0123456789"
}

// fnrCheck is 11 less the weighted sum of digits modulo 11, 0 for 11; ok is false for 10, which
// no number may carry.
func fnrCheck(digits string, weights []int) (k int, ok bool) {
	sum := 0
	for i, w := range weights {
		sum += w * int(digits[i]-'0')
	}
	k = 11 - sum%11
	return k % 11, k != 10
}

// digitsOf keeps the ASCII digits of s.
func digitsOf(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
