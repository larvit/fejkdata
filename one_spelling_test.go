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
		`{"format":"","w":[null,null,"a"]}`:                   "a null takes no weight",
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
		`{"format":"Malmö"}`:                                       `write "Malmö"`,
		`{"format":"{digits(3)}"}`:                                 `write "{digits(3)}"`,
		`[{"format":"a","weight":1},"b"]`:                          "weight 1",
		`{"format":"{x}","x":"v","repeat":1}`:                      "repeat 1",
		`{"format":"{x}","x":"v","separator":","}`:                 "separator",
		`{"format":"{x}","x":"v","repeat":2,"separator":""}`:       "default",
		`{"format":"","n":{"format":"1","datatype":"string"}}`:     `datatype "string" is the default`,
		`{"format":"{x}","x":{"format":"{y}","y":"v","group":""}}`: `group "" is the default`,
		`{"format":"{x}","x":{"format":"{y}","y":"v","group":1}}`:  "group must be a string",
	} {
		if _, err := compile(parse(t, src)); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("compile(%s) = %v, want an error mentioning %s", src, err, want)
		}
	}
	for _, ok := range []string{`[{"format":"a","weight":2},"b"]`, `{"format":"ab","repeat":2}`, `{"format":"{x}","x":"v"}`, `"{digits(3)}"`, `{"format":"{/cat.x}","group":"g"}`} {
		if _, err := compile(parse(t, ok)); err != nil {
			t.Errorf("compile(%s) = %v", ok, err)
		}
	}
}

func TestAGroupThatSplitsNothingIsRejected(t *testing.T) {
	files := map[string]string{"mail": `"{/word.w}@example.com"`, "word": `{"format":"{w}","w":["a","b"]}`}
	for src, want := range map[string]string{
		`{"format":"{x}","x":{"format":"{y}","y":["a","b"],"group":"g"}}`:              `group "g" splits nothing`,
		`{"format":"{x}","x":{"format":"{/word}","group":"g"}}`:                        `group "g" splits nothing`,
		`{"format":"{/word.w}","group":"g"}`:                                           "",
		`{"format":"{x}","x":{"format":"{/mail}","group":"g"}}`:                        "",
		`{"format":"{x}","x":{"format":"{y}","y":"{/word.w}","repeat":2,"group":"g"}}`: "",
	} {
		files["cat"] = src
		_, err := New(WithoutShippedData(), WithDataPath(writeData(t, files)))
		switch {
		case want == "" && err != nil:
			t.Errorf("New(%s) = %v, want a group over a reference path accepted, however deep", src, err)
		case want != "" && (err == nil || !strings.Contains(err.Error(), want)):
			t.Errorf("New(%s) = %v, want an error mentioning %s", src, err, want)
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
