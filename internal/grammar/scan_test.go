package grammar

import (
	"reflect"
	"testing"
)

func TestCutWordIgnoresAnApostropheOutsideAClauseOrInsideASelector(t *testing.T) {
	for _, c := range []struct{ body, head, rest string }{
		{"o'neil where not 'a where b'", "o'neil", "not 'a where b'"},
		{"t[a'b] where length 3", "t[a'b]", "length 3"},
		{"t[x where y] where length 3", "t[x where y]", "length 3"},
	} {
		head, rest, found := cutWord(c.body, whereWord, false)
		if !found || head != c.head || rest != c.rest {
			t.Errorf("cutWord(%q) = %q, %q, %v, want %q, %q", c.body, head, rest, found, c.head, c.rest)
		}
	}
}

func TestCutWordHonoursAQuoteOnceAClauseOpens(t *testing.T) {
	head, rest, found := cutWord("w where not 'a distinct b' distinct 2", distinctWord, false)
	if !found || head != "w where not 'a distinct b'" || rest != "2" {
		t.Errorf("cutWord(distinct) = %q, %q, %v, want %q, %q", head, rest, found, "w where not 'a distinct b'", "2")
	}
}

func TestLastAsOutsideSkipsAQuotedAs(t *testing.T) {
	before, name, found := lastAsOutside("not 'x as y' as n")
	if !found || before != "not 'x as y'" || name != "n" {
		t.Errorf("lastAsOutside = %q, %q, %v, want %q, %q", before, name, found, "not 'x as y'", "n")
	}
}

func TestCutOutsideAndIndexOutsideSkipASelectorAndIgnoreAnApostrophe(t *testing.T) {
	if before, after, found := cutOutside("a[x as y] as n", AsWord); !found || before != "a[x as y]" || after != "n" {
		t.Errorf("cutOutside(%q, as) = %q, %q, %v, want %q, %q", "a[x as y] as n", before, after, found, "a[x as y]", "n")
	}
	if i := indexOutside("o'neil|b", '|'); i != 6 {
		t.Errorf("indexOutside(%q, '|') = %d, want 6", "o'neil|b", i)
	}
	if i := indexOutside("a[x|y]|b", '|'); i != 6 {
		t.Errorf("indexOutside(%q, '|') = %d, want 6", "a[x|y]|b", i)
	}
}

func TestSplitArgsSkipsACommaInAQuoteOrASelector(t *testing.T) {
	for in, want := range map[string][]string{
		"'a,b', [c,d], e": {"'a,b'", "[c,d]", "e"},
		"x[a'b], c":       {"x[a'b]", "c"},
		"'[a', b":         {"'[a'", "b"},
	} {
		if got := splitArgs(in); !reflect.DeepEqual(got, want) {
			t.Errorf("splitArgs(%q) = %q, want %q", in, got, want)
		}
	}
}
