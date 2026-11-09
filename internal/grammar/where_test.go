package grammar

import (
	"reflect"
	"strconv"
	"strings"
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
		"{o'neil where length 3}": {Kind: PathRead, Body: "o'neil where length 3", Arms: []string{"o'neil"},
			Where: &Where{Length: &LengthRange{Lo: Length{N: 3}, Hi: Length{N: 3}}}},
		"{w where length m[1]}": {Kind: PathRead, Body: "w where length m[1]", Arms: []string{"w"},
			Where: &Where{Length: &LengthRange{Lo: Length{Name: "m[1]"}, Hi: Length{Name: "m[1]"}}}},
		"{w where length 5 and not 'horse' and not containing 'ou' as v}": {Kind: NameBind, Body: "w where length 5 and not 'horse' and not containing 'ou' as v",
			BoundRef: "w", Bound: "v", Where: &Where{Length: five, Not: []string{"horse"}, NotContaining: []string{"ou"}}},
		"{w where not cat}": {Kind: PathRead, Body: "w where not cat", Arms: []string{"w"},
			Where: &Where{NotNames: []string{"cat"}}},
		"{w where not 'cat', m[1] and not containing t}": {Kind: PathRead, Body: "w where not 'cat', m[1] and not containing t", Arms: []string{"w"},
			Where: &Where{Not: []string{"cat"}, NotNames: []string{"m[1]"}, NotContainingNames: []string{"t"}}},
		"{w where not containing 'ou' else 'none'}": {Kind: PathRead, Body: "w where not containing 'ou' else 'none'", Arms: []string{"w"},
			Where: &Where{NotContaining: []string{"ou"}, Else: &Fallback{Text: "none"}}},
		"{w where not t else ''}": {Kind: PathRead, Body: "w where not t else ''", Arms: []string{"w"},
			Where: &Where{NotNames: []string{"t"}, Else: &Fallback{}}},
		"{w where not t else f distinct 2 as v}": {Kind: NameBind, Body: "w where not t else f distinct 2 as v", BoundRef: "w", Bound: "v",
			Where: &Where{NotNames: []string{"t"}, Else: &Fallback{Name: "f"}}, Distinct: &Distinct{N: 2}},
		"{w where not 'else'}": {Kind: PathRead, Body: "w where not 'else'", Arms: []string{"w"},
			Where: &Where{Not: []string{"else"}}},
		"{lower(5) where length 5}": {Kind: BuiltinCall, Body: "lower(5) where length 5", Fn: "lower", Args: []string{"5"},
			Where: &Where{Length: five}},
		"{uuid() where not containing 'dead', t else id}": {Kind: BuiltinCall, Body: "uuid() where not containing 'dead', t else id", Fn: "uuid",
			Where: &Where{NotContaining: []string{"dead"}, NotContainingNames: []string{"t"}, Else: &Fallback{Name: "id"}}},
	} {
		t.Run(format, func(t *testing.T) {
			toks, err := ParseFormat(format)
			if err != nil {
				t.Errorf("ParseFormat(%q): %v", format, err)
				return
			}
			if len(toks) != 1 || !reflect.DeepEqual(toks[0], want) {
				t.Errorf("ParseFormat(%q) = %+v, want %+v", format, toks, want)
			}
		})
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
		"{w where not 'cat}",
		"{w where not 'cat',}",
		"{w where not 'cat' 'dog'}",
		"{w where not containing}",
		"{w where shape 5}",
		"{w where length 5 as}",
		"{w as v where length 5}",
		"{a|b where length 5}",
		"{w where length 5 where length 4}",
		"{w where not 5}",
		"{w where not -x}",
		"{w where not a.b}",
		"{w where not containing 'a', 9}",
		"{w where not and}",
		"{w where not 'a' else and}",
		"{w where else 'x'}",
		"{w where not 'a' else}",
		"{w where not 'a' else 5}",
		"{w where not 'a' else a.b}",
		"{w where not 'a' else 'x' and length 3}",
		"{w where not 'a' else 'x' else 'y'}",
		"{w where not 'a' else 'x' 'y'}",
		"{lower(5) where length 5 distinct 2 joined ','}",
		"{lower(5) where length 5 as v}",
		"{lower(5)|w where length 5}",
		"{lower(5 where length 5}",
	} {
		t.Run(format, func(t *testing.T) {
			if toks, err := ParseFormat(format); err == nil {
				t.Errorf("ParseFormat(%q) = %+v, want an error", format, toks)
			}
		})
	}
}

func TestParseFormatReadsADistinctClause(t *testing.T) {
	five := &LengthRange{Lo: Length{N: 5}, Hi: Length{N: 5}}
	for format, want := range map[string]Token{
		"{/w distinct 2 joined ','}": {Kind: PathRead, Body: "/w distinct 2 joined ','", Arms: []string{"/w"},
			Distinct: &Distinct{N: 2, Joins: true, Sep: ","}},
		"{w distinct 3 as e}": {Kind: NameBind, Body: "w distinct 3 as e", BoundRef: "w", Bound: "e", Distinct: &Distinct{N: 3}},
		"{/w where length 5 distinct 3 as v}": {Kind: NameBind, Body: "/w where length 5 distinct 3 as v", BoundRef: "/w", Bound: "v",
			Where: &Where{Length: five}, Distinct: &Distinct{N: 3}},
		"{w distinct 2 joined ' as ' as d}": {Kind: NameBind, Body: "w distinct 2 joined ' as ' as d", BoundRef: "w", Bound: "d",
			Distinct: &Distinct{N: 2, Joins: true, Sep: " as "}},
		"{w where not 'distinct' distinct 2 joined ''}": {Kind: PathRead, Body: "w where not 'distinct' distinct 2 joined ''", Arms: []string{"w"},
			Where: &Where{Not: []string{"distinct"}}, Distinct: &Distinct{N: 2, Joins: true}},
		"{o'neil distinct 2 joined ','}": {Kind: PathRead, Body: "o'neil distinct 2 joined ','", Arms: []string{"o'neil"},
			Distinct: &Distinct{N: 2, Joins: true, Sep: ","}},
		"{w distinct 2 joined ' where '}": {Kind: PathRead, Body: "w distinct 2 joined ' where '", Arms: []string{"w"},
			Distinct: &Distinct{N: 2, Joins: true, Sep: " where "}},
		"{w where not 'a distinct b' distinct 2 joined ''}": {Kind: PathRead, Body: "w where not 'a distinct b' distinct 2 joined ''", Arms: []string{"w"},
			Where: &Where{Not: []string{"a distinct b"}}, Distinct: &Distinct{N: 2, Joins: true}},
		"{w where not ']' distinct 2 joined ''}": {Kind: PathRead, Body: "w where not ']' distinct 2 joined ''", Arms: []string{"w"},
			Where: &Where{Not: []string{"]"}}, Distinct: &Distinct{N: 2, Joins: true}},
		"{w[a distinct b] distinct 2 joined ','}": {Kind: PathRead, Body: "w[a distinct b] distinct 2 joined ','", Arms: []string{"w[a distinct b]"},
			Distinct: &Distinct{N: 2, Joins: true, Sep: ","}},
	} {
		t.Run(format, func(t *testing.T) {
			toks, err := ParseFormat(format)
			if err != nil {
				t.Errorf("ParseFormat(%q): %v", format, err)
				return
			}
			if len(toks) != 1 || !reflect.DeepEqual(toks[0], want) {
				t.Errorf("ParseFormat(%q) = %+v, want %+v", format, toks, want)
			}
		})
	}
}

func TestParseFormatRefusesAMalformedDistinctClause(t *testing.T) {
	for _, format := range []string{
		"{w distinct}",
		"{w distinct 0 joined ','}",
		"{w distinct -1 joined ','}",
		"{w distinct 2.5 as v}",
		"{w distinct x joined ','}",
		"{w distinct 2}",
		"{w where length 5 distinct 2}",
		"{w distinct 2 joined}",
		"{w distinct 2 joined ,}",
		"{w distinct 2 joined ',' ','}",
		"{w distinct 2 and 3 joined ','}",
		"{w distinct 2 as}",
		"{w as v distinct 2}",
		"{a|b distinct 2 joined ','}",
		"{lower(5) distinct 2 joined ','}",
		"{w distinct 2 where length 5 as v}",
		"{w distinct 2 distinct 3 as v}",
		"{w where length 5 distinct}",
		"{w where length distinct 2 as v}",
	} {
		t.Run(format, func(t *testing.T) {
			if toks, err := ParseFormat(format); err == nil {
				t.Errorf("ParseFormat(%q) = %+v, want an error", format, toks)
			}
		})
	}
}

func TestCheckIdentifierRefusesAWordOfTheClausesInsideAName(t *testing.T) {
	for _, word := range []string{"where", "distinct"} {
		t.Run(word, func(t *testing.T) {
			inside := strconv.Quote(" " + word + " ")
			if err := CheckIdentifier("a " + word + " b"); err == nil || !strings.Contains(err.Error(), "contains "+inside) {
				t.Errorf("CheckIdentifier(%q) = %v, want an error naming %s", "a "+word+" b", err, inside)
			}
		})
	}
}

func TestCheckIdentifierRefusesAWordOfTheClausesEndingAName(t *testing.T) {
	for _, word := range []string{"where", "distinct"} {
		t.Run(word, func(t *testing.T) {
			end := strconv.Quote(" " + word)
			if err := CheckIdentifier("a " + word); err == nil || !strings.Contains(err.Error(), "ends in "+end) {
				t.Errorf("CheckIdentifier(%q) = %v, want an error naming %s at its end", "a "+word, err, end)
			}
		})
	}
}

func TestCheckIdentifierAcceptsAWordOfTheClausesAlone(t *testing.T) {
	for _, word := range []string{"where", "distinct"} {
		t.Run(word, func(t *testing.T) {
			if err := CheckIdentifier(word); err != nil {
				t.Errorf("CheckIdentifier(%q) = %v, want nil", word, err)
			}
		})
	}
}
