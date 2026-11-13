package grammar

import (
	"reflect"
	"testing"
)

func TestCutWordSkipsAQuoteOutsideAClauseAndInsideASelector(t *testing.T) {
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
		t.Errorf("got %q, %q, %v", head, rest, found)
	}
}

func TestLastAsOutsideQuotesSkipsAQuotedAs(t *testing.T) {
	before, name, found := lastAsOutsideQuotes("not 'x as y' as n")
	if !found || before != "not 'x as y'" || name != "n" {
		t.Errorf("got %q, %q, %v", before, name, found)
	}
}

func TestCutOutsideAndIndexOutsideSkipASelectorButNotAnApostrophe(t *testing.T) {
	if before, after, found := cutOutside("a[x as y] as n", AsWord); !found || before != "a[x as y]" || after != "n" {
		t.Errorf("cutOutside = %q, %q, %v", before, after, found)
	}
	if i := indexOutside("o'neil|b", '|'); i != 6 {
		t.Errorf("indexOutside = %d, want 6", i)
	}
	if i := indexOutside("a[x|y]|b", '|'); i != 6 {
		t.Errorf("indexOutside = %d, want 6", i)
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
