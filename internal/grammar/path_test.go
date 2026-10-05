package grammar

import (
	"reflect"
	"testing"
)

func TestSplitPathKeepsSelectorsAndStepsUpWhole(t *testing.T) {
	for path, want := range map[string][]string{
		"a":                       {"a"},
		"a.b":                     {"a", "b"},
		"misc.territory[SE].name": {"misc", "territory", "[SE]", "name"},
		"x[a.b]":                  {"x", "[a.b]"},
		"city[Oslo]..country":     {"city", "[Oslo]", "..", "country"},
		"a..b":                    {"a", "..", "b"},
	} {
		segs, err := SplitPath(path)
		if err != nil || !reflect.DeepEqual(segs, want) {
			t.Errorf("SplitPath(%q) = %q, %v, want %q", path, segs, err, want)
			continue
		}
		if got := JoinSegments(segs); got != path {
			t.Errorf("JoinSegments(%q) = %q, want %q", segs, got, path)
		}
	}
}

func TestSplitPathRefusesAStraySelectorBracket(t *testing.T) {
	for _, path := range []string{"[a]", "a.[b]", "a[b", "a]", "a[]", "a[b]c", "a[b[c]]"} {
		if _, err := SplitPath(path); err == nil {
			t.Errorf("SplitPath(%q) = nil error, want one", path)
		}
	}
}

func TestNameSegmentsDropSelectorsAndStepsUp(t *testing.T) {
	if got := NameSegments([]string{"a", "[b]", "..", "c"}); !reflect.DeepEqual(got, []string{"a", "c"}) {
		t.Errorf("NameSegments = %q, want [a c]", got)
	}
}
