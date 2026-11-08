package grammar

import (
	"reflect"
	"testing"
)

func TestParseFormatReadsAWhereClause(t *testing.T) {
	five := &LengthRange{Lo: Length{N: 5}, Hi: Length{N: 5}}
	for format, want := range map[string]Token{
		"{/w where length 5}": {Kind: PathRead, Body: "/w where length 5", Arms: []string{"/w"}, Where: &Where{Length: five}},
		"{w where length 4 to 6}": {Kind: PathRead, Body: "w where length 4 to 6", Arms: []string{"w"},
			Where: &Where{Length: &LengthRange{Lo: Length{N: 4}, Hi: Length{N: 6}}}},
		"{w where length k}": {Kind: PathRead, Body: "w where length k", Arms: []string{"w"},
			Where: &Where{Length: &LengthRange{Lo: Length{Name: "k"}, Hi: Length{Name: "k"}}}},
		"{w where length 3 to k}": {Kind: PathRead, Body: "w where length 3 to k", Arms: []string{"w"},
			Where: &Where{Length: &LengthRange{Lo: Length{N: 3}, Hi: Length{Name: "k"}}}},
		"{w where not 'cat', 'dog'}": {Kind: PathRead, Body: "w where not 'cat', 'dog'", Arms: []string{"w"},
			Where: &Where{Not: []string{"cat", "dog"}}},
		"{w where not containing 'cat'}": {Kind: PathRead, Body: "w where not containing 'cat'", Arms: []string{"w"},
			Where: &Where{NotContaining: []string{"cat"}}},
		"{w where not 'a as b'}": {Kind: PathRead, Body: "w where not 'a as b'", Arms: []string{"w"},
			Where: &Where{Not: []string{"a as b"}}},
		"{w where not 'rock and roll','x, y'}": {Kind: PathRead, Body: "w where not 'rock and roll','x, y'", Arms: []string{"w"},
			Where: &Where{Not: []string{"rock and roll", "x, y"}}},
		"{/w[a where b].x where length 5}": {Kind: PathRead, Body: "/w[a where b].x where length 5", Arms: []string{"/w[a where b].x"}, Where: &Where{Length: five}},
		"{w where length 5 and not 'horse' and not containing 'ou' as v}": {Kind: NameBind, Body: "w where length 5 and not 'horse' and not containing 'ou' as v",
			BoundRef: "w", Bound: "v", Where: &Where{Length: five, Not: []string{"horse"}, NotContaining: []string{"ou"}}},
	} {
		toks, err := ParseFormat(format)
		if err != nil {
			t.Errorf("ParseFormat(%q): %v", format, err)
			continue
		}
		if len(toks) != 1 || !reflect.DeepEqual(toks[0], want) {
			t.Errorf("ParseFormat(%q) = %+v, want %+v", format, toks, want)
		}
	}
}

func TestParseFormatRefusesAMalformedWhereClause(t *testing.T) {
	for _, format := range []string{
		"{w where}",
		"{w where length}",
		"{w where length 6 to 4}",
		"{w where length 5 to}",
		"{w where length -1}",
		"{w where length 5 and length 3}",
		"{w where length 5 and}",
		"{w where and length 5}",
		"{w where not cat}",
		"{w where not 'cat}",
		"{w where not 'cat',}",
		"{w where not 'cat' 'dog'}",
		"{w where not containing}",
		"{w where shape 5}",
		"{w where length 5 as}",
		"{w as v where length 5}",
		"{a|b where length 5}",
		"{lower(5) where length 5}",
		"{w where length 5 where length 4}",
	} {
		if toks, err := ParseFormat(format); err == nil {
			t.Errorf("ParseFormat(%q) = %+v, want an error", format, toks)
		}
	}
}

func TestCheckIdentifierRefusesWhere(t *testing.T) {
	if CheckIdentifier("a where b") == nil {
		t.Error(`CheckIdentifier("a where b") = nil, want an error`)
	}
	if err := CheckIdentifier("where"); err != nil {
		t.Errorf(`CheckIdentifier("where") = %v, want nil`, err)
	}
}
