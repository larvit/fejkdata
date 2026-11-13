package grammar

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

var segmentsOfEachPath = map[string][]string{
	"a":                       {"a"},
	"a.b":                     {"a", "b"},
	"misc.territory[SE].name": {"misc", "territory", "[SE]", "name"},
	"x[a.b]":                  {"x", "[a.b]"},
	"city[Oslo]..country":     {"city", "[Oslo]", "..", "country"},
	"a..b":                    {"a", "..", "b"},
}

func TestSplitPathKeepsSelectorsAndStepsUpWhole(t *testing.T) {
	for path, want := range segmentsOfEachPath {
		t.Run(path, func(t *testing.T) {
			segs, err := SplitPath(path)
			if err != nil || !reflect.DeepEqual(segs, want) {
				t.Errorf("SplitPath(%q) = %q, %v, want %q", path, segs, err, want)
			}
		})
	}
}

func TestJoinSegmentsSpellsThePathItsSegmentsWereSplitFrom(t *testing.T) {
	for path, segs := range segmentsOfEachPath {
		t.Run(path, func(t *testing.T) {
			if got := JoinSegments(segs); got != path {
				t.Errorf("JoinSegments(%q) = %q, want %q", segs, got, path)
			}
		})
	}
}

func TestSplitPathRefusesAStraySelectorBracket(t *testing.T) {
	for _, path := range []string{"[a]", "a.[b]", "a[b", "a]", "a[]", "a[b]c", "a[b[c]]"} {
		t.Run(path, func(t *testing.T) {
			if _, err := SplitPath(path); err == nil {
				t.Errorf("SplitPath(%q) = nil error, want one", path)
			}
		})
	}
}

func TestIdentifierSegmentsDropSelectorsAndStepsUp(t *testing.T) {
	if got := IdentifierSegments([]string{"a", "[b]", "..", "c"}); !reflect.DeepEqual(got, []string{"a", "c"}) {
		t.Errorf("IdentifierSegments = %q, want [a c]", got)
	}
}

func TestCheckSegmentsRefusesAnEmptySegment(t *testing.T) {
	for _, segs := range [][]string{{""}, {"a", ""}, {"", "a"}, {"a", "", "b"}} {
		t.Run(fmt.Sprintf("%q", segs), func(t *testing.T) {
			if CheckSegments(segs) == nil {
				t.Errorf("CheckSegments(%q) = nil, want an error", segs)
			}
		})
	}
}

func TestCheckSegmentsAcceptsSelectorsAndStepsUp(t *testing.T) {
	if err := CheckSegments([]string{"a", "[b]", "..", "c"}); err != nil {
		t.Errorf("CheckSegments = %v, want nil", err)
	}
}

func TestUnspellableInSelector(t *testing.T) {
	for s, want := range map[string]string{"SE": "", "a|b": "|", `x"y[`: `"`, "": ""} {
		t.Run(strconv.Quote(s), func(t *testing.T) {
			if got := UnspellableInSelector(s); got != want {
				t.Errorf("UnspellableInSelector(%q) = %q, want %q", s, got, want)
			}
		})
	}
}
