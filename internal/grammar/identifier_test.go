package grammar

import (
	"strconv"
	"testing"
)

func TestCheckIdentifierRefusesWhatAGrammarReserves(t *testing.T) {
	for _, name := range []string{"", "-", "a as b", "a.b", "a|b", "a(b", "a{b", "a}b", "a/b", "a[b", "a]b", `a"b`} {
		t.Run(strconv.Quote(name), func(t *testing.T) {
			if CheckIdentifier(name) == nil {
				t.Errorf("CheckIdentifier(%q) = nil, want an error", name)
			}
		})
	}
}

func TestCheckIdentifierAcceptsWhatNoGrammarReserves(t *testing.T) {
	for _, name := range []string{"a", "postal-code", "é", "a_b"} {
		t.Run(name, func(t *testing.T) {
			if err := CheckIdentifier(name); err != nil {
				t.Errorf("CheckIdentifier(%q) = %v, want nil", name, err)
			}
		})
	}
}
