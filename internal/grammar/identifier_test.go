package grammar

import "testing"

func TestCheckIdentifierRefusesWhatAGrammarReserves(t *testing.T) {
	for _, name := range []string{"", "-", "a as b", "a.b", "a|b", "a(b", "a{b", "a}b", "a/b", "a[b", "a]b", `a"b`} {
		if CheckIdentifier(name) == nil {
			t.Errorf("CheckIdentifier(%q) = nil, want an error", name)
		}
	}
	for _, name := range []string{"a", "postal-code", "é", "a_b"} {
		if err := CheckIdentifier(name); err != nil {
			t.Errorf("CheckIdentifier(%q) = %v, want nil", name, err)
		}
	}
}
