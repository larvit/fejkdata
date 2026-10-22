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

func fnrArg(a []string) error {
	if _, _, isCall := grammar.FuncCall(a[0]); isCall {
		return errors.New("fnr takes the field or path that holds the third individ digit")
	}
	return nil
}

func fnrPrep([]string) Call {
	return func(s *drawstate.State, emitted string, operands []string) string {
		if len(operands[0]) != 1 || operands[0][0] < '0' || operands[0][0] > '9' {
			return ""
		}
		return fnrTail(emitted, operands[0][0], s.IntN(10))
	}
}

// fnrTail completes a fødselsnummer after the last seven digits of emitted, its date and the
// first individ digit: a second individ digit, last, and the check digits. It tries each second
// digit from start on, since one whose check digit would be 10 is no number.
func fnrTail(emitted string, last byte, start int) string {
	prefix := digitsOf(emitted)
	prefix = strings.Repeat("0", max(0, 7-len(prefix))) + prefix[max(0, len(prefix)-7):]
	for i := range 10 {
		digits := prefix + string(rune('0'+(start+i)%10)) + string(last)
		k1, ok1 := fnrCheck(digits, fnrWeights[0])
		if !ok1 {
			continue
		}
		if k2, ok2 := fnrCheck(digits+string(rune('0'+k1)), fnrWeights[1]); ok2 {
			return digits[7:] + string(rune('0'+k1)) + string(rune('0'+k2))
		}
	}
	// Unreachable: k1's sum weighs the second digit by 5 and k2's, through k1, by -6 modulo 11,
	// so each check fails for at most one of the ten.
	panic(invariant.Broken("fnr found no valid completion of %q with %c", prefix, last))
}

func fnrCheck(digits string, weights []int) (k int, ok bool) {
	sum := 0
	for i, w := range weights {
		sum += w * int(digits[i]-'0')
	}
	k = 11 - sum%11
	return k % 11, k != 10
}

func digitsOf(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
