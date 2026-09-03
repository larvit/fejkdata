package fejkdata

import (
	"strings"
	"testing"
)

func TestOneItemChoiceIsRejected(t *testing.T) {
	for _, src := range []string{`["x"]`, `[{"format":"{d}","d":"x"}]`, `{"format":"{w}","w":["only"]}`} {
		if _, err := compile(parse(t, src)); err == nil || !strings.Contains(err.Error(), "one-item choice") {
			t.Errorf("compile(%s) = %v, want the one-item choice rejected", src, err)
		}
	}
}

func TestRepeatedChoiceItemIsRejected(t *testing.T) {
	for src, want := range map[string]string{
		`["a", "a", "b"]`: `{ "format": "a", "weight": 2 }`,
		`[{"format":"{x}","x":"1"},{"format":"{x}","x":"1"}]`: "repeats item",
		`{"format":"{w}","w":["", "", "x"]}`:                  `{ "format": "", "weight": 2 }`,
	} {
		if _, err := compile(parse(t, src)); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("compile(%s) = %v, want an error naming %s", src, err, want)
		}
	}
	if _, err := compile(parse(t, `[{"format":"a","weight":2}, "b"]`)); err != nil {
		t.Errorf("compile(weighted a, b) = %v", err)
	}
}

func TestInertObjectIsRejected(t *testing.T) {
	for src, want := range map[string]string{
		`{"format":"Malmö"}`:                                 `write "Malmö"`,
		`{"format":"{digits(3)}"}`:                           `write "{digits(3)}"`,
		`[{"format":"a","weight":1},"b"]`:                    "weight 1",
		`{"format":"{x}","x":"v","repeat":1}`:                "repeat 1",
		`{"format":"{x}","x":"v","separator":","}`:           "separator",
		`{"format":"{x}","x":"v","repeat":2,"separator":""}`: "default",
	} {
		if _, err := compile(parse(t, src)); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("compile(%s) = %v, want an error mentioning %s", src, err, want)
		}
	}
	for _, ok := range []string{`[{"format":"a","weight":2},"b"]`, `{"format":"ab","repeat":2}`, `{"format":"{x}","x":"v"}`, `"{digits(3)}"`} {
		if _, err := compile(parse(t, ok)); err != nil {
			t.Errorf("compile(%s) = %v", ok, err)
		}
	}
}

func TestInlineFolderSigilsAreRejected(t *testing.T) {
	f := shipped(t)
	for _, input := range []string{"{.sv_SE.person.last}", "{..sv_SE.person.last}"} {
		_, err := f.NewTemplate(input)
		if err == nil || !strings.Contains(err.Error(), "write {/sv_SE.person.last}") {
			t.Errorf("NewTemplate(%q) = %v, want an error naming the root spelling", input, err)
		}
	}
}
