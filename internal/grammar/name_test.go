package grammar

import "testing"

func TestCheckNameRefusesWhatAGrammarReserves(t *testing.T) {
	for _, name := range []string{"", "-", "a as b", "a.b", "a|b", "a(b", "a{b", "a}b", "a/b", "a[b", "a]b", `a"b`} {
		if CheckName(name) == nil {
			t.Errorf("CheckName(%q) = nil, want an error", name)
		}
	}
	for _, name := range []string{"a", "postal-code", "é", "a_b"} {
		if err := CheckName(name); err != nil {
			t.Errorf("CheckName(%q) = %v, want nil", name, err)
		}
	}
}
