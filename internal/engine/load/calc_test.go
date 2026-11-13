package load

import (
	"testing"

	"github.com/larvit/fejkdata/internal/grammar"
)

// TestParsedCallReadsOnlyAnOperandBuiltin pins the operands a parsed call carries for a
// call that is not a calc, and for one whose expression does not parse: none, either way.
func TestParsedCallReadsOnlyAnOperandBuiltin(t *testing.T) {
	for _, format := range []string{"{luhn()}", "{calc()}", "{calc(1 +)}", "{calc(()}"} {
		t.Run(format, func(t *testing.T) {
			if toks, err := grammar.ParseFormat(format); err != nil || len(toks) != 1 || ownReadsOf(toks[0]) != nil {
				t.Errorf("ParseFormat(%q) = %+v, %v, want one call reading no operand", format, toks, err)
			}
		})
	}
}

func TestParsedCallReadsBothOperandsOfACalc(t *testing.T) {
	if toks, err := grammar.ParseFormat("{calc(net * qty)}"); err != nil || len(toks) != 1 || len(ownReadsOf(toks[0])) != 2 {
		t.Errorf("ParseFormat({calc(net * qty)}) = %+v, %v, want both operands", toks, err)
	}
}

func ownReadsOf(tok grammar.Token) []string {
	own, _ := tokenReads(tok)
	return own
}
